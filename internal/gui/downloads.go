package gui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/xypwn/southpark-downloader-ui/internal/logic"
	"github.com/xypwn/southpark-downloader-ui/pkg/data"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func stringFilterFunc(value, pattern string) bool {
	for _, sp := range strings.Split(pattern, " ") {
		if !strings.Contains(strings.ToUpper(value), strings.ToUpper(sp)) {
			return false
		}
	}
	return true
}

func downloadFilterFunc(dl *logic.Download, filter downloadFilter) bool {
	status := dl.Progress().Status
	switch filter.Status {
	case downloadStatusFilterEnqueued:
		if status != logic.DownloadStatusWaiting {
			return false
		}
	case downloadStatusFilterDownloading:
		if status != logic.DownloadStatusFetchingMetadata &&
			status != logic.DownloadStatusDownloadingVideo &&
			status != logic.DownloadStatusPostprocessing &&
			status != logic.DownloadStatusDownloadingSubtitles {
			return false
		}
	case downloadStatusFilterCompleted:
		if status != logic.DownloadStatusDone {
			return false
		}
	}

	ep := dl.Params().Episode
	s := fmt.Sprintf(
		"%v Season%v Episode%v S%vE%v %v %v",
		ep.URL,
		ep.SeasonNumber,
		ep.EpisodeNumber,
		ep.SeasonNumber,
		ep.EpisodeNumber,
		ep.Title,
		ep.Description,
	)
	return stringFilterFunc(s, filter.Query)
}

type DownloadItem struct {
	widget.BaseWidget
	text   *widget.Label
	status *widget.Label
	reset  func()
	obj    fyne.CanvasObject
	mtx    sync.Mutex // for fields that aren't already thread-safe
}

func NewDownloadItem() *DownloadItem {
	res := &DownloadItem{
		text:   widget.NewLabel("PLACEHOLDER"),
		status: widget.NewLabel("PLACEHOLDER"),
		reset:  func() {},
	}
	res.ExtendBaseWidget(res)

	res.status.Alignment = fyne.TextAlignTrailing

	res.obj = container.NewBorder(
		nil,
		nil,
		res.text,
		nil,
		res.status,
	)

	return res
}

func (di *DownloadItem) Set(dl *logic.Download) {
	di.mtx.Lock()
	defer di.mtx.Unlock()

	di.reset()

	ep := dl.Params().Episode
	di.text.SetText(fmt.Sprintf("S%vE%v: %v", ep.SeasonNumber, ep.EpisodeNumber, ep.Title))
	client := dl.ProgressBinding().NewClient()
	statusChangedFunc := func(dp logic.DownloadProgress) {
		di.status.SetText(dp.String())
	}
	client.AddListener(statusChangedFunc)
	client.Examine(statusChangedFunc)

	di.reset = func() {
		dl.ProgressBinding().RemoveClient(client)
	}
}

func (di *DownloadItem) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(di.obj)
}

type downloadStatusFilter int

const (
	downloadStatusFilterAll downloadStatusFilter = iota
	downloadStatusFilterEnqueued
	downloadStatusFilterDownloading
	downloadStatusFilterCompleted
)

type downloadFilter struct {
	Query  string
	Status downloadStatusFilter
}

type Downloads struct {
	widget.BaseWidget
	filter *data.ListFilter[*logic.Download, downloadFilter]
	obj    fyne.CanvasObject
}

func NewDownloads(dls *logic.Downloads) *Downloads {
	res := &Downloads{
		filter: data.NewListFilter(
			dls.ListBinding,
			downloadFilterFunc,
		),
	}
	res.ExtendBaseWidget(res)

	sourceClient := dls.NewClient()
	filteredClient := res.filter.Filtered().NewClient()
	patternClient := res.filter.Pattern().NewClient()

	var list *widget.List
	{
		// I found fyne bindings to be the easiest when displaying
		// my custom binding lists
		data := binding.NewUntypedList()
		list = widget.NewListWithData(
			data,
			func() fyne.CanvasObject {
				return NewDownloadItem()
			},
			func(item binding.DataItem, obj fyne.CanvasObject) {
				wid := obj.(*DownloadItem)
				dl, _ := item.(binding.Untyped).Get()
				wid.Set(dl.(*logic.Download))
			},
		)
		filteredClient.AddListener(func(arr []*logic.Download) {
			list.UnselectAll()
			res := make([]any, len(arr))
			for i, v := range arr {
				res[i] = v
			}
			_ = data.Set(res)
		})
	}

	prioDown := func() {}
	prioUp := func() {}
	prioToBottom := func() {}
	prioToTop := func() {}

	prioDownBtn := widget.NewButtonWithIcon("", theme.MenuDropDownIcon(), func() { prioDown() })
	prioUpBtn := widget.NewButtonWithIcon("", theme.MenuDropUpIcon(), func() { prioUp() })
	prioToBottomBtn := widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() { prioToBottom() })
	prioToTopBtn := widget.NewButtonWithIcon("", theme.MoveUpIcon(), func() { prioToTop() })
	setPrioBtnsEnabled := func(enabled bool) {
		if enabled {
			prioDownBtn.Enable()
			prioUpBtn.Enable()
			prioToBottomBtn.Enable()
			prioToTopBtn.Enable()
		} else {
			prioDownBtn.Disable()
			prioUpBtn.Disable()
			prioToBottomBtn.Disable()
			prioToTopBtn.Disable()
		}
	}
	setPrioBtnsEnabled(false)

	list.OnSelected = func(id widget.ListItemID) {
		filteredClient.Examine(func(flt []*logic.Download) {
			if id < 0 || id >= len(flt) {
				return
			}
			sourceClient.Examine(func(src []*logic.Download) {
				item := flt[id]
				move := func(src *[]*logic.Download, id int, diff int) (int, bool) {
					id2 := id + diff
					if id2 < len(*src) && id2 >= 0 {
						(*src)[id], (*src)[id2] = (*src)[id2], (*src)[id]
						return id2, true
					}
					return id, false
				}
				var idInSrc int
				{
					for i, v := range src {
						if v == item {
							idInSrc = i
							break
						}
					}
				}
				setPrioBtnsEnabled(true)
				if id == 0 {
					prioUpBtn.Disable()
				}
				if idInSrc == 0 {
					prioToTopBtn.Disable()
				}
				if id == len(flt)-1 {
					prioDownBtn.Disable()
				}
				if idInSrc == len(src)-1 {
					prioToBottomBtn.Disable()
				}
				prioDown = func() {
					sourceClient.Change(func(src []*logic.Download) []*logic.Download {
						patternClient.Examine(func(pattern downloadFilter) {
							i := idInSrc
							ok := true
							for ok {
								i, ok = move(&src, i, 1)
								if i == 0 || downloadFilterFunc(src[i-1], pattern) {
									break
								}
							}
						})
						return src
					})
					// Calling Select here, because it would call OnSelected internally,
					// which would call sourceClient.Examine, which would cause a deadlock
					// if it were in sourceClient.Change
					list.Select(id + 1) // Select ignores bound violations
				}
				prioUp = func() {
					sourceClient.Change(func(src []*logic.Download) []*logic.Download {
						patternClient.Examine(func(pattern downloadFilter) {
							i := idInSrc
							ok := true
							for ok {
								i, ok = move(&src, i, -1)
								if i == len(src)-1 || downloadFilterFunc(src[i+1], pattern) {
									break
								}
							}
						})
						return src
					})
					list.Select(id - 1)
				}
				prioToBottom = func() {
					sourceClient.Change(func(src []*logic.Download) []*logic.Download {
						i := idInSrc
						ok := true
						for ok {
							i, ok = move(&src, i, 1)
						}
						return src
					})
					list.Select(list.Length() - 1) // Select ignores bound violations
				}
				prioToTop = func() {
					sourceClient.Change(func(src []*logic.Download) []*logic.Download {
						i := idInSrc
						ok := true
						for ok {
							i, ok = move(&src, i, -1)
						}
						return src
					})
					list.Select(0)
				}
			})
		})
	}

	list.OnUnselected = func(id widget.ListItemID) {
		prioDown = func() {}
		prioUp = func() {}
		prioToBottom = func() {}
		prioToTop = func() {}

		setPrioBtnsEnabled(false)
	}

	search := widget.NewEntry()
	search.PlaceHolder = "Search Downloads"
	search.ActionItem = widget.NewIcon(theme.SearchIcon())
	search.OnChanged = func(s string) {
		list.UnselectAll()
		patternClient.Change(func(df downloadFilter) downloadFilter {
			df.Query = s
			return df
		})
	}

	sel := widget.NewSelect([]string{"All", "Enqueued", "Downloading", "Completed"}, func(s string) {
		list.UnselectAll()
		patternClient.Change(func(df downloadFilter) downloadFilter {
			switch s {
			case "All":
				df.Status = downloadStatusFilterAll
			case "Enqueued":
				df.Status = downloadStatusFilterEnqueued
			case "Downloading":
				df.Status = downloadStatusFilterDownloading
			case "Completed":
				df.Status = downloadStatusFilterCompleted
			}
			return df
		})
	})
	sel.SetSelectedIndex(0)

	var clearFiltersButton *widget.Button
	content := container.NewMax()
	{
		clearFiltersButton = widget.NewButtonWithIcon(
			"Clear Filters",
			theme.ContentClearIcon(),
			func() {
				search.SetText("")
				sel.SetSelectedIndex(0)
			},
		)
		clearFiltersButton.Importance = widget.LowImportance

		emptyPlaceholder := widget.NewRichText(
			&widget.TextSegment{
				Style: widget.RichTextStyle{
					Alignment: fyne.TextAlignCenter,
					Inline:    false,
					SizeName:  theme.SizeNameHeadingText,
				},
				Text: "Nothing here",
			},
			&widget.TextSegment{
				Style: widget.RichTextStyle{
					Alignment: fyne.TextAlignCenter,
					Inline:    true,
					SizeName:  theme.SizeNameText,
				},
				Text: "Go to Episodes and start downloading ;)",
			},
		)
		noSearchResultsPlaceholder := container.NewVBox(
			widget.NewRichText(
				&widget.TextSegment{
					Style: widget.RichTextStyle{
						Alignment: fyne.TextAlignCenter,
						Inline:    false,
						SizeName:  theme.SizeNameHeadingText,
					},
					Text: "No results found",
				},
			),
		)

		placeholder := container.NewMax()

		var lock sync.Mutex
		init := true
		empty := false
		prevEmpty := false
		dataChangedFunc := func(arr []*logic.Download) {
			lock.Lock()
			defer lock.Unlock()
			empty = len(arr) == 0
			if init || empty != prevEmpty {
				content.RemoveAll()
				if empty {
					content.Add(placeholder)
				} else {
					content.Add(list)
				}
				prevEmpty = empty
			}
			init = false
		}
		patternChangedFunc := func(df downloadFilter) {
			lock.Lock()
			defer lock.Unlock()
			placeholder.RemoveAll()
			if df.Query != "" || df.Status != downloadStatusFilterAll {
				placeholder.Add(noSearchResultsPlaceholder)
				clearFiltersButton.Enable()
			} else {
				placeholder.Add(emptyPlaceholder)
				clearFiltersButton.Disable()
			}
		}

		// Listen to changes from anywhere (including this widget's search widgets)
		res.filter.Filtered().NewClient().AddListener(dataChangedFunc)
		res.filter.Pattern().NewClient().AddListener(patternChangedFunc)

		// Invoke once for current data
		filteredClient.Examine(dataChangedFunc)
		patternClient.Examine(patternChangedFunc)
	}

	res.obj = container.NewBorder(
		container.NewBorder(
			nil,
			nil,
			sel,
			container.NewHBox(
				clearFiltersButton,
				prioDownBtn,
				prioUpBtn,
				prioToBottomBtn,
				prioToTopBtn,
			),
			search,
		),
		nil,
		nil,
		nil,
		container.NewPadded(content),
	)

	return res
}

func (dll *Downloads) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(dll.obj)
}
