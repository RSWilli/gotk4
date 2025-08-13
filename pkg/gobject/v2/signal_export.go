package gobject

import (
	"unsafe"

	gopointer "github.com/go-gst/go-pointer"
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
	goAccuI := gopointer.Restore(unsafe.Pointer(data))

	goAccu := goAccuI.(SignalAccumulator)

	return gbool(goAccu(
		UnsafeSignalInvocationHintFromGlibBorrow(unsafe.Pointer(ihint)),
		ValueFromNative(unsafe.Pointer(return_accu)),
		ValueFromNative(unsafe.Pointer(handler_return)),
	))
}
