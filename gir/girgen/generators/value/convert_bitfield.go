package value

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

func newBitfieldParamConverter(meta *typesystem.TypeMetadata, girType *gir.Bitfield, param gir.Parameter) Converter {
	return NoopConverter{}
}
