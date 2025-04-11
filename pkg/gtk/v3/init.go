package gtk

// #cgo pkg-config: gtk+-3.0
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <gtk/gtk-a11y.h>
// #include <gtk/gtk.h>
// #include <gtk/gtkx.h>
import "C"

// Init binds to the gtk_init() function. Argument parsing is not
// supported.
func Init() {
	C.gtk_init(nil, nil)
}
