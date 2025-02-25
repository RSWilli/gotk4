package value

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

func newRecordParamConverter(ctx gencontext.GenerationContext, paramindex ConversionValueIndex, param gir.Parameter, meta *typesystem.TypeMetadata, girType *gir.Record) Converter {
	return NoopConverter{}
}

func newRecordReturnConverter(ret gir.ReturnValue, meta *typesystem.TypeMetadata, girType *gir.Record) Converter {
	return NoopConverter{}
}
