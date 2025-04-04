package gobject

import (
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
	glib.RegisterGValueMarshaler(TypeObject, marshalObject)
}

func marshalObject(p uintptr) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(unsafe.Pointer(p)))
	return Take(unsafe.Pointer(c)), nil
}

type ObjectInstance struct {
	*objectInstance
}

type objectInstance struct {
	native C.GObject
}
