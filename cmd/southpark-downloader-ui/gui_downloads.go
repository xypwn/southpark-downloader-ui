package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

func (g *GUI) makeDownloadsPanel() fyne.CanvasObject {
	priorityStrs := []string {
		"Very High",
		"High",
		"Normal",
		"Low",
		"Very Low",
	}

	downloads := widget.NewListWithData(
		g.Downloads.Handles,
		func() fyne.CanvasObject {
			cnt := container.NewMax(container.NewBorder(
				nil,
				nil,
				widget.NewLabel("PLACEHOLDER"),
				widget.NewSelect([]string{"PLACEHOLDER"}, func(string) {}),
				widget.NewLabel("PLACEHOLDER"),
			))
			cnt.Hide()
			return cnt
		},
		func(item binding.DataItem, obj fyne.CanvasObject) {
			v, err := item.(binding.Untyped).Get()
			if err != nil {
				return
			}
			handle := v.(*DownloadHandle)

			cnt := obj.(*fyne.Container)

			if cnt.Hidden {
				label := widget.NewLabel(fmt.Sprintf(
					"S%v E%v: %v",
					handle.Episode.SeasonNumber,
					handle.Episode.EpisodeNumber,
					handle.Episode.Title,
				))

				status := widget.NewLabelWithData(handle.StatusText)
				status.Alignment = fyne.TextAlignTrailing

				priority := widget.NewSelect(priorityStrs,
					func(s string) {
						var prio int
						for i, v := range priorityStrs {
							if v == s {
								prio = i - 2
								break
							}
						}
						_ = handle.Priority.Set(prio)
					},
				)
				priority.SetSelected("Normal")

				cnt.RemoveAll()
				cnt.Add(container.NewBorder(
					nil,
					nil,
					label,
					priority,
					status,
				))
				cnt.Show()
			}
		},
	)

	downloads.OnSelected = func(widget.ListItemID) {
		downloads.UnselectAll()
	}

	return downloads
}
