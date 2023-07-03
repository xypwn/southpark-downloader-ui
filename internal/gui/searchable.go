package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type Searchable struct {
	widget.BaseWidget
	noResults fyne.CanvasObject
	content   *fyne.Container
	obj       fyne.CanvasObject
}

func NewSearchable() *Searchable {
	res := &Searchable{
		content: container.NewMax(),
	}
	return res
}
