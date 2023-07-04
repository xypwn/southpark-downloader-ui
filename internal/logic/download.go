package logic

import (
	"context"
	"errors"
	"fmt"
	"os"

	"sync"

	"github.com/xypwn/southpark-downloader-ui/pkg/asynctask"
	"github.com/xypwn/southpark-downloader-ui/pkg/data"
	sp "github.com/xypwn/southpark-downloader-ui/pkg/southpark"
	"github.com/xypwn/southpark-downloader-ui/pkg/taskqueue"
)

type Quality int

const (
	QualityBest  Quality = Quality(1 << 24) // big number
	Quality1080p Quality = 1080
	Quality720p  Quality = 720
	Quality540p  Quality = 540
	Quality432p  Quality = 432
	Quality360p  Quality = 360
	Quality288p  Quality = 288
	Quality216p  Quality = 216
)

func DefaultQualities() []Quality {
	return []Quality{
		QualityBest,
		Quality1080p,
		Quality720p,
		Quality540p,
		Quality432p,
		Quality360p,
		Quality288p,
		Quality216p,
	}
}

func (v Quality) String() string {
	switch v {
	case QualityBest:
		return "Best"
	default:
		return fmt.Sprintf("%vp", int(v))
	}
}

type DownloadStatus int

const (
	DownloadStatusWaiting DownloadStatus = iota
	DownloadStatusFetchingMetadata
	DownloadStatusDownloadingVideo
	DownloadStatusDownloadingSubtitles
	DownloadStatusPostprocessing
	DownloadStatusDone
	DownloadStatusInterrupted
)

type DownloadProgress struct {
	Status DownloadStatus
	Value  float64 // -1 if unable to determine
}

func (v DownloadProgress) String() string {
	var text string
	switch v.Status {
	case DownloadStatusWaiting:
		text = "Waiting"
	case DownloadStatusFetchingMetadata:
		text = "Fetching Metadata"
	case DownloadStatusDownloadingVideo:
		text = "Downloading Video"
	case DownloadStatusDownloadingSubtitles:
		text = "Downloading Subtitles"
	case DownloadStatusPostprocessing:
		text = "Postprocessing"
	case DownloadStatusDone:
		text = "Done"
	}
	if v.Value == -1 {
		return text
	}
	return fmt.Sprintf("%v %.0f%%", text, v.Value*100)
}

type DownloadParams struct {
	Episode            sp.Episode
	MaxQuality         Quality
	TmpDirPath         string
	OutputVideoPath    string
	OutputSubtitlePath string
}

type Download struct {
	*asynctask.AsyncTask[struct{}, DownloadProgress, struct{}]
	mtx            sync.RWMutex
	params         DownloadParams
	progress       *data.Binding[DownloadProgress]
	progressClient *data.Client[DownloadProgress]
}

func (dl *Download) Params() DownloadParams {
	return dl.params
}

func (dl *Download) Progress() DownloadProgress {
	var res DownloadProgress
	dl.progressClient.Examine(func(dp DownloadProgress) {
		res = dp
	})
	return res
}

func (dl *Download) ProgressBinding() *data.Binding[DownloadProgress] {
	return dl.progress
}

type Downloads struct {
	*data.ListBinding[*Download]
	client *data.ListClient[*Download]
	queue  *taskqueue.TaskQueue[*Download]
}

func NewDownloads(cfgClient *data.Client[*Config], onError func(error)) *Downloads {
	var res *Downloads
	res = &Downloads{
		ListBinding: data.NewListBinding[*Download](),
	}
	var nConcurrentInit int
	cfgClient.Examine(func(c *Config) {
		nConcurrentInit = c.ConcurrentDownloads
	})
	res.client = res.NewClient()
	res.queue = taskqueue.New(
		nConcurrentInit,
		func(enqueued []*Download) int {
			var toMatch *Download
			res.client.Examine(func(arr []*Download) {
				for _, v := range arr {
					if v.Progress().Status == DownloadStatusWaiting {
						toMatch = v
						break
					}
				}
			})
			if toMatch == nil {
				onError(errors.New("internal error: Downloads: TaskQueue requested selection, but no downloads are waiting"))
				return 0
			}
			for i, v := range enqueued {
				if v == toMatch {
					return i
				}
			}
			onError(errors.New("internal error: Downloads: next item to be downloaded not found in queue"))
			return 0
		},
	)
	cfgClient.AddListener(func(c *Config) {
		res.queue.SetSize(c.ConcurrentDownloads)
	})
	return res
}

func (dls *Downloads) Add(ctx context.Context, params DownloadParams, onError func(error)) *Download {
	res := &Download{
		params:   params,
		progress: data.NewBinding[DownloadProgress](),
	}
	res.progressClient = res.progress.NewClient()

	doDownload := func(ctx context.Context, _ struct{}, setProgress func(DownloadProgress)) (struct{}, error) {
		dl := sp.NewDownloader(
			ctx,
			params.Episode,
			params.TmpDirPath,
			params.OutputVideoPath,
			func(fmts []sp.HLSFormat) (sp.HLSFormat, error) {
				// fmts are already sorted from best to worst
				for _, v := range fmts {
					if v.Height <= uint(params.MaxQuality) {
						return v, nil
					}
				}
				return sp.HLSFormat{}, fmt.Errorf("no viable format found for maximum quality of %v", params.MaxQuality.String())
			},
			params.OutputSubtitlePath,
		)

		setProgress(DownloadProgress{
			Status: DownloadStatusWaiting,
			Value:  -1,
		})

		if err := dls.queue.Acquire(ctx, res); err != nil {
			setProgress(DownloadProgress{
				Status: DownloadStatusInterrupted,
				Value:  -1,
			})
			return struct{}{}, err
		}
		defer dls.queue.Release()

		dl.OnStatusChanged = func(status sp.DownloaderStatus, progress float64) {
			var s DownloadStatus
			switch status {
			case sp.DownloaderStatusFetchingMetadata:
				s = DownloadStatusFetchingMetadata
			case sp.DownloaderStatusDownloadingVideo:
				s = DownloadStatusDownloadingVideo
			case sp.DownloaderStatusDownloadingSubtitles:
				s = DownloadStatusDownloadingSubtitles
			case sp.DownloaderStatusPostprocessing:
				s = DownloadStatusPostprocessing
			}
			setProgress(DownloadProgress{
				Status: s,
				Value:  progress,
			})
		}

		if err := dl.Do(); err != nil {
			setProgress(DownloadProgress{
				Status: DownloadStatusInterrupted,
				Value:  -1,
			})
			return struct{}{}, err
		}

		setProgress(DownloadProgress{
			Status: DownloadStatusDone,
			Value:  -1,
		})

		return struct{}{}, nil
	}

	res.AsyncTask = asynctask.New(
		ctx,
		doDownload,
		nil,
		func(_ struct{}, err error) {
			if err != nil {
				// Remove from downloads
				dls.client.Change(func(data []*Download) []*Download {
					for i, v := range data {
						if v == res {
							return append(data[:i], data[i+1:]...)
						}
					}
					onError(errors.New("internal error: remove download from downloads: cannot find download"))
					return data
				})

				if !errors.Is(err, context.Canceled) {
					onError(err)
				}
			}
		},
		func(dp DownloadProgress) {
			res.progressClient.Change(func(DownloadProgress) DownloadProgress {
				return dp
			})
		},
	)

	dls.client.Change(func(arr []*Download) []*Download {
		return append([]*Download{res}, arr...)
	})

	return res
}

// Used for caching downloads
type DownloadInfo struct {
	Params DownloadParams
}

type DownloadsInfo []DownloadInfo

func NewDownloadsInfo() DownloadsInfo {
	return DownloadsInfo{}
}

// Initializes dls to current data in dlInfoStor, then continually
// updates dlInfoStor if any changes occur in dls
func ConnectDownloadsToDownloadsInfo(ctx context.Context, dls *Downloads, dlInfoStor *StorageItem[DownloadsInfo], onError func(error)) {
	dlInfoCl := dlInfoStor.NewClient()

	dlInfoCl.Examine(func(di DownloadsInfo) {
		for _, v := range di {
			tmpDirPresent := false
			downloadFilesPresent := false
			if info, err := os.Stat(v.Params.OutputVideoPath); err == nil && !info.IsDir() {
				if info, err := os.Stat(v.Params.OutputSubtitlePath); err == nil && !info.IsDir() {
					downloadFilesPresent = true
				}
			}
			if _, err := os.Stat(v.Params.TmpDirPath); err == nil {
				tmpDirPresent = true
			}

			if tmpDirPresent {
				dl := dls.Add(ctx, v.Params, onError)
				dl.Go(struct{}{})
			} else {
				if downloadFilesPresent {
					dl := dls.Add(ctx, v.Params, onError)
					cl := dl.ProgressBinding().NewClient()
					cl.Change(func(DownloadProgress) DownloadProgress {
						return DownloadProgress{
							Status: DownloadStatusDone,
							Value:  -1,
						}
					})
					dl.ProgressBinding().RemoveClient(cl)
				}
			}
		}
	})

	dlsCl := dls.NewClient()
	dlsCl.AddListener(func(arr []*Download) {
		dlInfoCl.Change(func(di DownloadsInfo) DownloadsInfo {
			di = di[:0]
			for _, v := range arr {
				di = append(di, DownloadInfo{
					Params: v.Params(),
				})
			}
			return di
		})
	})
}
