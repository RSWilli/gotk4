package convert

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type CToGoClassReturnConverter struct {
	Param        *typesystem.Param
	SubConverter Converter
}

var _ Converter = (*CToGoClassReturnConverter)(nil)

func (c *CToGoClassReturnConverter) Convert(w file.File) {
	c.SubConverter.Convert(w)
}

// Metadata implements Converter.
func (c *CToGoClassReturnConverter) Metadata() string {
	return fmt.Sprintf("%s, returned class", c.SubConverter.Metadata())
}
