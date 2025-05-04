package gobject

// #include <glib-object.h>
// extern void _gotk4InterfaceInit(gpointer instance, gpointer ifaceData);
// extern void _gotk4ClassInit(gpointer gclass, gpointer classData);
// extern void _gotk4InstanceInit(GTypeInstance* instance, gpointer g_class);
// GType typeFromGObjectClass (GObjectClass *c) { return (G_OBJECT_CLASS_TYPE(c)); };
import "C"
import (
	"log"
	"reflect"
	"unsafe"

	"github.com/diamondburned/gotk4/pkg/core/userdata"
)

type ObjectOverrider[Instance Object] interface {
	// getObjectOverrides retrieves the object overrides from any extending overrider
	getObjectOverrides() ObjectOverrides[Instance]
}

// ObjectOverrides is the struct used to override the default implementation of virtual methods.
// it is generic over the extending instance type.
type ObjectOverrides[Instance Object] struct {
	// The constructed function is called by g_object_new() as the final step of the object creation process.
	// At the point of the call, all construction properties have been set on the object. The purpose of this
	// call is to allow for object initialisation steps that can only be performed after construction properties
	// have been set. constructed implementors should chain up to the constructed call of their parent class to
	// allow it to complete its initialisation.
	Constructed func(Instance)
	// The dispose function is supposed to drop all references to other objects, but keep the instance otherwise intact,
	// so that client method invocations still work. It may be run multiple times (due to reference loops). Before returning,
	// dispose should chain up to the dispose method of the parent class.
	Dispose func(Instance)
	// Instance finalization function, should finish the finalization of the instance begun in dispose and chain up to the
	// finalize method of the parent class.
	//
	// This is additionally wrapped by the bindings to clean up the instance data.
	Finalize func(Instance)
}

func (o ObjectOverrides[Instance]) getObjectOverrides() ObjectOverrides[Instance] {
	return o
}

// UnsafeApplyObjectOverrides applies the overrides to init the gclass by setting the trampoline functions.
// This is used by the bindings internally and only exported for visibility to other bindings code.
func UnsafeApplyObjectOverrides[Instance Object](gclass unsafe.Pointer, overrides ObjectOverrides[Instance]) {

}

type subClassData[InstanceT Object] struct {
	classInit    func(gclass unsafe.Pointer)
	instanceInit func(instance InstanceT)

	// newFromGlib is the function to create a new instance from the glib pointer in the instance init function.
	newFromGlib func(unsafe.Pointer) InstanceT
}

// UnsafeRegisterSubClass registers a new subclass of the given parentGtype. This is wrapped by the generated bindings
// for ease of use.
// This aligns with https://gjs.guide/guides/gobject/subclassing.html
func UnsafeRegisterSubClass[InstanceT Object, ClassT, OverridesT ObjectOverrider[InstanceT]](
	// user supplied arguments:
	classInit func(class ClassT),
	constructor func() InstanceT,
	overrides OverridesT,
	signals map[string]SignalDefinition,
	// parent class depending arguments, these can be autofilled by the generated bindings:
	parentGtype Type,
	parentClassFromUnsafePointer func(unsafe.Pointer) ClassT,
	parentApplyOverridesFunc func(gclass unsafe.Pointer, overrides OverridesT),
	parentWrapObject func(*ObjectInstance) Object,

	// user supplied interfaces:
	interfaceInits ...SubClassInterfaceInit[InstanceT],
) Type {
	if classInit == nil {
		classInit = func(class ClassT) {}
	}

	var typeQuery C.GTypeQuery
	C.g_type_query(C.GType(parentGtype), &typeQuery)
	if typeQuery._type == 0 {
		log.Panicln("GType", parentGtype, "is is unknown")
	}

	data := &subClassData[InstanceT]{
		classInit: func(gclass unsafe.Pointer) {
			// first override the virtual methods on the class
			parentApplyOverridesFunc(gclass, overrides)

			class := parentClassFromUnsafePointer(gclass)

			// gtype is the type of the instance
			gtype := Type(C.typeFromGObjectClass((*C.GObjectClass)(gclass)))

			// register the signals
			for name, signal := range signals {
				signal.registerFor(name, gtype)
			}

			// then allow the user to call some methods on the class, e.g. to supply metadata
			classInit(class)
		},
		newFromGlib: func(cInstance unsafe.Pointer) InstanceT {
			obj := wrapObject(cInstance)
			// parent is the pointer to the parent instance
			parent := parentWrapObject(obj)

			instance := constructor()

			// set the parent instance, aka the first embedded field.
			// the embedded field is not the parent interface though, but instead the instance struct
			// so we need to deref the pointer and set the first field

			parentInstance := reflect.ValueOf(parent).Elem()

			if parentInstance.Kind() == reflect.Ptr {
				parentInstance = parentInstance.Elem()
			}

			if parentInstance.Kind() != reflect.Struct {
				log.Panicln("parent instance is not a struct")
			}

			instanceValue := reflect.ValueOf(instance).Elem()

			if instanceValue.Kind() == reflect.Ptr {
				instanceValue = instanceValue.Elem()
			}

			if instanceValue.Kind() != reflect.Struct {
				log.Panicln("instance is not a struct")
			}

			// panic on initialized first field
			if !instanceValue.Field(0).IsZero() {
				log.Panicln("instance first field is already set")
			}

			// panic on type mismatch of the first field
			if instanceValue.Field(0).Type() != parentInstance.Type() {
				log.Panicf("instance first field is not of the same type as parents instance type. expected %s, got %s\n", parentInstance.Type(), instanceValue.Field(0).Type())
			}

			instanceValue.Field(0).Set(parentInstance)

			return instance
		},
	}

	dataKey := userdata.Register(data)

	// this can be allocated by go because the parameter is owned by the caller
	typeInfo := &C.GTypeInfo{
		// not needed:
		value_table:    nil,
		base_init:      nil,
		base_finalize:  nil,
		n_preallocs:    0,
		class_finalize: nil,

		instance_size: 0, // not required

		class_size:    C.guint16(typeQuery.class_size),
		class_init:    C.GClassInitFunc(C._gotk4ClassInit),
		instance_init: C.GInstanceInitFunc(C._gotk4InstanceInit),
		class_data:    C.gconstpointer(dataKey),
	}

	_ = typeInfo

	panic("not implemented")
}

type SignalDefinition struct {
	Flags SignalFlags
	// ParamTypes is a list of parameter types. The instance parameter can be omitted, as it is
	// automatically added by the bindings.
	ParamTypes []Type
	ReturnType Type

	// Accumulator is the SignalAccumulator. Can be nil.
	Accumulator SignalAccumulator

	// Handler must be a function with the correct signature of the signal. The first (or receiver) parameter
	// must be the instance type
	Handler any
}

func (sd SignalDefinition) registerFor(name string, typ Type) {
	NewSignal(name, typ, sd.Flags, sd.Handler, sd.Accumulator, sd.ParamTypes, sd.ReturnType)
}

type SubClassInterfaceInit[InstanceT Object] struct {
	InterfaceType  Type
	ApplyOverrides func(gclass unsafe.Pointer)
	FromGlib       func(unsafe.Pointer) any
}

// toInterfaceInfo returns the GInterfaceInfo struct for the given interface type.
func (i SubClassInterfaceInit[InstanceT]) toInterfaceInfo() *C.GInterfaceInfo {
	applyOverridesData := userdata.Register(i.ApplyOverrides)

	return &C.GInterfaceInfo{
		interface_init:     C.GInterfaceInitFunc(C._gotk4InterfaceInit),
		interface_finalize: nil,
		interface_data:     C.gpointer(applyOverridesData),
	}
}
