package convert

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)


type CToGoInterfaceReturnConverter struct {
	Param        *typesystem.Param
	SubConverter Converter
}

var _ Converter = (*CToGoInterfaceReturnConverter)(nil)

func (c *CToGoInterfaceReturnConverter) Convert(w file.File) {
	c.SubConverter.Convert(w)
}

// Metadata implements Converter.
func (c *CToGoInterfaceReturnConverter) Metadata() string {
	return fmt.Sprintf("%s, returned interface", c.SubConverter.Metadata())
}
