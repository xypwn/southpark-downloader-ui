package main

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"

	priosem "github.com/xypwn/southpark-downloader-ui/pkg/prioritysemaphore"
	sp "github.com/xypwn/southpark-downloader-ui/pkg/southpark"

	"fyne.io/fyne/v2/data/binding"
)

type Cache struct {
	sync.RWMutex
	Region  sp.RegionInfo
	Seasons map[sp.Language]Seasons
}

func (c *Cache) UpdateRegion(ctx context.Context) error {
	region, err := sp.GetRegionInfo(ctx)
	if err != nil {
		return err
	}

	c.Lock()
	defer c.Unlock()
	c.Region = region
	return nil
}

func (c *Cache) UpdateSeasons(ctx context.Context, language sp.Language) error {
	c.RLock()
	region := c.Region
	c.RUnlock()

	seasons, seriesMGID, err := sp.GetSeasons(ctx, region, language)
	if err != nil {
		return err
	}

	var res []Season

	for _, v := range seasons {
		res = append(res, Season{
			Season:   v,
			Episodes: nil,
		})
	}

	c.Lock()
	defer c.Unlock()
	c.Seasons[language] = Seasons{
		Seasons: res,
		SeriesMGID: seriesMGID,
	}

	return nil
}

func (c *Cache) UpdateEpisodes(ctx context.Context, language sp.Language, seasonIndex int) error {
	c.RLock()
	if seasonIndex >= len(c.Seasons[language].Seasons) {
		c.RUnlock()
		return errors.New("invalid season number")
	}
	season := c.Seasons[language].Seasons[seasonIndex]
	c.RUnlock()

	episodes, err := sp.GetEpisodes(ctx, season.Season)
	if err != nil {
		return err
	}

	c.Lock()
	c.Seasons[language].Seasons[seasonIndex].Episodes = episodes
	c.Unlock()

	return nil
}

type Seasons struct {
	Seasons []Season
	SeriesMGID string
}

type Season struct {
	sp.Season
	Episodes []sp.Episode
}

type DownloadStatus int

const (
	DownloadNotStarted DownloadStatus = iota
	DownloadWaiting
	DownloadFetchingMetadata
	DownloadDownloading
	DownloadPostprocessing
	DownloadCopying // Only appears on mobile
	DownloadDone
	DownloadCanceled
)

type DownloadHandle struct {
	Context    context.Context
	Do         func() error // Can be called asynchronously
	Cancel     func()
	Status     binding.Int // Of type DownloadStatus
	Progress   binding.Float // Either download or postprocessing, depending on status
	StatusText binding.String // Optional, managed by user
	Episode    sp.Episode
}

type Downloads struct {
	*priosem.Semaphore
	Handles binding.UntypedList // List[*DownloadHandle]

	mtx sync.RWMutex
}

func NewDownloads(nSimultaneousDownloads int) *Downloads {
	return &Downloads{
		Semaphore: priosem.New(nSimultaneousDownloads),
		Handles:   binding.NewUntypedList(),
	}
}

func (d *Downloads) Add(
	ctx context.Context,
	episode sp.Episode,
	tmpDirPath string,
	outputFilePath string,
	finalOutput io.WriteCloser, // Useful on mobile only, pass nil to not use
	priority int,
	statusData binding.Int, // Of type DownloadStatus
	progressData binding.Float,
) (*DownloadHandle, error) {
	dlCtx, cancel := context.WithCancel(ctx)
	handle := &DownloadHandle{
		Context:  dlCtx,
		Cancel:   cancel,
		Status:   statusData,
		Progress: progressData,
		Episode:  episode,
	}
	handle.Status.Set(int(DownloadNotStarted))
	dler := sp.NewDownloader(
		dlCtx,
		episode,
		tmpDirPath,
		outputFilePath,
		func(formats []sp.HLSFormat) (sp.HLSFormat, error) {
			if len(formats) > 0 {
				return formats[0], nil
			} else {
				return sp.HLSFormat{}, errors.New("no formats available")
			}
		},
	)
	dler.OnFinishGetMetadata = func() {
		handle.Status.Set(int(DownloadDownloading))
	}
	dler.OnProgress = func(progress float64, postprocessing bool) {
		handle.Progress.Set(progress)
	}
	dler.OnStartPostprocess = func() {
		handle.Status.Set(int(DownloadPostprocessing))
	}
	handle.Do = func() error {
		handle.Status.Set(int(DownloadWaiting))

		if err := d.Acquire(dlCtx, priority); err != nil {
			if errors.Is(err, context.Canceled) {
				handle.Status.Set(int(DownloadCanceled))
			}
			return err
		}
		defer d.Release()

		handle.Status.Set(int(DownloadFetchingMetadata))

		if err := dler.Do(); err != nil {
			if errors.Is(err, context.Canceled) {
				handle.Status.Set(int(DownloadCanceled))
			}
			return err
		}

		if finalOutput != nil {
			f, err := os.Open(outputFilePath)
			if err != nil {
				return err
			}

			handle.Status.Set(int(DownloadCopying))

			_, err = io.Copy(finalOutput, f)
			if err != nil {
				return err
			}

			os.Remove(outputFilePath)
		}

		handle.Status.Set(int(DownloadDone))
		return nil
	}
	d.mtx.Lock()
	err := d.Handles.Append(handle)
	d.mtx.Unlock()
	if err != nil {
		return nil, err
	}
	return handle, nil
}
