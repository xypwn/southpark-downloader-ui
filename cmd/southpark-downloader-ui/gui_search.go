package main

import (
	"context"
	"errors"

	"github.com/xypwn/southpark-downloader-ui/pkg/gui/fetchableresource"
	sp "github.com/xypwn/southpark-downloader-ui/pkg/southpark"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (g *GUI) makeSearchPanel() fyne.CanvasObject {
	search := widget.NewEntry()

	results := fetchableresource.New(
		context.Background(),
		makeProgressBarInfiniteTop(),
		func(ctx context.Context) (any, error) {
			g.Cache.Lock()
			region := g.Cache.Region
			if len(g.Cache.Seasons) == 0 {
				g.Cache.Unlock()
				return nil, errors.New("no series available")
			}
			seriesMGID := g.Cache.Seasons[0].SeriesMGID
			g.Cache.Unlock()

			results, err := sp.Search(
				context.Background(),
				region,
				seriesMGID,
				search.Text,
				0,
				20,
			)
			if err != nil {
				return nil, err
			}
			return results, nil
		},
		func(resource any) fyne.CanvasObject {
			results := resource.([]sp.SearchResult)
			episodes := container.NewVBox()
			for _, v := range results {
				episodes.Add(g.makeEpisodeBaseView(
					v.RawThumbnailURL,
					v.Title,
					v.Description,
					nil,
					nil,
				))
			}
			return container.NewVScroll(episodes)
		},
		nil,
	)

	search.SetPlaceHolder("Search for an episode")
	search.OnSubmitted = func(string) {
		results.Refetch()
	}
	search.ActionItem = widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
		results.Refetch()
	})

	return container.NewBorder(
		search,
		nil,
		nil,
		nil,
		results,
	)
}
