package convert

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type CToGoNullableStringConverter struct {
	Param        *typesystem.Param
	SubConverter *CToGoStringConverter
}

// Convert implements Converter.
func (c *CToGoNullableStringConverter) Convert(w file.CodeWriter) {
	fmt.Fprintf(w, "if %s != nil {\n", c.Param.CName)
	w.Indent()
	c.SubConverter.Convert(w)
	w.Unindent()
	fmt.Fprintf(w, "}\n")
}

// Metadata implements Converter.
func (c *CToGoNullableStringConverter) Metadata() string {
	return fmt.Sprintf("%s, nullable", c.SubConverter.Metadata())
}

var _ Converter = (*CToGoNullableStringConverter)(nil)

type GoToCNullableStringConverter struct {
	Param        *typesystem.Param
	SubConverter *GoToCStringConverter
}

// Convert implements Converter.
func (c *GoToCNullableStringConverter) Convert(w file.CodeWriter) {
	fmt.Fprintf(w, "if %s != \"\" {\n", c.Param.GoName)
	w.Indent()
	c.SubConverter.Convert(w)
	w.Unindent()
	fmt.Fprintf(w, "}\n")
}

// Metadata implements Converter.
func (c *GoToCNullableStringConverter) Metadata() string {
	return fmt.Sprintf("%s, nullable", c.SubConverter.Metadata())
}

var _ Converter = (*GoToCNullableStringConverter)(nil)
