package closer

import (
	"context"
	"slices"
)

type Closer struct {
	funcs []func()
}

func New() *Closer {
	return &Closer{
		funcs: make([]func(), 0),
	}
}

func (c *Closer) Add(fn func()) {
	if fn == nil {
		return
	}
	c.funcs = append(c.funcs, fn)
}

func (c *Closer) CloseAll(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		defer close(done)

		for _, fn := range slices.Backward(c.funcs) {
			fn()
		}
	}()

	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case <-done:
		return nil
	}
}
