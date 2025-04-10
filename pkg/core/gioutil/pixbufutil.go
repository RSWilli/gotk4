//go:build !nopixbuf

package gioutil

// type pixbufLoaderWriter struct {
// 	*gdkpixbuf.PixbufLoader
// }

// // PixbufLoaderWriter wraps a PixbufLoader to satsify io.WriteCloser.
// func PixbufLoaderWriter(l *gdkpixbuf.PixbufLoader) io.WriteCloser {
// 	return pixbufLoaderWriter{l}
// }

// func (w pixbufLoaderWriter) Write(b []byte) (int, error) {
// 	if err := w.PixbufLoader.Write(b); err != nil {
// 		return 0, err
// 	}
// 	return len(b), nil
// }
