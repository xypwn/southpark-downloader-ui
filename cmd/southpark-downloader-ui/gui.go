package main

import (
	"sync"

	"github.com/adrg/xdg"
	"github.com/xypwn/southpark-downloader-ui/pkg/gui/union"
	sp "github.com/xypwn/southpark-downloader-ui/pkg/southpark"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
)

type GUIState struct {
	sync.Mutex
	SelectedSeason *Season
	EpisodeLists   *union.Union
	SeasonLists    *union.Union
}

type SeasonID struct {
	Title    string
	Language sp.Language
}

type GUICurrentWindow struct {
	sync.Mutex
	CurrentWindow fyne.Window
}

func (w *GUICurrentWindow) Get() fyne.Window {
	w.Lock()
	win := w.CurrentWindow
	w.Unlock()
	return win
}

func (w *GUICurrentWindow) Set(win fyne.Window) {
	w.Lock()
	w.CurrentWindow = win
	w.Unlock()
}

type GUI struct {
	App           fyne.App
	CurrentWindow GUICurrentWindow
	State         GUIState
	Cache         Cache
	Downloads     *Downloads
}

func newGUI(app fyne.App, baseWindow fyne.Window) *GUI {
	return &GUI{
		App:           app,
		CurrentWindow: GUICurrentWindow{CurrentWindow: baseWindow},
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

func (g *GUI) newWindow(title string) fyne.Window {
	prevWin := g.CurrentWindow.Get()
	win := g.App.NewWindow(title)
	win.SetOnClosed(func() {
		g.CurrentWindow.Set(prevWin)
	})
	return win
}

func (g *GUI) getDownloadPath() string {
	return g.App.Preferences().StringWithFallback("DownloadPath", xdg.UserDirs.Download)
}

func (g *GUI) setDownloadPath(path string) {
	g.App.Preferences().SetString("DownloadPath", path)
}
