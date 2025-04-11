package gdkpixbuf

// #cgo pkg-config: gdk-pixbuf-2.0
// #cgo CFLAGS: -Wno-deprecated-declarations
// #include <gdk-pixbuf/gdk-pixbuf.h>
import "C"

// TODO: port this to the new bindings

// // NewPixbufFromImage creates a new Pixbuf from a stdlib image.Image. It
// // contains a fast path for *image.RGBA while resorting to
// // copying/converting the image otherwise.
// func NewPixbufFromImage(img image.Image) *Pixbuf {
// 	bounds := img.Bounds()
// 	var pixbuf *Pixbuf

// 	switch img := img.(type) {
// 	case *image.RGBA:
// 		bytes := glib.NewBytesWithGo(img.Pix)
// 		pixbuf = NewPixbufFromBytes(bytes, ColorspaceRGB, true, 8, bounds.Dx(), bounds.Dy(), img.Stride)
// 	default:
// 		pixbuf = NewPixbuf(ColorspaceRGB, true, 8, bounds.Dx(), bounds.Dy())
// 		pixbuf.ReadPixelBytes().Use(func(b []byte) {
// 			// For information on how this works, refer to
// 			// pkg/cairo/surface_image.go.
// 			rgba := image.RGBA{
// 				Pix:    b,
// 				Stride: bounds.Dx(),
// 				Rect:   bounds,
// 			}
// 			draw.Draw(&rgba, rgba.Rect, img, image.Point{}, draw.Over)
// 			swizzle.BGRA(rgba.Pix)
// 		})
// 	}

// 	return pixbuf
// }