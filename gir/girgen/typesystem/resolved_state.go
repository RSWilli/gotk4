package typesystem

import (
	stdcontext "context"
	"sync/atomic"
)

type resolvedState int

const (
	notResolvable resolvedState = iota
	okResolved
)

type lazyResolved struct {
	v      atomic.Pointer[resolvedState]
	notify chan struct{}

	resolveFn func(ctx stdcontext.Context) resolvedState
}

func newLazyResolved(resolveFn func(ctx stdcontext.Context) resolvedState) *lazyResolved {
	return &lazyResolved{
		notify: make(chan struct{}),
	}
}

// awaitResolved implements resolvedType.
func (l *lazyResolved) awaitResolved(ctx stdcontext.Context) resolvedState {
	v := l.v.Load()
	if v == nil {
		select {
		case <-l.notify:
			return l.awaitResolved(ctx)
		case <-ctx.Done():
			return notResolvable
		}
	}

	return *v
}

func (l *lazyResolved) resolve(ctx stdcontext.Context) {
	v := l.resolveFn(ctx)

	if !l.v.CompareAndSwap(nil, &v) {
		panic("could not swap lazy resolved from nil, did you call resolve concurrently")
	}

	close(l.notify)
}
