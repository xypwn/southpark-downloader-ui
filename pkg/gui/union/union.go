package union

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type Union struct {
	widget.BaseWidget

	objects      []fyne.CanvasObject
	activeObject int
	objectIdxMap map[any]int
}

func New(items ...*UnionItem) *Union {
	res := &Union{
		objectIdxMap: make(map[any]int),
	}
	for _, v := range items {
		res.Set(v)
	}

	res.ExtendBaseWidget(res)
	return res
}

func (u *Union) Set(item *UnionItem) {
	index, ok := u.objectIdxMap[item.ID]
	if !ok {
		u.objects = append(u.objects, item.CanvasObject)
		index = len(u.objects) - 1
		u.objectIdxMap[item.ID] = index
	} else {
		u.objects[index] = item.CanvasObject
	}
}

func (u *Union) Contains(id any) bool {
	_, ok := u.objectIdxMap[id]
	return ok
}

func (u *Union) SetActive(id any) {
	index, ok := u.objectIdxMap[id]
	if ok {
		u.activeObject = index
	}
	u.Refresh()
}

func (u *Union) CreateRenderer() fyne.WidgetRenderer {
	u.ExtendBaseWidget(u)

	return &unionRenderer{
		union: u,
	}
}

func (u *Union) getActiveObject() (obj fyne.CanvasObject, ok bool) {
	if len(u.objects) == 0 || u.objects[u.activeObject] == nil {
		return nil, false
	}
	return u.objects[u.activeObject], true
}

type UnionItem struct {
	fyne.CanvasObject
	ID any
}

func NewItem(id any, obj fyne.CanvasObject) *UnionItem {
	return &UnionItem{
		CanvasObject: obj,
		ID:           id,
	}
}

type unionRenderer struct {
	union *Union
}

func (r *unionRenderer) Destroy() {
}

func (r *unionRenderer) Layout(size fyne.Size) {
	if obj, ok := r.union.getActiveObject(); ok {
		obj.Resize(size)
	}
}

func (r *unionRenderer) MinSize() fyne.Size {
	if obj, ok := r.union.getActiveObject(); ok {
		return obj.MinSize()
	} else {
		return fyne.NewSize(0, 0)
	}
}

func (r *unionRenderer) Refresh() {
	if obj, ok := r.union.getActiveObject(); ok {
		obj.Refresh()
	}
}

func (r *unionRenderer) Objects() []fyne.CanvasObject {
	if len(r.union.objects) == 0 {
		return nil
	}
	index := r.union.activeObject
	return r.union.objects[index : index+1]
}
