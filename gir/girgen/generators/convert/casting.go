package convert

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type CToGoCastingConverter struct {
	Param *typesystem.Param
}

// Convert implements Converter.
func (c *CToGoCastingConverter) Convert(w file.CodeWriter) {

	fmt.Fprintf(w, "%s = %s(unsafe.Pointer(%s))\n", c.Param.GoName, c.Param.GoType(), c.Param.CName)
}

// Metadata implements Converter.
func (c *CToGoCastingConverter) Metadata() string {
	return fmt.Sprintf("%s, %s, casted", c.Param.Direction, c.Param.TransferOwnership)
}

var _ Converter = (*CToGoCastingConverter)(nil)

type GoToCCastingConverter struct {
	Param *typesystem.Param
}

// Convert implements Converter.
func (c *GoToCCastingConverter) Convert(w file.CodeWriter) {
	fmt.Fprintf(w, "%s = %s(%s)\n", c.Param.CName, c.Param.CGoType(), c.Param.GoName)
}

// Metadata implements Converter.
func (c *GoToCCastingConverter) Metadata() string {
	return fmt.Sprintf("%s, %s, casted", c.Param.Direction, c.Param.TransferOwnership)
}

var _ Converter = (*GoToCCastingConverter)(nil)
