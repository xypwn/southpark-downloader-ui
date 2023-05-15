package prioritysemaphore

import (
	"context"
	"sync"
)

type Semaphore struct {
	sync.Mutex
	waiters waiters
	val     int
	size    int
}

func New(n int) *Semaphore {
	return &Semaphore{
		size: n,
	}
}

func (s *Semaphore) Acquire(ctx context.Context, priority int) error {
	if s.val < s.size {
		s.Lock()
		s.val++
		s.Unlock()
		return nil
	}

	s.Lock()
	waiter := waiter{
		priority: priority,
		ready:    make(chan struct{}),
	}
	s.waiters.insert(waiter)
	s.Unlock()

	select {
	case <-ctx.Done():
		s.Lock()
		s.waiters.remove(waiter)
		s.Unlock()
		return ctx.Err()
	case <-waiter.ready:
		s.Lock()
		s.val++
		s.Unlock()
		return nil
	}
}

func (s *Semaphore) Release() {
	s.Lock()
	s.val--

	for i := 0; i < s.size-s.val; i++ {
		if len(s.waiters) > 0 {
			s.waiters[0].ready <- struct{}{}
			s.waiters = s.waiters[1:]
		}
	}

	s.Unlock()
}

type waiter struct {
	priority int
	ready    chan struct{}
}

// Always sorted by priority
type waiters []waiter

func (ww waiters) find(w waiter) int {
	// Binary search
	min := 0
	max := len(ww) - 1
	for min <= max {
		mid := (max + min) / 2
		if w.priority < ww[mid].priority {
			max = mid - 1
		} else if w.priority > ww[mid].priority {
			min = mid + 1
		} else {
			return mid
		}
	}

	return min
}

func (ww *waiters) insert(w waiter) {
	i := ww.find(w)
	*ww = append((*ww)[:i],
		append([]waiter{w}, (*ww)[i:]...)...)
	return
}

func (ww *waiters) remove(w waiter) {
	i := ww.find(w)
	*ww = append((*ww)[:i], (*ww)[i+1:]...)
	return
}
