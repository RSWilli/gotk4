package convert

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type CToGoStringConverter struct {
	Param *typesystem.Param
}

// Convert implements Converter.
func (c *CToGoStringConverter) Convert(w file.CodeWriter) {
	fmt.Fprintf(w, "%s = C.GoString((%s)(unsafe.Pointer(%s)))\n", c.Param.GoName, c.Param.CGoType(), c.Param.CName)

	switch c.Param.TransferOwnership {
	case typesystem.TransferFull:
		// GoString copies the param, so free it immediately
		fmt.Fprintf(w, "defer C.free(unsafe.Pointer(%s))\n", c.Param.CName)
	case typesystem.TransferNone:
		// C will free it
	default:
		panic(fmt.Sprintf("unexpected typesystem.TransferOwnership: %#v", c.Param.TransferOwnership))
	}
}

// Metadata implements Converter.
func (c *CToGoStringConverter) Metadata() string {
	return fmt.Sprintf("%s, %s, string", c.Param.Direction, c.Param.TransferOwnership)
}

var _ Converter = (*CToGoStringConverter)(nil)

type GoToCStringConverter struct {
	Param *typesystem.Param
}

// Convert implements Converter.
func (c *GoToCStringConverter) Convert(w file.CodeWriter) {
	// TODO: import unsafe

	fmt.Fprintf(w, "%s = (%s)(unsafe.Pointer(C.CString(%s)))\n", c.Param.CName, c.Param.CGoType(), c.Param.GoName)

	switch c.Param.TransferOwnership {
	case typesystem.TransferFull:
		return // C will free it
	case typesystem.TransferNone:
		fmt.Fprintf(w, "defer C.free(unsafe.Pointer(%s))\n", c.Param.CName)
	default:
		panic(fmt.Sprintf("unexpected typesystem.TransferOwnership: %#v", c.Param.TransferOwnership))
	}
}

// Metadata implements Converter.
func (c *GoToCStringConverter) Metadata() string {
	return fmt.Sprintf("%s, %s, string", c.Param.Direction, c.Param.TransferOwnership)
}

var _ Converter = (*GoToCStringConverter)(nil)
