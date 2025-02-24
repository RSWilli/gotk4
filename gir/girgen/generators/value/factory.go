package value

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
)

func NewInstanceParamConverter(ctx gencontext.GenerationContext, param gir.InstanceParameter) Converter {
	panic("unimplemented")
}

func NewReturnConverter(ctx gencontext.GenerationContext, param gir.ReturnValue) Converter {
	return nil
}

// NewParamConverter creates an appropriate converter for the given param. The index is used to allow the converter
// to create unique variable names for the conversion
func NewParamConverter(ctx gencontext.GenerationContext, paramindex int, param gir.Parameter) Converter {
	return nil
}
