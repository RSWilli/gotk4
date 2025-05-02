package gobject

import "github.com/diamondburned/gotk4/pkg/glib/v2"

// #include <glib-object.h>
// extern void _gotk4InterfaceInit(gpointer instance, gpointer ifaceData);
import "C"

func (s *SignalInvocationHint) SignalID() uint {
	return uint(s.native.signal_id)
}

func (s *SignalInvocationHint) Detail() glib.Quark {
	return glib.Quark(s.native.detail)
}

func (s *SignalInvocationHint) RunType() SignalFlags {
	return SignalFlags(s.native.run_type)
}

// SignalAccumulator is a special callback function that can be used to collect return values of the various callbacks that are called during a signal emission.
type SignalAccumulator func(ihint *SignalInvocationHint, return_accu *Value, handler_return *Value) bool

type Signal struct {
	name     string
	signalId C.guint
}

// NewSignal Creates a new signal. (This is usually done in the class initializer.) this is a wrapper around g_signal_newv
func NewSignal(
	name string,
	_type Type,
	flags SignalFlags,
	handler any,
	accumulator SignalAccumulator,
	param_types []Type,
	return_type Type,
) *Signal {
	panic("unimplemented")
}
