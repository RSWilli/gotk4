package gobject

import (
	"unsafe"

	"github.com/diamondburned/gotk4/pkg/core/userdata"
)

// #include <glib-object.h>
import "C"

//export _gotk4_signalAccumulator
func _gotk4_signalAccumulator(
	ihint *C.GSignalInvocationHint,
	return_accu *C.GValue,
	handler_return *C.GValue,
	data C.gpointer,
) C.gboolean {
	goAccuI := userdata.Load(unsafe.Pointer(data))

	goAccu := goAccuI.(SignalAccumulator)

	return gbool(goAccu(
		UnsafeSignalInvocationHintFromGlibBorrow(unsafe.Pointer(ihint)),
		ValueFromNative(unsafe.Pointer(return_accu)),
		ValueFromNative(unsafe.Pointer(handler_return)),
	))
}
