package value

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

func newClassParamConverter(ctx gencontext.GenerationContext, conversionValueIndex ConversionValueIndex, param gir.Parameter, meta *typesystem.TypeMetadata, girType *gir.Class) Converter {
	return NoopConverter{}
}
