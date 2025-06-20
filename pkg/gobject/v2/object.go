package gobject

import (
	"runtime"
	"unsafe"
)

// #cgo pkg-config: gobject-2.0
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <glib-object.h>
// static GObjectClass *_g_object_get_class(GObject *object) {
//   return (G_OBJECT_GET_CLASS(object));
// }
// static GType _g_type_from_instance(gpointer instance) {
//   return (G_TYPE_FROM_INSTANCE(instance));
// }
import "C"

// The base object type.
//
// This is an interface because the actual type will almost always extend
type Object interface {
	GoValueInitializer

	Emit(detailedSignal string, args ...any) any

	Connect(detailedSignal string, f interface{}) SignalHandle
	ConnectAfter(detailedSignal string, f interface{}) SignalHandle

	HandlerBlock(SignalHandle)
	HandlerUnblock(SignalHandle)
	HandlerDisconnect(SignalHandle)

	NotifyProperty(string, func()) SignalHandle
	ObjectProperty(string) interface{}
	SetObjectProperty(string, interface{})
	FreezeNotify()
	ThawNotify()
	StopEmission(string)

	// Parent virtual methods:

	ParentConstructed()
	ParentFinalize()

	// internal methods:

	isFloating() bool

	unsafeForceFloating()

	baseObject() *ObjectInstance
}

var _ Object = (*ObjectInstance)(nil)

const (
	TypeObject Type = C.G_TYPE_OBJECT
)

func init() {
	RegisterGValueMarshaler(TypeObject, marshalObject)

	RegisterObjectCasting(TypeObject, func(inst *ObjectInstance) Object {
		// this is the base type, so we can just return the instance
		return inst
	})
}

// marshalObject returns a concrete ObjectInstance, because this is only called when we do not know the actual
// extending type
func marshalObject(p unsafe.Pointer) (interface{}, error) {
	c := C.g_value_get_object((*C.GValue)(p))
	return newObject(unsafe.Pointer(c), true), nil
}

// UnsafeObjectFromGlibNone is used to convert raw C object pointers to go while taking a reference.
// the returned Object is casted correctly and needs a manual cast by the user to the correct extending interface
//
// This is used by the bindings internally.
func UnsafeObjectFromGlibNone(p unsafe.Pointer) Object {
	obj := newObject(p, true)

	return obj.cast()
}

// UnsafeObjectFromGlibBorrow is used to convert raw C object pointers to go without taking a reference or touching the
// floating reference. The returned Object is casted correctly and needs a manual cast by the user to the correct extending interface
//
// This is used by the bindings internally.
func UnsafeObjectFromGlibBorrow(p unsafe.Pointer) Object {
	obj := wrapObject(p)

	return obj.cast()
}

// UnsafeObjectFromGlibFull is used to convert raw C object pointers to go.
// the returned Object is casted correctly and needs a manual cast by the user to the correct extending interface
func UnsafeObjectFromGlibFull(p unsafe.Pointer) Object {
	obj := newObject(p, false)

	return obj.cast()
}

// UnsafeObjectToGlibNone is used to convert the Object to C.
func UnsafeObjectToGlibNone(obj Object) unsafe.Pointer {
	base := obj.baseObject()

	return base.unsafe()
}

// UnsafeObjectToGlibFull is used to convert the Object to C while removing the cleanup.
func UnsafeObjectToGlibFull(obj Object) unsafe.Pointer {
	base := obj.baseObject()

	runtime.SetFinalizer(base.objectInstance, nil)

	return base.unsafe()
}

func wrapObject(p unsafe.Pointer) *ObjectInstance {
	return &ObjectInstance{
		objectInstance: &objectInstance{
			native: (*C.GObject)(p),
		},
	}
}

// newObject returns a new object instance and attaches a cleanup to unref the c pointer on GC. if ref is true
// then a reference will be taken on the object.
func newObject(p unsafe.Pointer, ref bool) *ObjectInstance {
	obj := wrapObject(p)

	if ref {
		// if the object was floating this removes the floating ref.
		// if not, then this is equivalent to g_object_ref
		C.g_object_ref_sink(C.gpointer(obj.unsafe()))
	}

	runtime.SetFinalizer(
		obj.objectInstance,
		func(intern *objectInstance) {
			C.g_object_unref(C.gpointer(intern.native))
		},
	)

	return obj
}

type ObjectInstance struct {
	*objectInstance
}

// objectInstance is the object that is finalized
type objectInstance struct {
	native *C.GObject
}

// unsafeForceFloating implements Object.
func (v *ObjectInstance) unsafeForceFloating() {
	C.g_object_force_floating(v.native)
	runtime.KeepAlive(v)
}

// isFloating implements Object.
func (v *ObjectInstance) isFloating() bool {
	return C.g_object_is_floating(C.gpointer(v.unsafe())) != 0
}

// GoValueType implements GoValueInitializer.
func (obj *ObjectInstance) GoValueType() Type {
	// always use the type from the object instance, instead of the base type,
	// since this is inherited by all extending types
	return obj.typeFromInstance()
}

// SetGoValue implements GoValueInitializer.
func (obj *ObjectInstance) SetGoValue(v *Value) {
	v.SetObject(obj)
}

func (v *ObjectInstance) unsafe() unsafe.Pointer {
	if v == nil {
		return nil
	}
	if v.objectInstance == nil {
		panic("this Object is invalid")
	}

	return unsafe.Pointer(v.native)
}

func (v *ObjectInstance) baseObject() *ObjectInstance {
	return v
}

// ObjectProperty is a wrapper around g_object_get_property(). If the property's
// type cannot be resolved to a Go type, then InvalidValue is returned.
func (v *ObjectInstance) ObjectProperty(name string) interface{} {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))

	t := v.propertyType((*C.gchar)(cstr))
	if t == TypeInvalid {
		return InvalidValue
	}

	p := InitValue(t)

	C.g_object_get_property(v.native, (*C.gchar)(cstr), p.native())
	runtime.KeepAlive(v)

	return p.GoValue()
}

// SetObjectProperty is a wrapper around g_object_set_property().
func (v *ObjectInstance) SetObjectProperty(name string, value interface{}) {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))

	p := allocateValue()
	p.InitGoValue(value)
	p.SetGoValue(value)
	defer p.unset()

	C.g_object_set_property(v.native, (*C.gchar)(cstr), p.native())
	runtime.KeepAlive(v)
}

// NotifyProperty adds a handler that's called when the object's property is
// updated.
func (v *ObjectInstance) NotifyProperty(property string, f func()) SignalHandle {
	return v.Connect("notify::"+property, f)
}

// FreezeNotify increases the freeze count on object. If the freeze count is
// non-zero, the emission of “notify” signals on object is stopped. The signals
// are queued until the freeze count is decreased to zero. Duplicate
// notifications are squashed so that at most one GObject::notify signal is
// emitted for each property modified while the object is frozen.
//
// This is necessary for accessors that modify multiple properties to prevent
// premature notification while the object is still being modified.
func (v *ObjectInstance) FreezeNotify() {
	C.g_object_freeze_notify(v.native)
	runtime.KeepAlive(v)
}

// ThawNotify reverts the effect of a previous call to g_object_freeze_notify().
// The freeze count is decreased on object and when it reaches zero, queued
// “notify” signals are emitted.
//
// Duplicate notifications for each property are squashed so that at most one
// GObject::notify signal is emitted for each property, in the reverse order in
// which they have been queued.
//
// It is an error to call this function when the freeze count is zero.
func (v *ObjectInstance) ThawNotify() {
	C.g_object_thaw_notify(v.native)
	runtime.KeepAlive(v)
}

// HandlerBlock is a wrapper around g_signal_handler_block().
func (v *ObjectInstance) HandlerBlock(handle SignalHandle) {
	C.g_signal_handler_block(C.gpointer(v.unsafe()), C.gulong(handle))
	runtime.KeepAlive(v)
}

// HandlerUnblock is a wrapper around g_signal_handler_unblock().
func (v *ObjectInstance) HandlerUnblock(handle SignalHandle) {
	C.g_signal_handler_unblock(C.gpointer(v.unsafe()), C.gulong(handle))
	runtime.KeepAlive(v)
}

// HandlerDisconnect is a wrapper around g_signal_handler_disconnect().
func (v *ObjectInstance) HandlerDisconnect(handle SignalHandle) {
	C.g_signal_handler_disconnect(C.gpointer(v.unsafe()), C.gulong(handle))
	runtime.KeepAlive(v)
}

// StopEmission stops a signal’s current emission. It is a wrapper around
// g_signal_stop_emission_by_name().
func (v *ObjectInstance) StopEmission(s string) {
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr))

	C.g_signal_stop_emission_by_name(C.gpointer(v.unsafe()), (*C.gchar)(cstr))
	runtime.KeepAlive(v)
}

// PropertyType returns the Type of a property of the underlying GObject. If the
// property is missing, it will return TypeInvalid.
func (v *ObjectInstance) PropertyType(name string) Type {
	cstr := C.CString(name)
	defer C.free(unsafe.Pointer(cstr))

	return v.propertyType(cstr)
}

func (v *ObjectInstance) propertyType(cstr *C.gchar) Type {
	paramSpec := C.g_object_class_find_property(C._g_object_get_class(v.native), (*C.gchar)(cstr))
	runtime.KeepAlive(v)

	if paramSpec == nil {
		return TypeInvalid
	}

	return Type(paramSpec.value_type)
}

func (v *ObjectInstance) unsafePrivateData() unsafe.Pointer {
	private := C.g_type_instance_get_private((*C.GTypeInstance)(v.unsafe()), C.GType(v.typeFromInstance()))

	return unsafe.Pointer(private)
}

// TypeFromInstance is a wrapper around g_type_from_instance().
func (v *ObjectInstance) typeFromInstance() Type {
	return unsafeTypeFromObject(v.unsafe())
}

func unsafeTypeFromObject(instance unsafe.Pointer) Type {
	c := C._g_type_from_instance(C.gpointer(instance))
	return Type(c)
}

// cast casts v to the concrete Go type (e.g. *Object to *gtk.Entry).
func (v *ObjectInstance) cast() Object {
	if v.unsafe() == nil {
		// nil-typed interface != non-nil-typed nil-value interface
		return nil
	}

	// re-implement the gvalue marshaling here that takes the type from the instance
	// and walks up the inheritance chain to find the correct casting function
	//
	// we MUST NOT use gvalue here, because that would take a reference on the object,
	// and that is something the caller should decide
	//
	// we also can't ref and unref, because a floating reference would be cleaned up
	//
	// we KISS here: just don't touch the reference at all!

	typeFromInstance := v.typeFromInstance()

	for {
		objectCastingsLock.RLock()
		castFunc, exists := objectCastings[typeFromInstance]
		objectCastingsLock.RUnlock()

		if exists {
			return castFunc(v)
		}

		if typeFromInstance == TypeObject {
			// panic here to never block or return nil, Object should always be handled
			panic("type object must have a casting function registered")
		}

		typeFromInstance = typeFromInstance.Parent()
	}
}
