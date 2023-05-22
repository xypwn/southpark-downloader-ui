package data

import (
	"sort"
	"sync"
)

type listListener[T any] struct {
	Fn func([]T)
	ID uint
}

type ListBinding[T any] struct {
	mtx            sync.Mutex
	data           []T
	listeners      []listListener[T]
	nextListenerID uint
}

func NewListBinding[T any]() *ListBinding[T] {
	return &ListBinding[T]{}
}

func (b *ListBinding[T]) Get() []T {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	return b.data
}

func (b *ListBinding[T]) Set(value []T) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	b.data = value
	b.notifyListeners()
}

func (b *ListBinding[T]) Change(changer func([]T) []T) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	b.data = changer(b.data)
	b.notifyListeners()
}

func (b *ListBinding[T]) Len() int {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	return len(b.data)
}

func (b *ListBinding[T]) Append(value T) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	b.data = append(b.data, value)
	b.notifyListeners()
}

func (b *ListBinding[T]) Prepend(value T) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	b.data = append([]T{value}, b.data...)
	b.notifyListeners()
}

func (b *ListBinding[T]) Sort(comparator func(T, T) bool) {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	dataComparator := func(i, j int) bool {
		return comparator(b.data[i], b.data[j])
	}

	sort.Slice(b.data, dataComparator)
	b.notifyListeners()
}

func (b *ListBinding[T]) AddListener(fn func([]T)) (id uint) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	id = b.nextListenerID
	b.nextListenerID++
	b.listeners = append(b.listeners, listListener[T]{
		Fn: fn,
		ID: id,
	})
	fn(b.data)
	return
}

func (b *ListBinding[T]) RemoveListener(id uint) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	for i, v := range b.listeners {
		if v.ID == id {
			b.listeners = append(b.listeners[:i], b.listeners[i+1:]...)
			break
		}
	}
}

func (b *ListBinding[T]) notifyListeners() {
	for _, v := range b.listeners {
		v.Fn(b.data)
	}
}
