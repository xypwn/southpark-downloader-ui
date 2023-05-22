package prioritysemaphore

import (
	"context"
	"sync"
)

type Semaphore struct {
	mtx sync.Mutex
	waiters waiters
	val     int
	size    int
}

func New(n int) *Semaphore {
	return &Semaphore{
		size: n,
	}
}

func (s *Semaphore) Acquire(ctx context.Context, priority int, waiting func(Handle)) error {
	if s.val < s.size {
		s.mtx.Lock()
		s.val++
		s.mtx.Unlock()
		return nil
	}

	s.mtx.Lock()
	waiter := &waiter{
		priority: priority,
		ready:    make(chan struct{}),
	}
	s.waiters.insert(waiter)
	s.mtx.Unlock()

	waiting(Handle{
		semaphore: s,
		waiter: waiter,
	})

	select {
	case <-ctx.Done():
		s.mtx.Lock()
		s.waiters.remove(waiter)
		s.mtx.Unlock()
		return ctx.Err()
	case <-waiter.ready:
		s.mtx.Lock()
		s.val++
		s.mtx.Unlock()
		return nil
	}
}

func (s *Semaphore) Release() {
	s.mtx.Lock()
	s.val--

	for i := 0; i < s.size-s.val; i++ {
		if len(s.waiters) > 0 {
			s.waiters[0].ready <- struct{}{}
			s.waiters = s.waiters[1:]
		}
	}

	s.mtx.Unlock()
}

type Handle struct {
	semaphore *Semaphore
	waiter *waiter
}

func (h *Handle) GetPriority() int {
	h.semaphore.mtx.Lock()
	priority := h.waiter.priority
	h.semaphore.mtx.Unlock()
	return priority
}

func (h *Handle) SetPriority(priority int) (ok bool) {
	h.semaphore.mtx.Lock()

	// Re-insert with updated priority
	ok = h.semaphore.waiters.remove(h.waiter)
	if ok {
		h.waiter.priority = priority
		h.semaphore.waiters.insert(h.waiter)
	}

	h.semaphore.mtx.Unlock()

	return ok
}

type waiter struct {
	priority int
	ready    chan struct{}
}

// Always sorted by priority
type waiters []*waiter

func (ww waiters) binarySearch(priority int) int {
	// Binary search
	min := 0
	max := len(ww) - 1
	for min <= max {
		mid := (max + min) / 2
		if priority < ww[mid].priority {
			max = mid - 1
		} else if priority > ww[mid].priority {
			min = mid + 1
		} else {
			return mid
		}
	}

	return min
}

func (ww waiters) findLinear(w *waiter) (int, bool) {
	for i := range ww {
		if ww[i] == w {
			return i, true
		}
	}
	return -1, false
}

func (ww *waiters) insert(w *waiter) {
	i := ww.binarySearch(w.priority)
	*ww = append((*ww)[:i],
		append([]*waiter{w}, (*ww)[i:]...)...)
	return
}

func (ww *waiters) remove(w *waiter) (ok bool) {
	i, ok := ww.findLinear(w)
	if ok {
		*ww = append((*ww)[:i], (*ww)[i+1:]...)
	}
	return ok
}
