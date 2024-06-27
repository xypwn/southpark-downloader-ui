package southpark

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
)

type DownloaderStatus int

const (
	DownloaderStatusFetchingMetadata DownloaderStatus = iota
	DownloaderStatusDownloadingVideo
	DownloaderStatusDownloadingAudio
	DownloaderStatusDownloadingSubtitles
	DownloaderStatusPostprocessingVideo
	DownloaderStatusPostprocessingSubtitles
)

type Downloader struct {
	OnStatusChanged func(
		status DownloaderStatus,
		// Between [0;1] if progress can be estimated, -1 if progress can't be estimated;
		// always related to current status
		progress float64,
	)

	selectFormat       func([]HLSFormat) (HLSFormat, error)
	ctx                context.Context
	tmpDirPath         string
	outputVideoPath    string // Empty to download subs only
	outputSubtitlePath string // Empty to download video only
	episode            Episode
}

func NewDownloader(
	ctx context.Context,
	episode Episode,
	tmpDirPath string,
	outputVideoPath string, // Empty to download subs only
	selectVideoFormat func([]HLSFormat) (HLSFormat, error),
	outputSubtitlePath string, // Empty to download video only
) *Downloader {
	return &Downloader{
		OnStatusChanged:    func(DownloaderStatus, float64) {},
		selectFormat:       selectVideoFormat,
		ctx:                ctx,
		tmpDirPath:         tmpDirPath,
		outputVideoPath:    outputVideoPath,
		outputSubtitlePath: outputSubtitlePath,
		episode:            episode,
	}
}

func (d *Downloader) Do() error {
	d.OnStatusChanged(DownloaderStatusFetchingMetadata, -1)

	stream, err := GetEpisodeStream(d.ctx, d.episode, d.selectFormat)
	if err != nil {
		return fmt.Errorf("GetEpisodeStream: %w", err)
	}

	getSegFileName := func(n int) string {
		var ext string
		if n < len(stream.Video.Segments) {
			ext = "ts"
		} else if n < len(stream.Video.Segments)+len(stream.Audio.Segments) {
			ext = "aac"
		} else if n < len(stream.Video.Segments)+len(stream.Audio.Segments)+len(stream.Subs.Segments) {
			ext = "vtt"
		}
		return path.Join(d.tmpDirPath, fmt.Sprintf("Seg%04v.%v", n, ext))
	}

	if d.outputVideoPath != "" {
		d.OnStatusChanged(DownloaderStatusDownloadingVideo, 0)

		startTotalSegment := 0
		if _, err := os.Stat(d.tmpDirPath); err == nil {
			i := 0
			for {
				if _, err := os.Stat(getSegFileName(i)); err != nil {
					break
				}
				i++
			}
			if i > 0 {
				// Start at the segment before the last existing one, since
				// writing to the last one may have been partial
				startTotalSegment = i - 1
			}
		} else {
			if err := os.MkdirAll(d.tmpDirPath, os.ModePerm); err != nil {
				return fmt.Errorf("create temporary media directory: %w", err)
			}
		}

		currentTotalSegment := startTotalSegment
		if err := DownloadEpisodeStream(d.ctx, stream, startTotalSegment,
			func(segmentIdx int) {
				currentTotalSegment = segmentIdx
			},
			func(frame []byte, relSegIdx int) error {
				if err := os.WriteFile(getSegFileName(currentTotalSegment), frame, 0644); err != nil {
					return err
				}
				d.OnStatusChanged(DownloaderStatusDownloadingVideo, float64(relSegIdx)/float64(len(stream.Video.Segments)))
				return nil
			},
			func(frame []byte, relSegIdx int) error {
				if err := os.WriteFile(getSegFileName(currentTotalSegment), frame, 0644); err != nil {
					return err
				}
				d.OnStatusChanged(DownloaderStatusDownloadingAudio, float64(relSegIdx)/float64(len(stream.Audio.Segments)))
				return nil
			},
			func(frame []byte, relSegIdx int) error {
				if err := os.WriteFile(getSegFileName(currentTotalSegment), frame, 0644); err != nil {
					return err
				}
				d.OnStatusChanged(DownloaderStatusDownloadingSubtitles, float64(relSegIdx)/float64(len(stream.Subs.Segments)))
				return nil
			}); err != nil {
			return fmt.Errorf("GetEpisodeAsTS: %w", err)
		}

		d.OnStatusChanged(DownloaderStatusPostprocessingVideo, 0)

		outputFileMP4, err := os.Create(d.outputVideoPath)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}
		defer outputFileMP4.Close()

		var tsSegs []SegmentFile
		for i, seg := range stream.Video.Segments {
			tsSegs = append(tsSegs, SegmentFile{
				Filename: getSegFileName(i),
				Duration: seg.Duration,
			})
		}

		var aacSegs []SegmentFile
		for i, seg := range stream.Audio.Segments {
			aacSegs = append(aacSegs, SegmentFile{
				Filename: getSegFileName(len(stream.Video.Segments) + i),
				Duration: seg.Duration,
			})
		}

		if err := ConvertTSAndAACToMP4(tsSegs, aacSegs, outputFileMP4, func(progress float64) {
			d.OnStatusChanged(DownloaderStatusPostprocessingVideo, progress)
		}); err != nil {
			return fmt.Errorf("convert MPEG-TS and AAC to MP4: %w", err)
		}
	}

	if d.outputSubtitlePath != "" {
		var out bytes.Buffer

		if _, err := out.WriteString("WEBVTT\r\n\r\n"); err != nil {
			return fmt.Errorf("write subs: %w", err)
		}

		if len(stream.Subs.Segments) > 0 {
			startSeg := len(stream.Video.Segments) + len(stream.Audio.Segments)
			endSeg := len(stream.Video.Segments) + len(stream.Audio.Segments) + len(stream.Subs.Segments)
			for i := startSeg; i < endSeg; i++ {
				data, err := os.ReadFile(getSegFileName(i))
				if err != nil {
					return fmt.Errorf("read subs fragment: %w", err)
				}
				data = bytes.TrimPrefix(data, []byte("WEBVTT\r\n\r\n"))
				if _, err := out.Write(data); err != nil {
					return fmt.Errorf("write subs: %w", err)
				}
				d.OnStatusChanged(DownloaderStatusPostprocessingSubtitles, float64(i)/float64(endSeg))
			}
		}

		if err := os.WriteFile(d.outputSubtitlePath, out.Bytes(), 0666); err != nil {
			return fmt.Errorf("write VTT subtitles: %w", err)
		}
	}

	if err := os.RemoveAll(d.tmpDirPath); err != nil {
		return fmt.Errorf("remove temporary media directory: %w", err)
	}

	return nil
}
