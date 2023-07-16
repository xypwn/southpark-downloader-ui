package ellipsislabel

import (
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type EllipsisLabel struct {
	widget.BaseWidget

	mtx      sync.RWMutex
	text     string
	ellipsis string
	content  *widget.RichText
}

func New(text string) *EllipsisLabel {
	res := &EllipsisLabel{
		text:     text,
		ellipsis: "...",
		content:  widget.NewRichTextWithText(text),
	}
	res.ExtendBaseWidget(res)
	return res
}

func (el *EllipsisLabel) SetText(text string) {
	el.mtx.Lock()
	el.text = text
	el.mtx.Unlock()

	el.Refresh()
}

func (el *EllipsisLabel) SetStyle(style widget.RichTextStyle) {
	el.mtx.Lock()
	el.content.Segments[0].(*widget.TextSegment).Style = style
	el.mtx.Unlock()

	el.Refresh()
}

func (el *EllipsisLabel) SetEllipsis(ellipsis string) {
	el.mtx.Lock()
	el.ellipsis = ellipsis
	el.mtx.Unlock()

	el.Refresh()
}

func (el *EllipsisLabel) CreateRenderer() fyne.WidgetRenderer {
	el.mtx.RLock()
	defer el.mtx.RUnlock()

	el.ExtendBaseWidget(el)
	res := &ellipsisLabelRenderer{
		ellipsisLabel: el,
		objects:       []fyne.CanvasObject{el.content},
	}
	return res
}

type ellipsisLabelRenderer struct {
	ellipsisLabel *EllipsisLabel
	objects       []fyne.CanvasObject
}

func (r *ellipsisLabelRenderer) Destroy() {
}

func (r *ellipsisLabelRenderer) Layout(size fyne.Size) {
	r.ellipsisLabel.mtx.Lock()
	defer r.ellipsisLabel.mtx.Unlock()

	innerPadding := theme.InnerPadding()
	lineSpacing := theme.LineSpacing()
	ellipsis := r.ellipsisLabel.ellipsis
	maxTextWidth := size.Width - innerPadding*2

	segment := r.ellipsisLabel.content.Segments[0].(*widget.TextSegment)
	textSize := fyne.CurrentApp().Settings().Theme().Size(segment.Style.SizeName)
	height := textSize
	remaining := r.ellipsisLabel.text
	res := ""
	i := 0
	for len(remaining) > 0 {
		nextHeight := height + textSize + lineSpacing
		ellipsize := nextHeight > size.Height-innerPadding*2
		last := ellipsize

		//fmt.Println(size.Height - innerPadding * 2, nextHeight, ellipsize)

		wordBreak := -1
		var lnLen int
		if ellipsize {
			lnLen = maxTextLenWithWidth(remaining+ellipsis, maxTextWidth, segment.Style.TextStyle, textSize)
			lnLen -= len(ellipsis)
			if lnLen == len(remaining) {
				ellipsize = false
			}
		} else {
			lnLen = maxTextLenWithWidth(remaining, maxTextWidth, segment.Style.TextStyle, textSize)
			if lnLen != len(remaining) {
				for i := 0;; i++ {
					if lnLen-i < 0 {
						break
					}
					if lnLen-i >= len(remaining) {
						continue
					}
					if remaining[lnLen-i] == ' ' {
						wordBreak = i
						break
					}
				}
			}
		}
		if lnLen <= 0 {
			break
		}
		
		if wordBreak != -1 {
			lnLen -= wordBreak
		}

		ln := remaining[:lnLen]
		res += ln
		if ellipsize {
			res += ellipsis
		}
		res += "\n"
		
		if wordBreak != -1 {
			remaining = remaining[lnLen+1:]
		} else {
			remaining = remaining[lnLen:]
		}

		if ellipsize || last {
			break
		}

		height = nextHeight

		i++
	}
	segment.Text = res

	//fmt.Println("DONE", size.Height - innerPadding * 2, height)

	r.objects[0].Resize(size)
	r.objects[0].Refresh()
}

func (r *ellipsisLabelRenderer) MinSize() fyne.Size {
	r.ellipsisLabel.mtx.RLock()
	defer r.ellipsisLabel.mtx.RUnlock()

	innerPadding := theme.InnerPadding()

	segment := r.ellipsisLabel.content.Segments[0].(*widget.TextSegment)
	textSize := fyne.CurrentApp().Settings().Theme().Size(segment.Style.SizeName)

	return fyne.NewSize(innerPadding*2, textSize+innerPadding*2)
}

func (r *ellipsisLabelRenderer) Refresh() {
	r.ellipsisLabel.mtx.RLock()
	defer r.ellipsisLabel.mtx.RUnlock()

	r.objects[0].Refresh()
}

func (r *ellipsisLabelRenderer) Objects() []fyne.CanvasObject {
	r.ellipsisLabel.mtx.RLock()
	defer r.ellipsisLabel.mtx.RUnlock()

	return r.objects
}

func maxTextLenWithWidth(text string, target float32, style fyne.TextStyle, textSize float32) int {
	width := func(i int) float32 {
		return fyne.MeasureText(text[:i], textSize, style).Width
	}

	if width(len(text)) <= target {
		return len(text)
	}

	min := 0
	max := len(text)
	for min <= max {
		mid := (min + max) / 2
		midWidth := width(mid)
		if target < midWidth {
			max = mid - 1
		} else if target > midWidth {
			min = mid + 1
		} else {
			min = mid
			break
		}
	}
	if min-1 < 0 {
		return 0
	}
	return min - 1
}
