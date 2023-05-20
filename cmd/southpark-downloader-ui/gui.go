package main

import (
	"context"
	"errors"
	"fmt"
	"path"
	"sync"
	"time"

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
    "github.com/adrg/xdg"
)

type GUIState struct {
	sync.RWMutex
	SelectedSeason   *Season
	EpisodeLists     *union.Union
	SeasonLists      *union.Union
}

type SeasonID struct {
	Title    string
	Language sp.Language
}

type GUI struct {
	MainWindow fyne.Window
	State     GUIState
	Cache     Cache
	Downloads *Downloads
}

func newGUI(mainWindow fyne.Window) *GUI {
	return &GUI{
		MainWindow: mainWindow,
		Cache: Cache{
			Seasons: make(map[sp.Language]Seasons),
		},
		Downloads: NewDownloads(3),
	}
}

func (g *GUI) makeGUI() fyne.CanvasObject {
	return container.NewAppTabs(
		container.NewTabItemWithIcon(
			"Episodes", theme.ListIcon(), g.makeEpisodesPanel()),
		container.NewTabItemWithIcon(
			"Downloads", theme.DownloadIcon(), g.makeDownloadsPanel()),
		container.NewTabItemWithIcon(
			"Search", theme.SearchIcon(), g.makeSearchPanel()),
        container.NewTabItemWithIcon(
            "Preferences", theme.SettingsIcon(), g.makePreferencesPanel()),
	)
}

func (g *GUI) makeEpisodesPanel() fyne.CanvasObject {
	episodeLists := union.New()
	seasonLists := union.New()

	g.State.Lock()
	g.State.EpisodeLists = episodeLists
	g.State.SeasonLists = seasonLists
	g.State.Unlock()


	var availableLanguages []string
	for _, v := range g.Cache.Region.AvailableLanguages {
		availableLanguages = append(availableLanguages, v.String())
	}

	languageSelect := widget.NewSelect(
		availableLanguages,
		func(languageStr string) {
			language, ok := sp.LanguageFromString(languageStr)
			if !ok {
				panic("logic error: nonexistent language selected")
			}
			if !seasonLists.Contains(languageStr) {
				seasons := g.makeSeasonList(language)
				seasonLists.Set(
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

	var cnt fyne.CanvasObject
	if fyne.CurrentDevice().IsMobile() {
		cnt = seasons
	} else {
		hs := container.NewHSplit(
			seasons,
			episodeLists,
		)
		hs.Offset = 0.15
		cnt = hs
	}
	return cnt
}

func (g *GUI) makeSeasonList(language sp.Language) fyne.CanvasObject {
	var seasonList *fetchableresource.FetchableResource
	seasonList = fetchableresource.New(
		context.Background(),
		makeProgressBarInfiniteTop(),
		func(ctx context.Context) (any, error) {
			g.Cache.Lock()
			seasons := g.Cache.Seasons[language].Seasons
			g.Cache.Unlock()
			if seasons == nil {
				updateCtx, _ := context.WithTimeout(ctx, 20 * time.Second)
				err := g.Cache.UpdateSeasons(updateCtx, language)
				if err != nil {
					return nil, err
				}
				g.Cache.Lock()
				seasons = g.Cache.Seasons[language].Seasons
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
					g.State.EpisodeLists.Set(
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
		func(err error) fyne.CanvasObject {
			label := widget.NewLabelWithStyle(
				fmt.Sprintf("Error loading seasons: %v", err),
				fyne.TextAlignCenter,
				fyne.TextStyle{},
			)
			label.Wrapping = fyne.TextWrapWord
			button := widget.NewButtonWithIcon("Retry", theme.ViewRefreshIcon(), func() {
				seasonList.Refetch()
			})
			return container.NewBorder(label, button, nil, nil)
		},
	)
	return seasonList
}

func (g *GUI) makeEpisodeList(season Season, seasonIndex int) fyne.CanvasObject {
	var episodeList *fetchableresource.FetchableResource
	episodeList = fetchableresource.New(
		context.Background(),
		makeProgressBarInfiniteTop(),
		func(ctx context.Context) (any, error) {
			if season.Episodes == nil {
				updateCtx, _ := context.WithTimeout(ctx, 20 * time.Second)
				err := g.Cache.UpdateEpisodes(updateCtx, season.Language, seasonIndex)
				if err != nil {
					return nil, err
				}
			}
			return g.Cache.Seasons[season.Language].Seasons[seasonIndex], nil
		},
		func(resource any) fyne.CanvasObject {
			season := resource.(Season)

			vb := container.NewVBox()
			for _, v := range season.Episodes {
				vb.Add(g.makeEpisode(v))
			}
			return container.NewVScroll(vb)
		},
		func(err error) fyne.CanvasObject {
			label := widget.NewLabelWithStyle(
				fmt.Sprintf("Error loading episodes: %v", err),
				fyne.TextAlignCenter,
				fyne.TextStyle{},
			)
			label.Wrapping = fyne.TextWrapWord
			button := widget.NewButtonWithIcon("Retry", theme.ViewRefreshIcon(), func() {
				episodeList.Refetch()
			})
			return container.NewBorder(label, button, nil, nil)
		},
	)
	return episodeList
}

func (g *GUI) makeEpisode(episode sp.Episode) fyne.CanvasObject {
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
			case DownloadDownloadingVideo:
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
		if DownloadStatus(s) == DownloadDownloadingVideo {
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
				g.MainWindow,
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
            save := func(
                // Useful on mobile only; if non-nil,
                // output gets saved to a temporary
                // file before being copied to finalOut
                finalOut fyne.URIWriteCloser,
            ) {
                storageBase := fyne.CurrentApp().Storage().RootURI().Path()
                tmpDir := path.Join(storageBase, "tmp_"+baseName)
                var outVidFile string
                var outSubFile string
                if finalOut == nil {
					outPath := fyne.CurrentApp().Preferences().StringWithFallback("DownloadURI", xdg.UserDirs.Download)
                    outVidFile = path.Join(outPath, baseName+".mp4")
                    outSubFile = path.Join(outPath, baseName+".vtt")
                } else {
                    outVidFile = path.Join(storageBase, baseName+".mp4")
                }

                handle, err := g.Downloads.Add(
                    context.Background(),
                    episode,
                    tmpDir,
                    outVidFile,
					outSubFile,
                    finalOut,
                    0,
                    status,
                    progress,
                )
                handle.StatusText = statusText
                if err != nil {
                    if finalOut != nil {
                        finalOut.Close()
                    }
                    dialog.ShowError(err, g.MainWindow)
                    return
                }

                cancelButton.OnTapped = func() {
                    handle.Cancel()
                    button.SetActive("Download")
                }

                button.SetActive("Cancel")

                go func() {
                    if finalOut != nil {
                        defer finalOut.Close()
                    }
                    defer button.SetActive("Done")

                    if err := handle.Do(); err != nil {
                        if !errors.Is(err, context.Canceled) {
                            dialog.ShowError(err, g.MainWindow)
                        }
                        return
                    }
                }()
            }

            if fyne.CurrentDevice().IsMobile() {
                saveDialog := dialog.NewFileSave(
                    func(out fyne.URIWriteCloser, err error) {
                        if err != nil {
                            if out != nil {
                                out.Close()
                            }
                            dialog.ShowError(err, g.MainWindow)
                            return
                        }
                        if out == nil {
                            return
                        }
                        save(out)
                    },
                    g.MainWindow,
                )
                saveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".mp4"}))
                saveDialog.SetFileName(baseName + ".mp4")
                saveDialog.Show()
            } else {
                save(nil)
            }
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

	return g.makeEpisodeBaseView(
		episode.RawThumbnailURL,
		episode.Title,
		episode.Description,
		loader,
		button,
	)
}

func (g *GUI) makeEpisodeBaseView(rawThumbnailURL sp.RawThumbnailURL, title string, description string, thumbnailOverlay fyne.CanvasObject, button fyne.CanvasObject) fyne.CanvasObject {
	var imgMinSize fyne.Size
	if fyne.CurrentDevice().IsMobile() {
		winSize := g.MainWindow.Content().Size()
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
				rawThumbnailURL.GetThumbnailURL(320, 180, true))
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

	text := widget.NewRichTextFromMarkdown("## " + title + "\n" + description)
	text.Wrapping = fyne.TextWrapWord

	thumbnailCnt := container.NewMax(thumbnail)
	if thumbnailOverlay != nil {
		thumbnailCnt.Add(thumbnailOverlay)
	}

	if fyne.CurrentDevice().IsMobile() {
		return container.NewPadded(
			container.NewBorder(
				thumbnailCnt,
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
			thumbnailCnt,
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

func (g *GUI) makePreferencesPanel() fyne.CanvasObject {
    getDLPathEntryPath := func() string {
        return fyne.CurrentApp().Preferences().StringWithFallback("DownloadURI", xdg.UserDirs.Download)
    }
    dlPathButton := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
        fo := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
            if err != nil {
                dialog.ShowError(err, g.MainWindow)
                return
            }
            if uri == nil {
                return
            }
            fyne.CurrentApp().Preferences().SetString("DownloadURI", uri.Path())
        }, g.MainWindow)
        uri := storage.NewFileURI(getDLPathEntryPath())
        list, err := storage.ListerForURI(uri)
        if err == nil {
            fo.SetLocation(list)
        }
        fo.Show()
    })
    dlPathLabel := widget.NewLabel("Download Save Path:")
    dlPathEntry := widget.NewEntry()
    dlPathEntry.OnChanged = func(s string) {
        fyne.CurrentApp().Preferences().SetString("DownloadURI", s)
    }
    dlPathEntry.Validator = func(s string) error {
        uri := storage.NewFileURI(s)
        canList, err := storage.CanList(uri)
        if err != nil {
            return err
        }
        if !canList {
            return errors.New("cannot list URI")
        }
        return nil
    }
    dlPathEntry.SetText(getDLPathEntryPath())
    fyne.CurrentApp().Preferences().AddChangeListener(func() {
        dlPathEntry.SetText(getDLPathEntryPath())
    })
    return container.NewVBox(
        container.NewBorder(
            nil,
            nil,
            dlPathLabel,
            dlPathButton,
            dlPathEntry,
        ),
    )
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
