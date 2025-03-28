package convert

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type CToGoRecordPointerConverter struct {
	Param  *typesystem.Param
	Record typesystem.Type // Foreign Record or Record
}

// Convert implements Converter.
func (c *CToGoRecordPointerConverter) Convert(w file.CodeWriter) {
	// TODO import unsafe

	fmt.Fprintf(w, "%s = %s(unsafe.Pointer(%s))\n", c.Param.GoName, c.conversionFunc(), c.Param.CName)

	if c.Param.TransferOwnership == typesystem.TransferBorrow {
		// TODO import runtime

		fmt.Fprintf(w, "runtime.AddCleanup(%s, func(_ %s) {}, %s)\n", c.Param.GoName, c.Param.BorrowFrom.Type.GoType(), c.Param.BorrowFrom.GoName)
	}
}

// Metadata implements Converter.
func (c *CToGoRecordPointerConverter) Metadata() string {
	return fmt.Sprintf("%s, %s, record", c.Param.Direction, c.Param.TransferOwnership)
}

func (c *CToGoRecordPointerConverter) conversionFunc() string {
	r := typesystem.UnderlyingType(c.Record).(*typesystem.Record)

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

	if foreign, ok := c.Record.(*typesystem.ForeignType); ok {
		return foreign.AddForeignNamespace(f)
	}

	return f
}

var _ Converter = (*CToGoRecordPointerConverter)(nil)

// GoToCRecordPointerConverter is needed because go's true differs from c's true
type GoToCRecordPointerConverter struct {
	Param  *typesystem.Param
	Record typesystem.Type // Foreign Record or Record
}

// Convert implements Converter.
func (c *GoToCRecordPointerConverter) Convert(w file.CodeWriter) {
	fmt.Fprintf(w, "%s = (%s)(%s(%s))\n", c.Param.CName, c.Param.Type.CGoType(), c.conversionFunc(), c.Param.GoName)
}

// Metadata implements Converter.
func (c *GoToCRecordPointerConverter) Metadata() string {
	return fmt.Sprintf("%s, %s, record", c.Param.Direction, c.Param.TransferOwnership)
}

func (c *GoToCRecordPointerConverter) conversionFunc() string {
	r := typesystem.UnderlyingType(c.Record).(*typesystem.Record)

	var f string

	switch c.Param.TransferOwnership {
	case typesystem.TransferFull:
		f = r.GoUnsafeToGlibFullFunction
	case typesystem.TransferNone:
		f = r.GoUnsafeToGlibNoneFunction
	default:
		panic(fmt.Sprintf("unexpected typesystem.TransferOwnership: %#v", c.Param.TransferOwnership))
	}

	if foreign, ok := c.Record.(*typesystem.ForeignType); ok {
		return foreign.AddForeignNamespace(f)
	}

	return f
}

var _ Converter = (*GoToCRecordPointerConverter)(nil)
