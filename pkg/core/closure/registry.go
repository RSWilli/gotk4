package closure

import (
	"sync"
	"unsafe"

	"github.com/diamondburned/gotk4/pkg/core/profile"
)

var closureRegistry sync.Map // unsafe.Pointer(*C.GClosure) -> *FuncStack

// Register registers the given GClosure callback. This panics if the
// GClosure is already registered.
func Register(gclosure unsafe.Pointer, callback *FuncStack) {
	if _, ok := closureRegistry.Load(gclosure); ok {
		panic("closure already registered")
	}

	closureRegistry.Store(gclosure, callback)

	profile.Track(uintptr(gclosure), 1)
}

func Load(gclosure unsafe.Pointer) *FuncStack {
	fs, ok := closureRegistry.Load(gclosure)
	if !ok {
		return nil
	}
	return fs.(*FuncStack)
}

// Delete deletes the given GClosure callback. This panics if the
// GClosure is not registered.
func Delete(gclosure unsafe.Pointer) {
	if _, ok := closureRegistry.Load(gclosure); !ok {
		panic("closure not registered")
	}
	closureRegistry.Delete(gclosure)
	profile.Untrack(uintptr(gclosure))
}
