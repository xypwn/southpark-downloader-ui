package asynctask

import (
	"context"
	"sync"
)

type AsyncTask[Param any, Progress any, Result any] struct {
	mtx        sync.Mutex
	parentCtx  context.Context
	ctx        context.Context
	ctxCancel  func()
	asyncFn    func(context.Context, Param, func(Progress)) (Result, error)
	onStart    func()
	onFinish   func(Result, error)
	onProgress func(Progress)
}

func New[Param any, Progress any, Result any](
	ctx context.Context,
	asyncFn func(ctx context.Context, param Param, setProgress func(Progress)) (Result, error), // Run asynchronously when calling Go()
	onStart func(), // Run in the main thread before asyncFn, optional
	onFinish func(Result, error), // Run in the async thread when asyncFn is done or errored
	onProgress func(Progress), // Called when progress changes, optional
) *AsyncTask[Param, Progress, Result] {
	if onStart == nil {
		onStart = func() {}
	}
	if onProgress == nil {
		onProgress = func(Progress) {}
	}
	return &AsyncTask[Param, Progress, Result]{
		parentCtx:  ctx,
		asyncFn:    asyncFn,
		onStart:    onStart,
		onFinish:   onFinish,
		onProgress: onProgress,
	}
}

// ok is false if a task was already running.
// Creates a goroutine and runs the task asynchronously
func (t *AsyncTask[Param, Progress, Result]) Go(param Param) (ok bool) {
	t.mtx.Lock()
	if t.ctxCancel != nil {
		t.mtx.Unlock()
		return false
	}
	t.ctx, t.ctxCancel = context.WithCancel(t.parentCtx)
	t.mtx.Unlock()

	t.onStart()

	go func() {
		res, err := t.asyncFn(t.ctx, param, t.onProgress)

		t.mtx.Lock()
		t.ctx, t.ctxCancel = nil, nil
		t.mtx.Unlock()

		t.onFinish(res, err)
	}()

	return true
}

// ok is false if no task was running.
// Can also be canceled using the context given upon creation,
// but in that case Go() can't be run again.
func (t *AsyncTask[Param, Progress, Result]) Cancel() (ok bool) {
	t.mtx.Lock()
	cancel := t.ctxCancel
	t.mtx.Unlock()

	if cancel == nil {
		return false
	}

	cancel()

	return true
}
