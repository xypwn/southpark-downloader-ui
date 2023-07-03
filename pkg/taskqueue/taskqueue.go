package taskqueue

import (
	"context"
	"sync"
)

type TaskQueue[Data any] struct {
	mtx      sync.Mutex
	waiters  waiters[Data]
	val      int
	size     int
	selectFn func([]Data) int
}

func New[Data any](size int, selectFn func([]Data) int) *TaskQueue[Data] {
	return &TaskQueue[Data]{
		size:     size,
		selectFn: selectFn,
	}
}

func (s *TaskQueue[Data]) SetSize(size int) {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if size > s.size {
		for i := s.size; len(s.waiters) > 0 && i < size; i++ {
			s.makeOneReady()
		}
	}
}

func (s *TaskQueue[Data]) Acquire(ctx context.Context, data Data) error {
	if s.val < s.size {
		s.mtx.Lock()
		s.val++
		s.mtx.Unlock()
		return nil
	}

	s.mtx.Lock()
	waiter := &waiter[Data]{
		data:  data,
		ready: make(chan struct{}),
	}
	s.waiters.insert(waiter)
	s.mtx.Unlock()

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

func (s *TaskQueue[Data]) Release() {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.val--
	for i := 0; i < s.size-s.val; i++ {
		if len(s.waiters) > 0 {
			s.makeOneReady()
		}
	}
}

// Not thread-safe on its own!
func (s *TaskQueue[Data]) makeOneReady() {
	pool := make([]Data, 0, len(s.waiters))
	for _, v := range s.waiters {
		pool = append(pool, v.data)
	}
	i := s.selectFn(pool)
	s.waiters[i].ready <- struct{}{}
	s.waiters.removeIndex(i)
}

type waiter[Data any] struct {
	data  Data
	ready chan struct{}
}

type waiters[Data any] []*waiter[Data]

func (ww waiters[Data]) findLinear(w *waiter[Data]) (int, bool) {
	for i := range ww {
		if ww[i] == w {
			return i, true
		}
	}
	return -1, false
}

func (ww *waiters[Data]) insert(w *waiter[Data]) {
	*ww = append(*ww, w)
}

func (ww *waiters[Data]) remove(w *waiter[Data]) (ok bool) {
	i, ok := ww.findLinear(w)
	if ok {
		ww.removeIndex(i)
	}
	return ok
}

func (ww *waiters[Data]) removeIndex(i int) {
	*ww = append((*ww)[:i], (*ww)[i+1:]...)
}
