package main

import (
	"context"
	"errors"
	"fmt"
	"path"
	"sync"

	"github.com/xypwn/southpark-downloader-ui/pkg/gui/fetchableresource"
	"github.com/xypwn/southpark-downloader-ui/pkg/gui/union"
	sp "github.com/xypwn/southpark-downloader-ui/pkg/southpark"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type GUIState struct {
	sync.RWMutex
	SelectedLanguage sp.Language
	SelectedSeason   *Season
	EpisodeLists     *union.Union
	SeasonLists      *union.Union
}

func (s *GUIState) getSelectedLanguage() sp.Language {
	s.RLock()
	defer s.RUnlock()
	return s.SelectedLanguage
}

type SeasonID struct {
	Title    string
	Language sp.Language
}

type GUI struct {
	State     GUIState
	Cache     Cache
	Downloads *Downloads
}

func newGUI() *GUI {
	return &GUI{
		Cache: Cache{
			Seasons: make(map[sp.Language][]Season),
		},
		Downloads: NewDownloads(3),
	}
}

func (g *GUI) makeGUI() fyne.CanvasObject {
	return container.NewAppTabs(
		container.NewTabItem(
			"Episodes", g.makeEpisodesPanel()),
		container.NewTabItem(
			"Downloads", g.makeDownloadsPanel()))
}

func (g *GUI) makeEpisodesPanel() fyne.CanvasObject {
	search := g.makeEpisodeSearch()
	episodeLists := union.New()
	seasonLists := union.New()

	g.State.Lock()
	g.State.EpisodeLists = episodeLists
	g.State.SeasonLists = seasonLists
	g.State.Unlock()


	languageSelect := widget.NewSelect(
		[]string{
			sp.LanguageEnglish.String(),
			sp.LanguageGerman.String(),
		}, func(languageStr string) {
			language, ok := sp.LanguageFromString(languageStr)
			if !ok {
				panic("logic error: nonexistent language selected")
			}
			if !seasonLists.Contains(languageStr) {
				seasons := g.makeSeasonsList(language)
				seasonLists.Add(
					union.NewItem(languageStr, seasons),
				)
			}
			seasonLists.SetActive(languageStr)
		})

	languageSelect.SetSelected(sp.LanguageEnglish.String())

	seasons := container.NewBorder(
		languageSelect,
		nil,
		nil,
		nil,
		seasonLists,
	)

	var mainCnt fyne.CanvasObject
	if fyne.CurrentDevice().IsMobile() {
		mainCnt = seasons
	} else {
		hs := container.NewHSplit(
			seasons,
			episodeLists,
		)
		hs.Offset = 0.15
		mainCnt = hs
	}

	return container.NewBorder(
		search,
		nil,
		nil,
		nil,
		mainCnt,
	)
}

func (g *GUI) makeEpisodeSearch() fyne.CanvasObject {
	search := widget.NewEntry()
	search.SetPlaceHolder("Search")
	search.ActionItem = widget.NewButtonWithIcon("", theme.SearchIcon(), func() {
	})
	return search
}

func (g *GUI) makeSeasonsList(language sp.Language) fyne.CanvasObject {
	return fetchableresource.New(
		context.Background(),
		makeProgressBarInfiniteTop(),
		func(ctx context.Context) (any, error) {
			g.Cache.Lock()
			seasons := g.Cache.Seasons[language]
			g.Cache.Unlock()
			if seasons == nil {
				err := g.Cache.UpdateSeasons(ctx, language)
				if err != nil {
					return nil, err
				}
				g.Cache.Lock()
				seasons = g.Cache.Seasons[language]
				g.Cache.Unlock()
			}
			return seasons, nil
		},
		func(resource any) fyne.CanvasObject {
			seasons := resource.([]Season)
			list := widget.NewList(
				func() int {
					g.Cache.Lock()
					defer g.Cache.Unlock()
					g.State.Lock()
					defer g.State.Unlock()

					return len(seasons)
				},
				func() fyne.CanvasObject {
					return widget.NewLabel("")
				},
				func(id widget.ListItemID, object fyne.CanvasObject) {
					g.Cache.Lock()
					defer g.Cache.Unlock()
					g.State.Lock()
					defer g.State.Unlock()

					object.(*widget.Label).SetText(seasons[len(seasons)-1-id].Title)
				})
			list.OnSelected = func(id int) {
				seasonIndex := len(seasons) - 1 - id
				season := seasons[seasonIndex]

				g.State.Lock()
				g.State.SelectedSeason = &seasons[seasonIndex]
				g.State.Unlock()

				if !g.State.EpisodeLists.Contains(season.Title) {
					g.State.EpisodeLists.Add(
						union.NewItem(
							season.Title,
							g.makeEpisodeList(season, seasonIndex),
						),
					)
				}

				g.State.EpisodeLists.SetActive(season.Title)

				if fyne.CurrentDevice().IsMobile() {
					list.UnselectAll()

					child := fyne.CurrentApp().NewWindow(season.Title)
					g.State.Lock()
					child.SetContent(g.State.EpisodeLists)
					g.State.Unlock()
					child.Show()
				}
			}
			return list
		},
		nil,
	)
}

func (g *GUI) makeEpisodeList(season Season, seasonIndex int) fyne.CanvasObject {
	return fetchableresource.New(
		context.Background(),
		makeProgressBarInfiniteTop(),
		func(ctx context.Context) (any, error) {
			if season.Episodes == nil {
				err := g.Cache.UpdateEpisodes(context.Background(), season.Language, seasonIndex)
				if err != nil {
					return nil, err
				}
			}
			return g.Cache.Seasons[season.Language][seasonIndex], nil
		},
		func(resource any) fyne.CanvasObject {
			season := resource.(Season)

			vb := container.NewVBox()
			for _, v := range season.Episodes {
				vb.Add(g.makeEpisode(v))
			}
			return container.NewVScroll(vb)
		},
		nil,
	)
}

func (g *GUI) makeEpisode(episode sp.Episode) fyne.CanvasObject {
	mainWindow := fyne.CurrentApp().Driver().AllWindows()[0]

	var imgMinSize fyne.Size
	if fyne.CurrentDevice().IsMobile() {
		winSize := fyne.CurrentApp().Driver().AllWindows()[0].Content().Size()
		var minWidth float32
		if winSize.Width < winSize.Height {
			minWidth = winSize.Width
		} else {
			minWidth = winSize.Height
		}
		imgMinSize = fyne.NewSize(160, minWidth*9/16)
	} else {
		imgMinSize = fyne.NewSize(160, 90)
	}

	var placeholder fyne.CanvasObject
	{
		img := canvas.NewImageFromResource(theme.FileImageIcon())
		img.FillMode = canvas.ImageFillContain
		img.ScaleMode = canvas.ImageScaleFastest
		img.SetMinSize(imgMinSize)
		placeholder = container.NewMax(
			img,
			makeProgressBarInfiniteBottom(),
		)
	}

	thumbnail := fetchableresource.New(
		context.Background(),
		placeholder,
		func(ctx context.Context) (any, error) {
			resource, err := fyne.LoadResourceFromURLString(
				episode.GetThumbnailURL(320, 180, true))
			if err != nil {
				return nil, err
			}
			return resource, nil
		},
		func(resource any) fyne.CanvasObject {
			img := canvas.NewImageFromResource(resource.(fyne.Resource))
			img.FillMode = canvas.ImageFillContain
			img.ScaleMode = canvas.ImageScaleFastest
			img.SetMinSize(imgMinSize)
			return img
		},
		nil,
	)

	text := widget.NewRichTextFromMarkdown("## " + episode.Title + "\n" + episode.Description)
	text.Wrapping = fyne.TextWrapWord

	status := binding.NewInt()
	statusText := binding.NewString()
	progress := binding.NewFloat()

	var loader fyne.CanvasObject
	var loadingBar *union.Union
	{
		progressBar := widget.NewProgressBarWithData(progress)

		// Text display handled by statusText / label in loader
		progressBar.TextFormatter = func() string { return "" }

		loadingBar = union.New(
			union.NewItem(
				"Infinite",
				makeProgressBarInfiniteBottom(),
			),
			union.NewItem(
				"Progress",
				progressBar,
			),
		)

		label := widget.NewLabelWithData(statusText)
		label.Alignment = fyne.TextAlignCenter
		label.TextStyle = fyne.TextStyle{Bold: true}

		loader = container.NewBorder(
			nil,
			container.NewMax(
				loadingBar,
				label,
			),
			nil,
			nil,
		)
		loader.Hide()
	}

	status.AddListener(binding.NewDataListener(
		func() {
			v, err := status.Get()
			if err != nil {
				return
			}
			switch DownloadStatus(v) {
			case DownloadNotStarted:
				loader.Hide()
				statusText.Set("Not started")
			case DownloadWaiting:
				loader.Show()
				loadingBar.SetActive("Infinite")
				statusText.Set("Waiting")
			case DownloadFetchingMetadata:
				loader.Show()
				loadingBar.SetActive("Infinite")
				statusText.Set("Fetching metadata")
			case DownloadDownloading:
				loader.Show()
				loadingBar.SetActive("Progress")
				// Text handled by progress
			case DownloadPostprocessing:
				loader.Show()
				loadingBar.SetActive("Progress")
				// Text handled by progress
			case DownloadCopying:
				loader.Show()
				loadingBar.SetActive("Infinite")
				statusText.Set("Copying")
			case DownloadDone:
				loader.Hide()
				statusText.Set("Done")
			case DownloadCanceled:
				loader.Hide()
				statusText.Set("Canceled")
			}
		}))

	progress.AddListener(binding.NewDataListener(func() {
		p, err := progress.Get()
		if err != nil {
			return
		}
		s, err := status.Get()
		if err != nil {
			return
		}
		var action string
		if DownloadStatus(s) == DownloadDownloading {
			action = "Downloading"
		} else {
			action = "Postprocessing"
		}
		statusText.Set(fmt.Sprintf("%v %.0f%%", action, p*100))
	}))

	var button *union.Union

	cancelButton := widget.NewButtonWithIcon(
		"",
		theme.CancelIcon(),
		func() {},
	)

	unavailableButton := widget.NewButtonWithIcon(
		"",
		theme.ErrorIcon(),
		func() {
			dialog.ShowInformation(
				"Episode unavailable",
				"This episode is currently unavailable",
				mainWindow,
			)
		},
	)

	doneButton := widget.NewButtonWithIcon(
		"",
		theme.ConfirmIcon(),
		func() {},
	)

	downloadButton := widget.NewButtonWithIcon(
		"",
		theme.DownloadIcon(),
		func() {
			baseName := sp.GetDownloadOutputFileName(episode)
			saveDialog := dialog.NewFileSave(
				func(out fyne.URIWriteCloser, err error) {
					if out == nil {
						return
					}
					if err != nil {
						out.Close()
						dialog.ShowError(err, mainWindow)
						return
					}

					fmt.Println(out)

					//dialog.ShowInformation("Path", out.URI().String(), mainWindow)

					storageBase := fyne.CurrentApp().Storage().RootURI().Path()
					tmpDir := path.Join(storageBase, "tmp_"+baseName)
					outFile := path.Join(storageBase, baseName+".mp4")
					//tmpDir := path.Join(fyne.CurrentApp().Storage().RootURI().Path(), "tmp_"+baseName)
					//outFile := path.Join("sdcard", "Spdl", baseName+".mp4")
					//storage.Writer(uri)
					handle, err := g.Downloads.Add(
						context.Background(),
						episode,
						tmpDir,
						outFile,
						out,
						0,
						status,
						progress,
					)
					handle.StatusText = statusText
					if err != nil {
						out.Close()
						dialog.ShowError(err, mainWindow)
						return
					}

					cancelButton.OnTapped = func() {
						handle.Cancel()
						button.SetActive("Download")
					}

					button.SetActive("Cancel")

					go func() {
						defer out.Close()
						defer button.SetActive("Done")

						if err := handle.Do(); err != nil {
							if !errors.Is(err, context.Canceled) {
								dialog.ShowError(err, mainWindow)
							}
							return
						}
					}()
				},
				mainWindow,
			)
			saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".mp4"}))
			saveDialog.SetFileName(baseName + ".mp4")
			saveDialog.Show()
		},
	)

	button = union.New(
		union.NewItem(
			"Cancel",
			cancelButton,
		),
		union.NewItem(
			"Done",
			doneButton,
		),
		union.NewItem(
			"Download",
			downloadButton,
		),
		union.NewItem(
			"Unavailable",
			unavailableButton,
		),
	)
	if episode.Unavailable {
		button.SetActive("Unavailable")
	} else {
		button.SetActive("Download")
	}

	if fyne.CurrentDevice().IsMobile() {
		return container.NewPadded(
			container.NewBorder(
				container.NewMax(thumbnail, loader),
				nil,
				nil,
				button,
				text,
			),
		)
	} else {
		return container.NewBorder(
			nil,
			nil,
			container.NewMax(thumbnail, loader),
			button,
			text,
		)
	}
}

func (g *GUI) makeDownloadsPanel() fyne.CanvasObject {
	episodes := widget.NewListWithData(
		g.Downloads.Handles,
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Wrapping = fyne.TextWrapWord
			return label
		},
		func(item binding.DataItem, obj fyne.CanvasObject) {
			v, err := item.(binding.Untyped).Get()
			if err != nil {
				return
			}
			handle := v.(*DownloadHandle)

			label := obj.(*widget.Label)

			label.SetText(fmt.Sprintf(
				"S%v E%v: %v",
				handle.Episode.SeasonNumber,
				handle.Episode.EpisodeNumber,
				handle.Episode.Title,
			))
		},
	)

	statuses := widget.NewListWithData(
		g.Downloads.Handles,
		func() fyne.CanvasObject {
			label := widget.NewLabel("PLACEHOLDER")
			label.Wrapping = fyne.TextWrapWord
			return label
		},
		func(item binding.DataItem, obj fyne.CanvasObject) {
			v, err := item.(binding.Untyped).Get()
			if err != nil {
				return
			}
			handle := v.(*DownloadHandle)

			label := obj.(*widget.Label)

			// Shitty hack, but we only want the label
			// to be bound once
			if label.Text == "PLACEHOLDER" {
				label.Text = ""
				label.Bind(handle.StatusText)
			}
		},
	)

	hs := container.NewHSplit(episodes, statuses)
	hs.Offset = 0.6
	return hs
}

func makeProgressBarInfiniteTop() fyne.CanvasObject {
	return container.NewBorder(
		widget.NewProgressBarInfinite(),
		nil,
		nil,
		nil,
	)
}

func makeProgressBarInfiniteBottom() fyne.CanvasObject {
	return container.NewBorder(
		nil,
		widget.NewProgressBarInfinite(),
		nil,
		nil,
	)
}
