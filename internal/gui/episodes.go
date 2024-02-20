package gui

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/xypwn/southpark-downloader-ui/internal/logic"
	"github.com/xypwn/southpark-downloader-ui/pkg/data"
	sp "github.com/xypwn/southpark-downloader-ui/pkg/southpark"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type EpisodeList struct {
	widget.BaseWidget
	obj fyne.CanvasObject
}

func NewEpisodeList(
	ctx context.Context,
	season logic.Season,
	dls *logic.Downloads,
	cacheClient *data.Client[*logic.Cache],
	cfgClient *data.Client[*logic.Config],
	onInfo func(title, text string),
	onError func(error),
	setClipboard func(string),
	mobile bool,
) (episodeList *EpisodeList, destroy func()) {
	var destroyMtx sync.Mutex
	var destroyFns []func()

	res := &EpisodeList{
		obj: NewLoadable(
			ctx,
			func(ctx context.Context) (fyne.CanvasObject, error) {
				var eps []sp.Episode

				update := false
				cacheClient.Examine(func(c *logic.Cache) {
					s := c.Series[season.Region.Host].Seasons[season.Language][season.Index]
					if s.Episodes == nil {
						update = true
					} else {
						eps = s.Episodes
					}
				})

				if update {
					var mgid string
					var err error
					eps, mgid, err = logic.GetSeason(ctx, season.Season)
					if err != nil {
						return nil, err
					}
					cacheClient.Change(func(c *logic.Cache) *logic.Cache {
						s := &c.Series[season.Region.Host].Seasons[season.Language][season.Index]
						s.Episodes = eps
						s.MGID = mgid
						return c
					})
				}

				vbox := container.NewVBox()
				for _, v := range eps {
					ep := v
					epWid, destroy := NewEpisode(
						ctx,
						onInfo,
						onError,
						dls,
						cfgClient,
						v.EpisodeMetadata,
						func() (sp.Episode, error) {
							return ep, nil
						},
						false,
						true,
						mobile,
					)
					destroyMtx.Lock()
					destroyFns = append(destroyFns, destroy)
					destroyMtx.Unlock()
					vbox.Add(container.NewPadded(epWid))
				}
				return container.NewVScroll(container.NewPadded(vbox)), nil
			},
			setClipboard,
		),
	}
	res.ExtendBaseWidget(res)

	return res, func() {
		destroyMtx.Lock()
		for _, fn := range destroyFns {
			fn()
		}
		destroyMtx.Unlock()
	}
}

func (el *EpisodeList) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(el.obj)
}

type EpisodesPanel struct {
	widget.BaseWidget
	obj fyne.CanvasObject
}

func NewEpisodesPanel(
	ctx context.Context,
	dls *logic.Downloads,
	cacheStor *logic.StorageItem[*logic.Cache],
	cfgClient *data.Client[*logic.Config],
	onInfo func(title, text string),
	onError func(error),
	setClipboard func(string),
	mobile bool,
) *EpisodesPanel {
	res := &EpisodesPanel{}
	res.ExtendBaseWidget(res)

	cache := cacheStor.NewClient()

	res.obj = NewLoadable(
		ctx,
		func(ctx context.Context) (fyne.CanvasObject, error) {
			region, seasons, mgid, err := logic.GetSeries(ctx)
			if err != nil {
				return nil, err
			}

			cache.Change(func(c *logic.Cache) *logic.Cache {
				c.Series[region.Host] = logic.Series{
					Region:  region,
					Seasons: seasons,
					MGID:    mgid,
				}
				return c
			})

			cleanupEpisodesFn := func() {}

			episodes := container.NewMax()

			seasonLists := make(map[sp.Language]*widget.List)
			var selectedSeason atomic.Int32
			for k, v := range seasons {
				seasons := v
				seasonList := widget.NewList(
					func() int {
						return len(seasons)
					},
					func() fyne.CanvasObject {
						return widget.NewLabel("PLACEHOLDER")
					},
					func(id widget.ListItemID, obj fyne.CanvasObject) {
						obj.(*widget.Label).SetText(seasons[len(seasons)-1-id].Title)
					},
				)
				seasonList.OnSelected = func(id widget.ListItemID) {
					selectedSeason.Store(int32(id))
					cleanupEpisodesFn()
					season := seasons[len(seasons)-1-id]
					episodes.RemoveAll()
					episodeList, destroy := NewEpisodeList(ctx, season, dls, cache, cfgClient, onInfo, onError, setClipboard, mobile)
					cleanupEpisodesFn = destroy
					episodes.Add(episodeList)
				}
				seasonLists[k] = seasonList
			}

			if len(region.AvailableLanguages) == 0 {
				return nil, errors.New("no languages available in your region")
			}

			languageOpts := make([]string, len(region.AvailableLanguages))
			for i, v := range region.AvailableLanguages {
				languageOpts[i] = v.String()
			}

			seasonsCnt := container.NewMax()

			type searchQuery struct {
				Language sp.Language
				Text     string
			}
			query := data.NewBinding[searchQuery]()
			querySetterCl := query.NewClient()
			queryListenerCl := query.NewClient()

			var selectedLanguage atomic.Int32
			languageSel := widget.NewSelect(
				languageOpts,
				func(s string) {
					var newLang sp.Language
					for _, v := range region.AvailableLanguages {
						if s == v.String() {
							newLang = v
							break
						}
					}
					seasonsCnt.RemoveAll()
					seasonsCnt.Add(seasonLists[newLang])
					seasonLists[newLang].UnselectAll()
					seasonLists[newLang].Select(int(selectedSeason.Load()))
					querySetterCl.Change(func(sq searchQuery) searchQuery {
						selectedLanguage.Store(int32(newLang))
						sq.Language = newLang
						return sq
					})
				},
			)
			languageSel.SetSelectedIndex(0)

			languageSelHelp := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
				onInfo(
					"Available Languages",
					"The South Park web service is highly regional.\nTo gain access to German episodes, you need a German IP address.\nYou can use a VPN to achieve this.",
				)
			})
			languageSelHelp.Hide()

			if len(region.AvailableLanguages) == 1 {
				languageSel.Disable()
				languageSelHelp.Show()
			}

			split := container.NewHSplit(
				container.NewBorder(
					nil,
					nil,
					nil,
					nil,
					seasonsCnt,
				),
				episodes,
			)
			if mobile {
				split.Horizontal = false
				split.SetOffset(0.2)
			} else {
				split.SetOffset(0.1)
			}

			mainCnt := container.NewMax(split)

			var clearSearchButton *widget.Button
			var cleanupSearchResultsFns []func()
			queryListenerCl.AddListener(func(sq searchQuery) {
				for _, v := range cleanupSearchResultsFns {
					v()
				}
				cleanupSearchResultsFns = nil
				mainCnt.RemoveAll()
				selLanguage := sp.Language(selectedLanguage.Load())
				if sq.Text != "" {
					clearSearchButton.Enable()
					mainCnt.Add(NewLoadable(ctx,
						func(ctx context.Context) (fyne.CanvasObject, error) {
							text := "Results for \"" + sq.Text + "\" in " + selLanguage.String() + ":"
							results, err := sp.Search(ctx, region, mgid, sq.Text, 0, 35)
							if err != nil {
								return nil, err
							}
							resultFound := false
							vbox := container.NewVBox()
							for _, v := range results {
								result := v
								if result.Language == selLanguage {
									ep, destroy := NewEpisode(
										ctx,
										onInfo,
										onError,
										dls,
										cfgClient,
										result,
										func() (sp.Episode, error) {
											return sp.GetEpisode(ctx, region, result.URL)
										},
										true,
										true,
										mobile,
									)
									vbox.Add(ep)
									cleanupSearchResultsFns = append(cleanupSearchResultsFns, destroy)
									resultFound = true
								}
							}
							if !resultFound {
								text = "No Results for \"" + sq.Text + "\" in " + selLanguage.String() + " :("
							}
							return container.NewPadded(
								container.NewBorder(
									widget.NewRichText(
										&widget.TextSegment{
											Style: widget.RichTextStyle{
												Alignment: fyne.TextAlignCenter,
												Inline:    false,
												SizeName:  theme.SizeNameHeadingText,
											},
											Text: text,
										},
									),
									nil,
									nil,
									nil,
									container.NewVScroll(vbox),
								),
							), nil
						},
						setClipboard,
					))
				} else {
					clearSearchButton.Disable()
					mainCnt.Add(split)
					seasonLists[selLanguage].UnselectAll()
					seasonLists[selLanguage].Select(int(selectedSeason.Load()))
				}
			})

			search := widget.NewEntry()
			search.PlaceHolder = "Search Episodes"
			search.ActionItem = widget.NewIcon(theme.SearchIcon())
			search.OnChanged = func(s string) {
				querySetterCl.Change(func(sq searchQuery) searchQuery {
					sq.Text = s
					return sq
				})
			}

			clearSearchButton = widget.NewButtonWithIcon(
				"Clear Search",
				theme.ContentClearIcon(),
				func() {
					search.SetText("")
				},
			)
			clearSearchButton.Importance = widget.LowImportance
			clearSearchButton.Disable()

			languageAndInfo := container.NewBorder(
				nil,
				nil,
				nil,
				languageSelHelp,
				languageSel,
			)
			var searchAndLanguage *fyne.Container
			if mobile {
				searchAndLanguage = container.NewVBox(
					container.NewBorder(
						nil,
						nil,
						nil,
						clearSearchButton,
						search,
					),
					languageAndInfo,
				)
			} else {
				searchAndLanguage = container.NewBorder(
					nil,
					nil,
					languageAndInfo,
					clearSearchButton,
					search,
				)
			}

			return container.NewBorder(
				searchAndLanguage,
				nil,
				nil,
				nil,
				mainCnt,
			), nil
		},
		setClipboard,
	)

	return res
}

func (e *EpisodesPanel) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(e.obj)
}
