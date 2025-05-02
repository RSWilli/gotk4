package gobject

// #include <glib-object.h>
import "C"
import (
	"unsafe"

	"github.com/diamondburned/gotk4/pkg/core/userdata"
)

// _gotk4InterfaceInit is the function that is called by the GObject system when the interface is initialized on
// the subclass. This is called only once and applies the interface overrides.
//
//export _gotk4InterfaceInit
func _gotk4InterfaceInit(instance C.gpointer, ifaceData C.gpointer) {
	ptr := unsafe.Pointer(ifaceData)
	// if the interfaceData is deleted then we can never extend the new class.
	// defer userdata.Delete(ptr)

	// Call the downstream interface init handlers
	data := userdata.Load(ptr).(func(gclass unsafe.Pointer))

	data(unsafe.Pointer(instance))
}

// _gotk4ClassInit is the function that is called by the GObject system when the class is initialized on
// the subclass. This is called only once and applies the class overrides.
//
//export _gotk4ClassInit
func _gotk4ClassInit(gclass C.gpointer, classData C.gpointer) {
	panic("unimplemented")
}

// _gotk4InstanceInit is the function that is called by the GObject system when the instance is initialized on
// the subclass. This is called for each instance and applies the instance overrides.
//
//export _gotk4InstanceInit
func _gotk4InstanceInit(instance *C.GTypeInstance, gclass C.gpointer) {
	panic("unimplemented")
}
