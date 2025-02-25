package value

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

func newEnumReturnConverter(ret gir.ReturnValue, meta *typesystem.TypeMetadata, girType *gir.Enum) Converter {
	return NoopConverter{}
}

func newEnumParamConverter(meta *typesystem.TypeMetadata, girType *gir.Enum, param gir.Parameter) Converter {
	return NoopConverter{}
}
