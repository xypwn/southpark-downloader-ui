package main

import (
	"errors"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (g *GUI) makePreferencesPanel() fyne.CanvasObject {
	getDLPathEntryPath := func() string {
		return g.getDownloadPath()
	}
	dlPathButton := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
		fo := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, g.CurrentWindow.Get())
				return
			}
			if uri == nil {
				return
			}
			g.setDownloadPath(uri.Path())
		}, g.CurrentWindow.Get())
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
		g.setDownloadPath(s)
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
	g.App.Preferences().AddChangeListener(func() {
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
