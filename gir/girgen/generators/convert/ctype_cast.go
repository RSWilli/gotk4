package convert

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type OverridenCtypeConverter struct {
	Param        *typesystem.Param
	Type         typesystem.OverriddenCType
	SubConverter Converter
}

var _ Converter = (*OverridenCtypeConverter)(nil)

func (c *OverridenCtypeConverter) Convert(w file.File) {
	c.SubConverter.Convert(w)
}

// Metadata implements Converter.
func (c *OverridenCtypeConverter) Metadata() string {
	return fmt.Sprintf("%s, casted %s", c.SubConverter.Metadata(), c.Type.OriginalCGoType(c.Param.CTypePointers))
}
