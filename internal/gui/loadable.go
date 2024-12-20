package gui

import (
	"context"
	"sync"

	"github.com/xypwn/southpark-downloader-ui/pkg/asynctask"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type Loadable struct {
	widget.BaseWidget
	mtx      sync.RWMutex
	loader   fyne.CanvasObject
	retryBtn *widget.Button
	errStr   string
	errText  *widget.Label
	copyBtn  *widget.Button
	retry    fyne.CanvasObject
	content  *fyne.Container
	obj      fyne.CanvasObject
}

func NewLoadable(
	ctx context.Context,
	asyncFn func(ctx context.Context) (fyne.CanvasObject, error),
	setClipboard func(string),
) *Loadable {
	progress := widget.NewProgressBarInfinite()
	progress.Start()
	res := &Loadable{
		loader: container.NewBorder(
			progress,
			nil,
			nil,
			nil,
		),
		retryBtn: widget.NewButtonWithIcon("Retry", theme.ViewRefreshIcon(), func() {}),
		errText:  widget.NewLabel(""),
		copyBtn:  widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {}),
		content:  container.NewStack(),
	}
	res.ExtendBaseWidget(res)

	res.errText.Wrapping = fyne.TextWrapWord
	res.errText.Alignment = fyne.TextAlignCenter
	res.copyBtn.OnTapped = func() {
		res.mtx.RLock()
		defer res.mtx.RUnlock()
		setClipboard(res.errStr)
	}
	res.copyBtn.Importance = widget.LowImportance

	task := asynctask.New(
		ctx,
		func(ctx context.Context, _ struct{}, _ func(struct{})) (fyne.CanvasObject, error) {
			return asyncFn(ctx)
		},
		nil,
		func(obj fyne.CanvasObject, err error) {
			res.loader.Hide()
			if err != nil {
				res.mtx.Lock()
				defer res.mtx.Unlock()
				res.errStr = err.Error()
				res.errText.SetText("Error: " + res.errStr)
				res.retry.Show()
				return
			}
			res.retry.Hide()
			res.content.Add(obj)
		},
		nil,
	)

	res.retryBtn.OnTapped = func() {
		res.loader.Show()
		res.retry.Hide()
		task.Go(struct{}{})
	}

	task.Go(struct{}{})

	res.retry = container.NewBorder(
		container.NewVBox(
			res.errText,
			res.retryBtn,
			res.copyBtn,
		),
		nil,
		nil,
		nil,
	)

	res.obj = container.NewStack(res.loader, res.retry, res.content)

	res.loader.Show()
	res.retry.Hide()

	return res
}

func (r *Loadable) CreateRenderer() fyne.WidgetRenderer {
	r.mtx.RLock()
	defer r.mtx.RUnlock()

	return widget.NewSimpleRenderer(r.obj)
}
