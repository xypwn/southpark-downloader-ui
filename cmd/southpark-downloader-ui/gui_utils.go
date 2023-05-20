package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

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
