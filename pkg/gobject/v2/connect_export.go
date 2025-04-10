package gobject

import (
	"fmt"
	"os"
	"reflect"
	"unsafe"

	"github.com/diamondburned/gotk4/pkg/core/closure"
)

// #include <glib.h>
// #include <glib-object.h>
import "C"

// _gotk4_goMarshal is called by the GLib runtime when a closure needs to be invoked.
// The closure will be invoked with as many arguments as it can take, from 0 to
// the full amount provided by the call. If the closure asks for more parameters
// than there are to give, a warning is printed to stderr and the closure is
// not run.
//
//export _gotk4_goMarshal
func _gotk4_goMarshal(
	gclosure *C.GClosure,
	retValue *C.GValue,
	nParams C.guint,
	params *C.GValue,
	invocationHint C.gpointer,
	marshalData C.gpointer,
) {
	fs := closure.Load(unsafe.Pointer(gclosure))
	defer fs.TryRepanic()

	nTotalParams := int(nParams)

	fValue := fs.Value()
	fType := fValue.Type()

	// nCbParams is the number of parameters in the callback function.
	nCbParams := fType.NumIn()

	if nCbParams != nTotalParams {
		panic(fmt.Sprintf("callback function has %d parameters, but %d were provided", nCbParams, nTotalParams))
	}

	// Create a slice of reflect.Values as arguments to call the function.
	gValues := unsafe.Slice(params, nCbParams)
	args := make([]reflect.Value, 0, nCbParams)

	// Fill beginning of args, up to the minimum of the total number of callback
	// parameters and parameters from the glib runtime.
	for i := range nCbParams {
		v := ValueFromNative(unsafe.Pointer(&gValues[i]))
		val := v.GoValue()

		if val == InvalidValue {
			fmt.Fprintf(os.Stderr,
				"invalid value for arg %d: %v\n", i, v)
			return
		}

		rv := reflect.ValueOf(val)
		args = append(args, rv.Convert(fType.In(i)))
	}

	// Call closure with args. If the callback returns one or more
	// values, save the GValue equivalent of the first.
	rv := fValue.Call(args)
	if retValue != nil && len(rv) > 0 {
		g := NewValue(rv[0].Interface())

		C.g_value_copy(g.native(), retValue)

	}
}

//export _gotk4_removeClosure
func _gotk4_removeClosure(obj *C.GObject, gclosure *C.GClosure) {
	closure.Delete(unsafe.Pointer(gclosure))
}
