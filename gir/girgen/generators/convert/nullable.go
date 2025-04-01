package convert

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type CToGoNullableConverter struct {
	Param        *typesystem.Param
	SubConverter Converter
}

// Convert implements Converter.
func (c *CToGoNullableConverter) Convert(w file.CodeWriter) {
	fmt.Fprintf(w, "if %s != nil {\n", c.Param.CName)
	w.Indent()
	c.SubConverter.Convert(w)
	w.Unindent()
	fmt.Fprintf(w, "}\n")
}

// Metadata implements Converter.
func (c *CToGoNullableConverter) Metadata() string {
	return fmt.Sprintf("%s, nullable", c.SubConverter.Metadata())
}

var _ Converter = (*CToGoNullableConverter)(nil)

// GoToCNullableConverter is needed because go's true differs from c's true
type GoToCNullableConverter struct {
	Param        *typesystem.Param
	SubConverter Converter
}

// Convert implements Converter.
func (c *GoToCNullableConverter) Convert(w file.CodeWriter) {
	fmt.Fprintf(w, "if %s != nil {\n", c.Param.GoName)
	w.Indent()
	c.SubConverter.Convert(w)
	w.Unindent()
	fmt.Fprintf(w, "}\n")
}

// Metadata implements Converter.
func (c *GoToCNullableConverter) Metadata() string {
	return fmt.Sprintf("%s, nullable", c.SubConverter.Metadata())
}

var _ Converter = (*GoToCNullableConverter)(nil)
