package main

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

func main() {
	app := app.NewWithID("org.nobrain.southparkdownloaderui")
	window := app.NewWindow("Southpark Downloader")

	gui := newGUI()

	gui.Cache.UpdateRegion(context.Background())

	window.SetContent(gui.makeGUI())

	window.Resize(fyne.NewSize(800, 450))

	window.ShowAndRun()
}
