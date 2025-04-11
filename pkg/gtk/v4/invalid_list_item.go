package gtk

import (
	"math"
)

// #cgo pkg-config: gtk4
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <gtk/gtk.h>
import "C"

// InvalidListPosition is the value used to refer to a guaranteed
// invalid position in a [gio.ListModel].
//
// This value may be returned from some functions, others may accept it
// as input. Its interpretation may differ for different functions.
//
// Refer to each function’s documentation for if this value is allowed
// and what it does.
const InvalidListPosition = math.MaxUint32
