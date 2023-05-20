package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

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
