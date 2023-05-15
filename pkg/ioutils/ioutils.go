package ioutils

import (
	"context"
	"io"
)

// Cancellable reader type
type CtxReader struct {
	io.Reader
	ctx context.Context
}

func NewCtxReader(ctx context.Context, r io.Reader) io.Reader {
	return &CtxReader{
		ctx:    ctx,
		Reader: r,
	}
}

func (r *CtxReader) Read(p []byte) (n int, err error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.Reader.Read(p)
}
