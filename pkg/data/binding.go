package data

import (
	"sync"
)

type listener[T any] struct {
	Fn func(T)
	ID uint
}

type Binding[T any] struct {
	mtx            sync.Mutex
	data           T
	listeners      []listener[T]
	nextListenerID uint
}

func NewBinding[T any]() *Binding[T] {
	return &Binding[T]{}
}

func (b *Binding[T]) Get() T {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	return b.data
}

func (b *Binding[T]) Set(value T) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	b.data = value
	b.notifyListeners()
}

func (b *Binding[T]) Change(changer func(T) T) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	b.data = changer(b.data)
	b.notifyListeners()
}

func (b *Binding[T]) AddListener(fn func(T)) (id uint) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	id = b.nextListenerID
	b.nextListenerID++
	b.listeners = append(b.listeners, listener[T]{
		Fn: fn,
		ID: id,
	})
	fn(b.data)
	return
}

func (b *Binding[T]) RemoveListener(id uint) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	for i, v := range b.listeners {
		if v.ID == id {
			b.listeners = append(b.listeners[:i], b.listeners[i+1:]...)
			break
		}
	}
}

func (b *Binding[T]) notifyListeners() {
	for _, v := range b.listeners {
		v.Fn(b.data)
	}
}
