package fetchableresource

import (
	"context"
	"errors"
	"sync"

	"southpark-downloader-ui/pkg/gui/union"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type FetchableResource struct {
	widget.BaseWidget

	mtx        sync.Mutex
	ctx        context.Context
	isCanceled bool
	err        error
	content    *union.Union
}

func New(
	ctx context.Context,
	loading fyne.CanvasObject,
	fetch func(context.Context) (any, error),
	makeContent func(any) fyne.CanvasObject,
	canceled fyne.CanvasObject,
) *FetchableResource {
	res := &FetchableResource{
		ctx: ctx,
		content: union.New(
			union.NewItem(
				"Loading",
				loading),
			union.NewItem(
				"Canceled",
				canceled),
		),
	}

	go func() {
		resource, err := fetch(ctx)

		res.mtx.Lock()
		defer res.mtx.Unlock()

		if err != nil {
			if errors.Is(err, context.Canceled) {
				res.content.SetActive("Canceled")
				res.isCanceled = true
			} else {
				res.err = err
			}
			return
		}

		res.content.Add(union.NewItem("Content", makeContent(resource)))
		res.content.SetActive("Content")
	}()

	res.ExtendBaseWidget(res)
	return res
}

func (fr *FetchableResource) IsCanceled() bool {
	fr.mtx.Lock()
	defer fr.mtx.Unlock()
	return fr.isCanceled
}

func (fr *FetchableResource) GetError() error {
	fr.mtx.Lock()
	defer fr.mtx.Unlock()
	return fr.err
}

func (fr *FetchableResource) CreateRenderer() fyne.WidgetRenderer {
	fr.ExtendBaseWidget(fr)

	return widget.NewSimpleRenderer(fr.content)
}
