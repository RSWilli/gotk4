package gobject

import (
	"runtime"
	"unsafe"

	"github.com/diamondburned/gotk4/pkg/glib/v2"
)

// #cgo pkg-config: gobject-2.0
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <glib-object.h>
import "C"

const (
	TypeObject glib.Type = C.G_TYPE_OBJECT
)

func init() {
	glib.RegisterGValueMarshaler(TypeObject, marshalvalue)
}

func marshalValue(p uintptr) (interface{}, error) {
	c := C.g_value_get_boxed((*C.GValue)(unsafe.Pointer(p)))
	if c == nil {
		return nil, nil
	}

	v := &value{(*C.GValue)(unsafe.Pointer(c))}
	runtime.SetFinalizer(v, (*value).unset)

	return &Value{v}, nil
}
