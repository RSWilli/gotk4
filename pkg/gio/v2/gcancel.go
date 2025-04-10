// Package gcancel provides a converter between gio.Cancellable and
// context.Context.
package gio

// #cgo pkg-config: gio-2.0
// #include <gio/gio.h>
import "C"

import (
	"context"
	"time"
	"unsafe"

	"github.com/diamondburned/gotk4/pkg/gobject/v2"
)

type ctxKey uint8

const (
	_ ctxKey = iota
	cancellableKey
)

// Cancellable is a wrapper around the GCancellable object. It satisfies the
// context.Context interface.
type Cancellable struct {
	*gobject.ObjectInstance
	ctx  context.Context
	done <-chan struct{}
}

var _ context.Context = (*Cancellable)(nil)

// Cancel will set cancellable to cancelled. It is the same as calling the
// cancel callback given after context creation.
func (c *Cancellable) Cancel() {
	// Save a Cgo call: if the channel is already closed, then ignore.
	select {
	case <-c.done:
		return
	default:
	}

	if c.ObjectInstance == nil {
		panic("bug: Cancel called on nil Cancellable object")
	}

	native := (*C.GCancellable)(unsafe.Pointer(gobject.UnsafeObjectToGlibNone(c)))
	C.g_cancellable_cancel(native)
}

// IsCancelled checks if a cancellable job has been cancelled.
func (c *Cancellable) IsCancelled() bool {
	// Fast paths: check the contexts, which will be closed by our goroutines.

	select {
	case <-c.done:
		return true
	default:
	}

	select {
	case <-c.ctx.Done():
		return true
	default:
	}

	// nil obj == no cancellable.
	if c.ObjectInstance == nil {
		return false
	}

	native := (*C.GCancellable)(unsafe.Pointer(gobject.UnsafeObjectToGlibNone(c)))
	return C.g_cancellable_is_cancelled(native) != 0
}

// Deadline returns the deadline of the parent context.
func (c *Cancellable) Deadline() (time.Time, bool) {
	return c.ctx.Deadline()
}

// Value returns the values of the parent context.
func (c *Cancellable) Value(key interface{}) interface{} {
	return c.ctx.Value(key)
}

// Done returns the channel that's closed once the cancellable is cancelled.
func (c *Cancellable) Done() <-chan struct{} {
	return c.done
}

// Err returns context.Canceled if the cancellable is already cancelled,
// otherwise nil is returned.
func (c *Cancellable) Err() error {
	if c.IsCancelled() {
		return context.Canceled
	}
	return nil
}

var nilCancellable = &Cancellable{
	ObjectInstance: nil,
	ctx:            context.Background(),
	done:           nil,
}

func UnsafeGCancellableToGlibNone(ctx context.Context) unsafe.Pointer {
	cancel := GCancellableFromContext(ctx)

	if cancel == nil {
		return nil
	}

	// TODO: cancel borrows from context, link the lifetimes together.

	return gobject.UnsafeObjectToGlibNone(cancel)
}

// CancellableFromContext gets the underlying Cancellable instance from the
// given context. If ctx does not contain the Cancellable instance, then a
// context with a nil Object field is returned. It is mostly for internal use;
// users should use WithCancel instead.
func GCancellableFromContext(ctx context.Context) *Cancellable {
	if obj := fromContext(ctx, false); obj != nil {
		return obj
	}
	return nilCancellable
}

// NewCancellableContext creates a new context.Context from the given
// *GCancellable. If the pointer is nil, then context.Background() is used.
func NewCancellableContext(cancellable unsafe.Pointer) context.Context {
	cval := (*C.GCancellable)(cancellable)
	if cval == nil {
		return context.Background()
	}

	obj := &Cancellable{
		ObjectInstance: gobject.UnsafeObjectFromGlibFull(unsafe.Pointer(cancellable)).(*gobject.ObjectInstance),
		ctx:            context.Background(),
	}

	done := make(chan struct{})
	obj.Connect("cancelled", func() { close(done) })
	obj.done = done

	return obj
}

// WithCancel behaves similarly to context.WithCancel, except the created
// context is of type Cancellable. This is useful if the user wants to reuse the
// same Cancellable instance for multiple calls.
//
// This function costs a goroutine to do this unless the given context is
// previously created with WithCancel, is otherwise a Cancellable instance, or
// is an instance from context.Background() or context.TODO(), but it should be
// fairly cheap otherwise.
func WithCancel(ctx context.Context) (context.Context, context.CancelFunc) {
	cancellable := fromContext(ctx, true)
	return context.WithValue(cancellable, cancellableKey, cancellable), cancellable.Cancel
}

func fromContext(ctx context.Context, create bool) *Cancellable {
	if ctx == nil {
		panic("given ctx is nil")
	}

	// If the context is already a cancellable, then return that.
	if v, ok := ctx.(*Cancellable); ok {
		return v
	}

	// If the context inherits a cancellable somewhere, then use it, but only if
	// the Done channel is still the same as the context's. We don't want to
	// mistakenly use the wrong channel.
	v, ok := ctx.Value(cancellableKey).(*Cancellable)
	if ok && ctx.Done() == v.done {
		return v
	}

	if !create { // TODO: why do we not cancel on parent here?
		return nil
	}

	cancellable := &Cancellable{
		ObjectInstance: gobject.UnsafeObjectFromGlibFull(unsafe.Pointer(C.g_cancellable_new())).(*gobject.ObjectInstance),
		ctx:            ctx,
	}

	done := make(chan struct{})
	cancellable.Connect("cancelled", func(_ gobject.Object) { close(done) })
	cancellable.done = done

	// Only need this if the parent context isn't Background.
	if ctx != context.Background() && ctx != context.TODO() {
		go cancelOnParent(ctx, done, cancellable)
	}

	return cancellable
}

func cancelOnParent(ctx context.Context, done chan struct{}, cancellable *Cancellable) {
	select {
	case <-ctx.Done():
		cancellable.Cancel()
	case <-cancellable.done:
	}
}
