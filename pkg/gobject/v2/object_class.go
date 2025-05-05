package gobject

import "unsafe"

// #cgo pkg-config: gobject-2.0
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <glib-object.h>
import "C"

type ObjectClass struct {
	*objectClass
}

// objectClass is the struct that is finalized
type objectClass struct {
	native *C.GObjectClass
}

func UnsafeObjectClassFromGlibBorrow(p unsafe.Pointer) *ObjectClass {
	return &ObjectClass{&objectClass{(*C.GObjectClass)(p)}}
}

// UnsafeAddPrivateData registers a private structure of the given size for the class
//
// this should not be called by user code, but only by the generated bindings
func (o *ObjectClass) UnsafeAddPrivateData(size uintptr) {
	// https://docs.gtk.org/gobject/method.TypeClass.add_private.html

	// FIXME: this is deprecated, but the alternative is a macro?
	C.g_type_class_add_private(C.gpointer(o.native), C.gsize(size))
}

func RegisterObjectSubClass[InstanceT Object](
	name string,
	classInit func(class *ObjectClass),
	constructor func() InstanceT,
	overrides ObjectOverrides[InstanceT],
	signals map[string]SignalDefinition,
	interfaceInits ...SubClassInterfaceInit[InstanceT],
) Type {
	return UnsafeRegisterSubClass(
		name,
		classInit,
		constructor,
		overrides,
		signals,
		TypeObject,
		UnsafeObjectClassFromGlibBorrow,
		UnsafeApplyObjectOverrides,
		func(obj *ObjectInstance) Object {
			return obj
		},
		interfaceInits...,
	)
}
