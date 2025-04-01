package convert

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type CToGoClassPointerConverter struct {
	Param *typesystem.Param
	Class typesystem.Type // Foreign Record or Record
}

// Convert implements Converter.
func (c *CToGoClassPointerConverter) Convert(w file.CodeWriter) {

	fmt.Fprintf(w, "%s = %s(unsafe.Pointer(%s))\n", c.Param.GoName, c.conversionFunc(), c.Param.CName)

	if c.Param.TransferOwnership == typesystem.TransferBorrow {
		// TODO import runtime

		fmt.Fprintf(w, "runtime.AddCleanup(%s, func(_ %s) {}, %s)\n", c.Param.GoName, c.Param.BorrowFrom.Type.GoType(), c.Param.BorrowFrom.GoName)
	}
}

// Metadata implements Converter.
func (c *CToGoClassPointerConverter) Metadata() string {
	return fmt.Sprintf("%s, %s, record", c.Param.Direction, c.Param.TransferOwnership)
}

func (c *CToGoClassPointerConverter) conversionFunc() string {
	r := typesystem.UnderlyingType(c.Class).(*typesystem.Class)

	var f string

	switch c.Param.TransferOwnership {
	case typesystem.TransferFull:
		f = r.GoUnsafeFromGlibFullFunction
	case typesystem.TransferBorrow:
		f = r.GoUnsafeFromGlibBorrowFunction
	case typesystem.TransferNone:
		f = r.GoUnsafeFromGlibNoneFunction
	default:
		panic(fmt.Sprintf("unexpected typesystem.TransferOwnership: %#v", c.Param.TransferOwnership))
	}

	if foreign, ok := c.Class.(*typesystem.ForeignType); ok {
		return foreign.AddForeignNamespace(f)
	}

	return f
}

var _ Converter = (*CToGoClassPointerConverter)(nil)

type GoToCClassPointerConverter struct {
	Param *typesystem.Param
	Class typesystem.Type // Foreign Record or Record
}

// Convert implements Converter.
func (c *GoToCClassPointerConverter) Convert(w file.CodeWriter) {
	fmt.Fprintf(w, "%s = (%s)(%s(%s))\n", c.Param.CName, c.Param.Type.CGoType(), c.conversionFunc(), c.Param.GoName)
}

// Metadata implements Converter.
func (c *GoToCClassPointerConverter) Metadata() string {
	return fmt.Sprintf("%s, %s, record", c.Param.Direction, c.Param.TransferOwnership)
}

func (c *GoToCClassPointerConverter) conversionFunc() string {
	r := typesystem.UnderlyingType(c.Class).(*typesystem.Class)

	var f string

	switch c.Param.TransferOwnership {
	case typesystem.TransferFull:
		f = r.GoUnsafeToGlibFullFunction
	case typesystem.TransferNone:
		f = r.GoUnsafeToGlibNoneFunction
	default:
		panic(fmt.Sprintf("unexpected typesystem.TransferOwnership: %#v", c.Param.TransferOwnership))
	}

	if foreign, ok := c.Class.(*typesystem.ForeignType); ok {
		return foreign.AddForeignNamespace(f)
	}

	return f
}

var _ Converter = (*GoToCClassPointerConverter)(nil)
