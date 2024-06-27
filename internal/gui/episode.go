package gui

import (
	"bytes"
	"context"
	"fmt"
	"image/color"
	"path"
	"strings"
	"sync"

	"github.com/xypwn/southpark-downloader-ui/internal/logic"
	"github.com/xypwn/southpark-downloader-ui/pkg/asynctask"
	"github.com/xypwn/southpark-downloader-ui/pkg/data"
	"github.com/xypwn/southpark-downloader-ui/pkg/gui/ellipsislabel"
	"github.com/xypwn/southpark-downloader-ui/pkg/httputils"
	sp "github.com/xypwn/southpark-downloader-ui/pkg/southpark"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/skratchdot/open-golang/open"
)

type Episode struct {
	widget.BaseWidget
	mtx               sync.RWMutex
	thumbnailTask     *asynctask.AsyncTask[string, struct{}, *canvas.Image]
	placeholderImage  *canvas.Image
	thumbnailImageCnt *fyne.Container
	thumbnailImage    *canvas.Image
	thumbnailSize     fyne.Size
	thumbnailProgress *widget.ProgressBarInfinite
	thumbnailOverlay  *canvas.Rectangle
	thumbnailText     *canvas.Text
	progressInfinite  *widget.ProgressBarInfinite
	progressDiscrete  *widget.ProgressBar
	progressText      *widget.Label
	title             *ellipsislabel.EllipsisLabel
	description       *ellipsislabel.EllipsisLabel
	button            *widget.Button
	obj               fyne.CanvasObject
}

func NewEpisode(
	ctx context.Context,
	onInfo func(title, text string),
	onError func(error),
	dls *logic.Downloads,
	cfgClient *data.Client[*logic.Config],
	metadata sp.EpisodeMetadata,
	getEpisode func() (sp.Episode, error),
	showSeasonNumber bool,
	showEpisodeNumber bool,
	mobile bool,
) (ep *Episode, destroy func()) {
	var doDestroyMtx sync.Mutex
	doDestroy := func() {}

	var res *Episode
	res = &Episode{
		thumbnailTask: asynctask.New(
			ctx,
			func(ctx context.Context, url string, setProgress func(struct{})) (*canvas.Image, error) {
				data, err := httputils.GetBodyWithContext(ctx, url)
				if err != nil {
					return nil, err
				}

				return canvas.NewImageFromReader(bytes.NewReader(data), url), nil
			},
			func() {
				res.mtx.Lock()
				defer res.mtx.Unlock()

				res.thumbnailImage = nil
				res.thumbnailImageCnt.RemoveAll()
				res.thumbnailProgress.Show()
				res.placeholderImage.Show()
			},
			func(img *canvas.Image, err error) {
				res.thumbnailProgress.Hide()

				if err != nil {
					return
				}

				res.placeholderImage.Hide()

				res.mtx.Lock()
				defer res.mtx.Unlock()

				img.SetMinSize(res.thumbnailSize)
				img.FillMode = canvas.ImageFillContain
				res.thumbnailImageCnt.Add(img)
				res.thumbnailImage = img
			},
			nil,
		),
		placeholderImage:  canvas.NewImageFromResource(theme.FileImageIcon()),
		thumbnailImageCnt: container.NewMax(),
		thumbnailProgress: widget.NewProgressBarInfinite(),
		thumbnailOverlay:  canvas.NewRectangle(color.RGBA{16, 16, 16, 160}),
		thumbnailText:     canvas.NewText("", color.RGBA{255, 255, 255, 255}),
		progressInfinite:  widget.NewProgressBarInfinite(),
		progressDiscrete:  widget.NewProgressBar(),
		progressText:      widget.NewLabel(""),
		title:             ellipsislabel.New(""),
		description:       ellipsislabel.New(""),
		button:            widget.NewButtonWithIcon("", theme.MoreHorizontalIcon(), func() {}),
	}
	res.ExtendBaseWidget(res)

	res.title.SetStyle(widget.RichTextStyleSubHeading)

	res.placeholderImage.FillMode = canvas.ImageFillContain

	res.thumbnailProgress.Hide()
	res.progressInfinite.Hide()
	res.progressDiscrete.Hide()
	res.progressDiscrete.TextFormatter = func() string { return "" }
	res.thumbnailOverlay.Hide()
	res.thumbnailText.Hide()
	res.progressText.Alignment = fyne.TextAlignCenter
	res.progressText.TextStyle = fyne.TextStyle{Bold: true}
	res.progressText.Hide()

	res.thumbnailSize = fyne.NewSize(192, 108)
	res.placeholderImage.SetMinSize(res.thumbnailSize)

	thumbnail := container.NewMax(
		res.placeholderImage,
		res.thumbnailImageCnt,
		res.thumbnailOverlay,
		res.thumbnailText,
		container.NewBorder(
			nil,
			container.NewMax(
				res.thumbnailProgress,
				res.progressInfinite,
				res.progressDiscrete,
				res.progressText,
			),
			nil,
			nil,
		),
	)

	if mobile {
		res.obj = container.NewBorder(
			res.title,
			res.button,
			thumbnail,
			nil,
			res.description,
		)
	} else {
		res.obj = container.NewBorder(
			nil,
			nil,
			thumbnail,
			res.button,
			container.NewBorder(
				res.title,
				nil,
				nil,
				nil,
				res.description,
			),
		)
	}

	res.mtx.Lock()
	titlePrefix := ""
	if showSeasonNumber {
		titlePrefix += fmt.Sprintf("S%v", metadata.SeasonNumber)
	}
	if showEpisodeNumber {
		titlePrefix += fmt.Sprintf("E%v", metadata.EpisodeNumber)
	}
	if showSeasonNumber || showEpisodeNumber {
		titlePrefix += ": "
	}
	res.title.SetText(titlePrefix + metadata.Title)
	res.description.SetText(metadata.Description)

	if metadata.Unavailable {
		res.thumbnailOverlay.Show()
		res.thumbnailText.Show()
		res.thumbnailText.Text = "Unavailable"
		res.thumbnailText.TextSize = theme.TextHeadingSize()
		res.thumbnailText.TextStyle = fyne.TextStyle{Bold: true}
		res.thumbnailText.Alignment = fyne.TextAlignCenter
		res.button.SetIcon(theme.ErrorIcon())
		res.button.OnTapped = func() {
			onInfo(
				"Episode Unavailable",
				"This episode is currently unavailable due to rights and restrictions",
			)
		}
	} else {
		res.thumbnailOverlay.Hide()
		res.thumbnailText.Hide()
		res.button.SetIcon(theme.DownloadIcon())

		var dl *logic.Download

		dlsClient := dls.NewClient()
		dlsClient.Examine(func(arr []*logic.Download) {
			for _, v := range arr {
				if v.Params().Episode.Is(metadata) {
					dl = v
					break
				}
			}
		})
		dls.RemoveClient(dlsClient)

		var doDownload func()

		listenToDL := func() (destroy func()) {
			onProgress := func(dp logic.DownloadProgress) {
				if dp.Status == logic.DownloadStatusDone {
					res.progressDiscrete.Hide()
					res.progressInfinite.Hide()
					res.progressText.Hide()
					res.button.SetIcon(theme.MediaPlayIcon())
					res.button.OnTapped = func() {
						open.Start(dl.Params().OutputVideoPath)
					}
					return
				}

				if dp.Status == logic.DownloadStatusInterrupted {
					res.progressDiscrete.Hide()
					res.progressInfinite.Hide()
					res.progressText.Hide()
					res.button.SetIcon(theme.ViewRefreshIcon())
					res.button.OnTapped = doDownload
					return
				}

				res.button.SetIcon(theme.CancelIcon())
				res.button.OnTapped = func() {
					dl.Cancel()
				}

				if dp.Value == -1 {
					res.progressDiscrete.Hide()
					res.progressInfinite.Show()
				} else {
					res.progressDiscrete.SetValue(dp.Value)
					res.progressInfinite.Hide()
					res.progressDiscrete.Show()
				}
				res.progressText.Show()
				res.progressText.SetText(dp.String())
			}

			client := dl.ProgressBinding().NewClient()
			client.AddListener(func(dp logic.DownloadProgress) {
				res.mtx.Lock()
				onProgress(dp)
				res.mtx.Unlock()
			})

			onProgress(dl.Progress())

			return func() {
				dl.ProgressBinding().RemoveClient(client)
			}
		}

		res.button.OnTapped = func() {
			var downloadPath string
			var maxQuality logic.Quality
			var outputFilePattern string
			cfgClient.Examine(func(c *logic.Config) {
				downloadPath = c.DownloadPath
				maxQuality = c.MaximumQuality
				outputFilePattern = c.OutputFilePattern
			})

			toValidFilename := func(s string) string {
				var result strings.Builder
				for i := 0; i < len(s); i++ {
					b := s[i]
					if ('a' <= b && b <= 'z') ||
						('A' <= b && b <= 'Z') ||
						('0' <= b && b <= '9') {
						result.WriteByte(b)
					} else {
						result.WriteByte('_')
					}
				}
				return result.String()
			}

			ep, err := getEpisode()
			if err != nil {
				onError(err)
			}

			outputBase := strings.NewReplacer(
				"$S", fmt.Sprintf("%02v", ep.SeasonNumber),
				"$E", fmt.Sprintf("%02v", ep.EpisodeNumber),
				"$L", ep.Language.String(),
				"$T", toValidFilename(ep.Title),
				"$Q", maxQuality.String(),
			).Replace(outputFilePattern)

			doDownload = func() {
				dl = dls.Add(
					ctx,
					logic.DownloadParams{
						Episode:            ep,
						MaxQuality:         maxQuality,
						TmpDirPath:         path.Join(downloadPath, "~TMP_"+outputBase),
						OutputVideoPath:    path.Join(downloadPath, outputBase+".mp4"),
						OutputSubtitlePath: path.Join(downloadPath, outputBase+".vtt"),
					},
					onError,
				)

				doDestroyMtx.Lock()
				doDestroy = listenToDL()
				doDestroyMtx.Unlock()

				dl.Go(struct{}{})
			}

			doDownload()
		}

		if dl != nil {
			doDestroyMtx.Lock()
			doDestroy = listenToDL()
			doDestroyMtx.Unlock()
		}
	}

	res.thumbnailText.Refresh()

	thumbnailURL := metadata.GetThumbnailURL(
		uint(res.thumbnailSize.Width),
		uint(res.thumbnailSize.Height),
		true,
	)
	res.mtx.Unlock()

	for !res.thumbnailTask.Go(thumbnailURL) {
		//_ = res.thumbnailTask.Cancel()
	}

	return res, func() {
		doDestroyMtx.Lock()
		doDestroy()
		doDestroyMtx.Unlock()
	}
}

func (e *Episode) CreateRenderer() fyne.WidgetRenderer {
	e.mtx.RLock()
	defer e.mtx.RUnlock()

	return widget.NewSimpleRenderer(e.obj)
}
