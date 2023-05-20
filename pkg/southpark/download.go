package southpark

import (
	"os"
	"path"
	"fmt"
	"io"
	"context"
)

func GetDownloadOutputFileName(episode Episode) string {
	return fmt.Sprintf(
		"South_Park_%v_S%02v_E%02v_%v",
		episode.Language.String(),
		episode.SeasonNumber,
		episode.EpisodeNumber,
		toValidFilename(episode.Title),
	)
}

type Downloader struct {
	OnFinishGetMetadata func()
	OnStartPostprocess  func()
	OnProgress  func(_ float64, postprocessing bool) // If not postprocessing, it is downloading
	OnStartDownloadSubtitles func()

	selectFormat   func([]HLSFormat) (HLSFormat, error)
	ctx            context.Context
	tmpDirPath     string
	outputVideoPath string // Empty to download subs only
    outputSubtitlePath string // Empty to download video only
	episode        Episode
}

func NewDownloader(
    ctx context.Context,
    episode Episode,
    tmpDirPath string,
    outputVideoPath string, // Empty to download subs only
    selectFormat func([]HLSFormat) (HLSFormat, error),
    outputSubtitlePath string, // Empty to download video only
) *Downloader {
	return &Downloader{
		OnFinishGetMetadata: func() {},
		OnStartPostprocess:  func() {},
		OnProgress:  func(_ float64, postprocessing bool) {},
		OnStartDownloadSubtitles: func() {},
		selectFormat:        selectFormat,
		ctx:                 ctx,
		tmpDirPath:          tmpDirPath,
		outputVideoPath:      outputVideoPath,
        outputSubtitlePath: outputSubtitlePath,
		episode:             episode,
	}
}

func (d *Downloader) Do() error {
	parts, err := GetEpisodeParts(d.ctx, d.episode, d.selectFormat)
	if err != nil {
		return fmt.Errorf("GetEpisodeParts: %w", err)
	}

	d.OnFinishGetMetadata()

	if d.outputVideoPath != "" {
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
				// Start at the last existing segment
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
			d.OnProgress(float64(currentSegment) / float64(totalSegments), false)
			currentSegment++
			return nil
		}); err != nil {
			return fmt.Errorf("GetEpisodeAsTS: %w", err)
		}

		d.OnStartPostprocess()

		outputFileMP4, err := os.Create(d.outputVideoPath)
		if err != nil {
			return fmt.Errorf("create output file: %w", err)
		}
		defer outputFileMP4.Close()

		tsReader, tsWriter := io.Pipe()

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
				d.OnProgress(float64(i) / float64(totalSegments), true)
			}
			tsWriter.Close()
		}()
		if err := ConvertTSToMP4(tsReader, outputFileMP4); err != nil {
			return fmt.Errorf("convert MPEG-TS to mp4: %w", err)
		}
		if convertErr != nil {
			return fmt.Errorf("convert MPEG-TS to mp4: %w", err)
		}
		if err := os.RemoveAll(d.tmpDirPath); err != nil {
			return fmt.Errorf("remove temporary media directory: %w", err)
		}
	}

	if d.outputSubtitlePath != "" {
		d.OnStartDownloadSubtitles()

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
