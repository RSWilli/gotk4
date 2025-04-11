package gtk

// #cgo pkg-config: gtk4
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <gtk/gtk.h>
import "C"

// TODO: port this to the new bindings:

// // NewDialogWithFlags is a slightly more advanced version of NewDialog,
// // allowing the user to construct a new dialog with the given
// // constructor-only dialog flags.
// //
// // It is a wrapper around Gtk.Dialog.new_with_buttons in C.
// func NewDialogWithFlags(title string, parent *Window, flags DialogFlags) *Dialog {
// 	ctitle := C.CString(title)
// 	defer C.free(unsafe.Pointer(ctitle))

// 	w := C._gotk4_gtk3_dialog_new2(
// 		(*C.gchar)(unsafe.Pointer(ctitle)),
// 		(*C.GtkWindow)(unsafe.Pointer(coreglib.InternObject(parent).Native())),
// 		(C.GtkDialogFlags)(flags),
// 	)
// 	runtime.KeepAlive(parent)

// 	return wrapDialog(coreglib.Take(unsafe.Pointer(w)))
// }

// // NewMessageDialog creates a new message dialog. This is a simple
// // dialog with some text taht the user may want to see. When the user
// // clicks a button, a "response" signal is emitted with response IDs
// // from ResponseType.
// func NewMessageDialog(parent *Window, flags DialogFlags, typ MessageType, buttons ButtonsType) *MessageDialog {
// 	w := C._gotk4_gtk_message_dialog_new2(
// 		(*C.GtkWindow)(unsafe.Pointer(coreglib.InternObject(parent).Native())),
// 		(C.GtkDialogFlags)(flags),
// 		(C.GtkMessageType)(typ),
// 		(C.GtkButtonsType)(buttons),
// 	)
// 	runtime.KeepAlive(parent)

// 	return wrapMessageDialog(coreglib.Take(unsafe.Pointer(w)))
// }