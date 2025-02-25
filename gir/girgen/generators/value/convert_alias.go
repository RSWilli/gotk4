package value

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

func newAliasParamConverter(meta *typesystem.TypeMetadata, girType *gir.Alias, param gir.Parameter) Converter {
	return NoopConverter{}
}
