package main

import (
	"context"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/dialog"
)

func main() {
	app := app.NewWithID("org.nobrain.southparkdownloaderui")
	window := app.NewWindow("Southpark Downloader")

	gui := newGUI(window)

	err := gui.Cache.UpdateRegion(context.Background())
	if err != nil {
		dialog.ShowError(err, window)
	}

	window.SetContent(gui.makeGUI())

	window.Resize(fyne.NewSize(800, 450))

	window.ShowAndRun()
}
