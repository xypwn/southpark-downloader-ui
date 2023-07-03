package data

import (
	"sync"
)

// A Client can use the Get and Change methods,
// and set up a listener using SetListenerFn. When
// the value is changed by a client, all other
// Clients EXCEPT the one which made the change
// are notified via their ListenerFn.
type Client[T any] struct {
	mtx         sync.RWMutex
	parent      *Binding[T]
	listenerFns []func(T)
}

// Gets called whenever a DIFFERENT CLIENT calls
// the Change method.
func (c *Client[T]) AddListener(fn func(T)) {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	c.listenerFns = append(c.listenerFns, fn)
}

// Allows you to look at the array and extract some data.
// Do NOT exfiltrate any references to data.
// Do NOT modify the data.
// These restrictions apply to ensure thread safety.
func (c *Client[T]) Examine(examiner func(T)) {
	c.parent.mtx.RLock()
	defer c.parent.mtx.RUnlock()
	examiner(c.parent.data)
}

// Changes the value and notifies all OTHER Clients
// by calling their ListenerFn function
func (c *Client[T]) Change(changer func(T) T) {
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
type Binding[T any] struct {
	mtx     sync.RWMutex
	data    T
	clients []*Client[T]
}

func NewBinding[T any]() *Binding[T] {
	res := &Binding[T]{}
	res.NewClient() // default client
	return res
}

// Creates a new client used to access and manipulate the data.
// See Client.
func (b *Binding[T]) NewClient() *Client[T] {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	c := &Client[T]{
		parent: b,
	}
	b.clients = append(b.clients, c)
	return c
}

// Removes the given client, if it existed in the binding.
// Returns whether removal was successful.
func (b *Binding[T]) RemoveClient(client *Client[T]) bool {
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
// exclude may be nil.
func (b *Binding[T]) notifyClients(exclude *Client[T]) {
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
