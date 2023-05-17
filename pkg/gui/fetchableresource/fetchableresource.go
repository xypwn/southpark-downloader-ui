package fetchableresource

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/xypwn/southpark-downloader-ui/pkg/gui/union"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type FetchableResource struct {
	widget.BaseWidget

	mtx        sync.Mutex
	ctx        context.Context
	fetching atomic.Bool
	fetch func(context.Context) (any, error)
	makeContent func(any) fyne.CanvasObject
	makeErrContent func(error) fyne.CanvasObject
	err        error
	onErr      func(error)
	content    *union.Union
}

func New(
	ctx context.Context,
	loading fyne.CanvasObject,
	fetch func(context.Context) (any, error),
	makeContent func(any) fyne.CanvasObject,
	makeErrContent func(error) fyne.CanvasObject,
) *FetchableResource {
	res := &FetchableResource{
		ctx: ctx,
		fetch: fetch,
		makeContent: makeContent,
		makeErrContent: makeErrContent,
		content: union.New(
			union.NewItem(
				"Loading",
				loading),
		),
	}

	res.tryFetchAsync()

	res.ExtendBaseWidget(res)
	return res
}

func (fr *FetchableResource) SetOnError(onErr func(error)) {
	fr.mtx.Lock()
	defer fr.mtx.Unlock()
	fr.onErr = onErr
}

func (fr *FetchableResource) IsFetching() bool {
	return fr.fetching.Load()
}

func (fr *FetchableResource) Refetch() error {
	if fr.IsFetching() {
		return errors.New("FetchableResource.Refetch: already fetching a resource")
	}

	fr.tryFetchAsync()
	return nil
}

func (fr *FetchableResource) CreateRenderer() fyne.WidgetRenderer {
	fr.ExtendBaseWidget(fr)

	return widget.NewSimpleRenderer(fr.content)
}

func (fr *FetchableResource) tryFetchAsync() {
	go func() {
		fr.fetching.Store(true)
		defer fr.fetching.Store(false)

		fr.mtx.Lock()
		fr.content.SetActive("Loading")
		fr.mtx.Unlock()

		resource, err := fr.fetch(fr.ctx)

		fr.mtx.Lock()
		defer fr.mtx.Unlock()

		if err != nil {
			var errContent fyne.CanvasObject
			if fr.makeErrContent != nil {
				errContent = fr.makeErrContent(err)
			}
			if fr.onErr != nil {
				fr.onErr(err)
			}
			fr.content.Set(union.NewItem("Error", errContent))
			fr.content.SetActive("Error")
			return
		}

		fr.content.Set(union.NewItem("Content", fr.makeContent(resource)))
		fr.content.SetActive("Content")
	}()
}
