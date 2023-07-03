package southpark

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/xypwn/southpark-downloader-ui/pkg/ioutils"
)

type DownloaderStatus int

const (
	DownloaderStatusFetchingMetadata DownloaderStatus = iota
	DownloaderStatusDownloadingVideo
	DownloaderStatusDownloadingSubtitles
	DownloaderStatusPostprocessing
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

	parts, err := GetEpisodeParts(d.ctx, d.episode, d.selectFormat)
	if err != nil {
		return fmt.Errorf("GetEpisodeParts: %w", err)
	}

	if d.outputVideoPath != "" {
		d.OnStatusChanged(DownloaderStatusDownloadingVideo, 0)

		getSegFileName := func(n int) string {
			return path.Join(d.tmpDirPath, fmt.Sprintf("Seg%04v.ts", n))
		}

		startSegment := 0
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
				startSegment = i - 1
			}
		} else {
			if err := os.MkdirAll(d.tmpDirPath, os.ModePerm); err != nil {
				return fmt.Errorf("create temporary media directory: %w", err)
			}
		}

		totalSegments := GetPartsTotalHLSSegments(parts)
		currentSegment := startSegment
		if err := GetEpisodeTSVideo(d.ctx, parts, startSegment, func(frame []byte) error {
			if err := os.WriteFile(getSegFileName(currentSegment), frame, 0644); err != nil {
				return err
			}
			d.OnStatusChanged(DownloaderStatusDownloadingVideo, float64(currentSegment)/float64(totalSegments))
			currentSegment++
			return nil
		}); err != nil {
			return fmt.Errorf("GetEpisodeAsTS: %w", err)
		}

		d.OnStatusChanged(DownloaderStatusPostprocessing, 0)

		outputFileMP4, err := os.Create(d.outputVideoPath)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}
		defer outputFileMP4.Close()

		baseTsReader, tsWriter := io.Pipe()
		tsReader := ioutils.NewCtxReader(d.ctx, baseTsReader)

		var convertErr error
		go func() {
			for i := 0; i < totalSegments; i++ {
				tsFileName := getSegFileName(i)
				tsData, err := os.ReadFile(tsFileName)
				if err != nil {
					convertErr = err
					break
				}
				tsWriter.Write(tsData)
				d.OnStatusChanged(DownloaderStatusPostprocessing, float64(i)/float64(totalSegments))
			}
			tsWriter.Close()
		}()
		if err := ConvertTSToMP4(tsReader, outputFileMP4); err != nil {
			return fmt.Errorf("convert MPEG-TS to mp4: %w", err)
		}
		if convertErr != nil {
			return fmt.Errorf("convert MPEG-TS to mp4: %w", convertErr)
		}
		if err := os.RemoveAll(d.tmpDirPath); err != nil {
			return fmt.Errorf("remove temporary media directory: %w", err)
		}
	}

	if d.outputSubtitlePath != "" {
		d.OnStatusChanged(DownloaderStatusDownloadingSubtitles, -1)

		subs, err := GetEpisodeVTTSubtitles(d.ctx, parts)
		if err != nil {
			return fmt.Errorf("GetEpisodeVTTSubtitles: %w", err)
		}

		if err := os.WriteFile(d.outputSubtitlePath, subs, 0666); err != nil {
			return fmt.Errorf("write VTT subtitles: %w", err)
		}
	}

	return nil
}
