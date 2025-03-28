package convert

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

// GoToCCastingConverter is used for all types that are directly castable
type GoToCCastingConverter struct {
	Param *typesystem.Param
}

// Convert implements Converter.
func (c *GoToCCastingConverter) Convert(w file.CodeWriter) {
	fmt.Fprintf(w, "%s = %s(%s)\n", c.Param.CName, c.Param.Type.CGoType(), c.Param.GoName)
}

// Metadata implements Converter.
func (c *GoToCCastingConverter) Metadata() string {
	return fmt.Sprintf("%s, casted", c.Param.Direction)
}

var _ Converter = (*GoToCCastingConverter)(nil)

// CToGoCastingConverter is used for all types that are directly castable
type CToGoCastingConverter struct {
	Param *typesystem.Param
}

// Convert implements Converter.
func (c *CToGoCastingConverter) Convert(w file.CodeWriter) {
	fmt.Fprintf(w, "%s = %s(%s)\n", c.Param.GoName, c.Param.Type.GoType(), c.Param.CName)
}

// Metadata implements Converter.
func (c *CToGoCastingConverter) Metadata() string {
	return fmt.Sprintf("%s, casted", c.Param.Direction)
}

var _ Converter = (*CToGoCastingConverter)(nil)
