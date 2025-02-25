package value

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

func newPrimitiveParamConverter(meta *typesystem.TypeMetadata, param gir.Parameter) Converter {
	return NoopConverter{}
}
