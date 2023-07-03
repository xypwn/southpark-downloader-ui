package data

import (
	"sync"
)

// A Client can use the Examine and Change methods,
// and set up a listener using SetListenerFn. When
// the value is changed by a client, all other
// Clients EXCEPT the one which made the change
// are notified via their ListenerFn.
type ListClient[T any] struct {
	mtx         sync.RWMutex
	parent      *ListBinding[T]
	listenerFns []func([]T)
}

// Gets called whenever a DIFFERENT CLIENT calls
// the Change method.
func (c *ListClient[T]) AddListener(fn func([]T)) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.listenerFns = append(c.listenerFns, fn)
}

// Allows you to look at the array and extract some data (e.g. length or an item).
// Do NOT exfiltrate the list itself (you may create a deep copy).
// Do NOT modify the data.
// These restrictions apply to ensure thread safety.
func (c *ListClient[T]) Examine(examiner func([]T)) {
	c.parent.mtx.RLock()
	defer c.parent.mtx.RUnlock()
	examiner(c.parent.data)
}

// Changes the value and notifies all OTHER Clients
// by calling their ListenerFn function.
func (c *ListClient[T]) Change(changer func([]T) []T) {
	c.parent.mtx.Lock()
	defer c.parent.mtx.Unlock()
	c.parent.data = changer(c.parent.data)
	c.parent.notifyClients(c)
}

// A binding represents a piece of data which
// needs to be synced between different pieces
// of code (Clients). Clients may run on separate
// threads safely.
// See Client.
type ListBinding[T any] struct {
	mtx     sync.RWMutex
	data    []T
	clients []*ListClient[T]
}

func NewListBinding[T any]() *ListBinding[T] {
	return &ListBinding[T]{}
}

// Creates a new client used to access and manipulate the data.
// See Client.
func (b *ListBinding[T]) NewClient() *ListClient[T] {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	c := &ListClient[T]{
		parent: b,
	}
	b.clients = append(b.clients, c)
	return c
}

// Removes the given client, if it existed in the binding.
// Returns whether removal was successful.
func (b *ListBinding[T]) RemoveClient(client *ListClient[T]) bool {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	for i, v := range b.clients {
		if v == client {
			b.clients = append(b.clients[:i], b.clients[i+1:]...)
			return true
		}
	}
	return false
}

// NOT inherently thread-safe! Expects the CALLER to LOCK b.mtx!
// See Change.
func (b *ListBinding[T]) notifyClients(exclude *ListClient[T]) {
	for _, v := range b.clients {
		if v == exclude {
			continue
		}

		v.mtx.RLock()
		for _, fn := range v.listenerFns {
			if fn != nil {
				fn(b.data)
			}
		}
		v.mtx.RUnlock()
	}
}
