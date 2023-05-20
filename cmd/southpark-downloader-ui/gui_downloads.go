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

	priorityStrs := []string {
		"Very High",
		"High",
		"Normal",
		"Low",
		"Very Low",
	}

	priorities := widget.NewListWithData(
		g.Downloads.Handles,
		func() fyne.CanvasObject {
			sel := widget.NewSelect(priorityStrs, func(string) {})
			sel.Hide()
			sel.SetSelected("Normal")
			return sel
		},
		func(item binding.DataItem, obj fyne.CanvasObject) {
			v, err := item.(binding.Untyped).Get()
			if err != nil {
				return
			}
			handle := v.(*DownloadHandle)

			sel := obj.(*widget.Select)

			// Shitty hack, but we only want the selection
			// to be bound once
			if sel.Hidden {
				sel.OnChanged = func(s string) {
					var prio int
					for i, v := range priorityStrs {
						if v == s {
							prio = i - 2
							break
						}
					}
					_ = handle.Priority.Set(prio)
				}
				sel.Show()
			}
		},
	)

	hs2 := container.NewHSplit(statuses, priorities)
	hs2.Offset = 0.5
	hs1 := container.NewHSplit(episodes, hs2)
	hs1.Offset = 0.6
	return hs1
}
