// package profile contains the pprof functionality for tracking memory leaks by counting the references taken
// when a c type is wrapped in a Go type.
package profile

import (
	"io"
	"runtime"
	"runtime/pprof"
	"sync"
)

var prof *pprof.Profile

var disabled bool

var initOnce sync.Once
var initProfile = func() {
	initOnce.Do(func() {
		if disabled {
			return
		}

		objects := "gotk4/alive-objects"
		prof = pprof.Lookup(objects)
		if prof == nil {
			prof = pprof.NewProfile(objects)
		}
	})
}

// Disable disables the pprof profile for tracking memory leaks. This only prevents further
// tracking of objects, but does not remove existing tracked objects from the profile, nor
// does it remove the profile itself if it was created before.
//
// this function should thus be called before any objects are tracked, e.g. at the start of the application.
func Disable() {
	disabled = true

	prof = nil
}

// Track takes an uintpr and adds it to the profile. The uintpr must be only added once. Keep in mind
// that simply adding C pointers may not be unique.
//
// skip allows to skip more stack frames when tracking the object. skip=0 means the
// stack trace starts at the call to Track.
func Track(ptr uintptr, skip int) {
	initProfile()
	if prof == nil {
		return
	}

	prof.Add(ptr, 1+skip)

	debugTrack(ptr)
}

// Untrack removes an uintpr to a go object that wraps a C type from the
// pprof profile. Don't use the C pointer directly, because we can have multiple references to
// the same C type in Go, which would panic in pprof.
func Untrack(ptr uintptr) {
	if prof == nil {
		return
	}

	prof.Remove(ptr)

	debugUntrack(ptr)
}

// Count returns [pprof.Profile.Count]
func Count() int {
	initProfile()
	if prof == nil {
		return 0
	}

	return prof.Count()
}

// WriteTo writes the profile to w
func WriteTo(w io.Writer, debug int) error {
	initProfile()
	if prof == nil {
		return nil
	}

	return prof.WriteTo(w, debug)
}

var (
	debugMutex       sync.Mutex
	debuggingEnabled bool
	// debugStackTraces storesthe pc values of the stack trace
	debugStackTraces map[uintptr][]uintptr
)

// EnableDebugging enables a secondary tracking of objects for debugging purposes.
func EnableDebugging() {
	debuggingEnabled = true

	debugStackTraces = make(map[uintptr][]uintptr)
}

// debugTrack adds the stack trace of the current goroutine to the debug tracking map.
// This is only used if debugging is enabled.
func debugTrack(ptr uintptr) {
	if !debuggingEnabled {
		return
	}

	debugMutex.Lock()
	defer debugMutex.Unlock()

	var stack [10]uintptr
	n := runtime.Callers(2, stack[:])
	debugStackTraces[ptr] = stack[:n]
}

// debugUntrack remove and logs the stack trace for the given pointer.
func debugUntrack(ptr uintptr) {
	if !debuggingEnabled {
		return
	}

	debugMutex.Lock()
	defer debugMutex.Unlock()

	stack := debugStackTraces[ptr]

	delete(debugStackTraces, ptr)

	frames := runtime.CallersFrames(stack)

	println("untracking object tracked at")

	for {
		frame, more := frames.Next()
		if frame.Function == "" {
			break
		}

		println("  ", frame.Function, "at", frame.File, ":", frame.Line)
		if !more {
			break
		}
	}
}
