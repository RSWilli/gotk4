package value

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
)

func NewInstanceParamConverter(ctx gencontext.GenerationContext, param gir.InstanceParameter) Converter {
	panic("unimplemented")
}

func NewReturnConverter(ctx gencontext.GenerationContext, ret gir.ReturnValue) Converter {
	if ret.AnyType.Type == nil {
		return nil // TODO: handle array types
	}

	if ret.Type.Name == "none" {
		return NoopConverter{}
	}

	meta := ctx.LookupType(ret.Type.CType)

	if meta == nil {
		return nil // unknown type
	}

	switch {
	case meta.CGoBaseType == "C.gboolean":
		return newBooleanReturnConverter(meta)
	case meta.GoType() == "string":
		return newStringReturnConverter(meta)
	case meta.CGoBaseType == "C.gpointer":
		return nil // cannot convert gpointer
	}

	switch girType := meta.GirType.(type) {
	case *gir.Enum:
		return newEnumReturnConverter(ret, meta, girType)
	case *gir.Record:
		return newRecordReturnConverter(ret, meta, girType)
	}

	panic("unhandled case for returnvalue conversion: " + meta.CGoType() + " " + meta.GoType())
}

// NewParamConverter creates an appropriate converter for the given param. The index is used to allow the converter
// to create unique variable names for the conversion
//
// goCallBackTypeName
func NewParamConverter(ctx gencontext.GenerationContext, goCallBackTypeName string, paramindex int, param gir.Parameter) Converter {
	if param.AnyType.Type == nil {
		return nil // TODO: handle array types
	}

	meta := ctx.LookupType(param.Type.CType)

	if meta == nil {
		return nil // unknown type
	}

	// primitive cases
	switch {
	case meta.CGoType == "C.gpointer" && param.Closure != nil && goCallBackTypeName != "":
		// we are converting a user_data param that should yield the function from a gbox
		return newUserDataParamConverter(ctx, goCallBackTypeName, ConversionValueIndex(paramindex), param)
	case meta.CGoType == "C.gboolean":
		return newBooleanParamConverter(meta, param)
	case meta.CGoType == "C.guint":
		return newPrimitiveParamConverter(meta, param)
	case meta.CGoType == "C.gdouble":
		return newPrimitiveParamConverter(meta, param)
	case meta.CGoType == "C.gssize":
		return newPrimitiveParamConverter(meta, param)
	case meta.CGoType == "C.gsize":
		return newPrimitiveParamConverter(meta, param)
	case meta.CGoType == "C.guint64":
		return newPrimitiveParamConverter(meta, param)
	}

	if meta.CGoType == "C.gpointer" {
		return nil // cannot handle gpointer if not a userdata arg
	}

	// cases that might be in other packages generated
	switch girType := meta.GirType.(type) {
	case *gir.Record:
		return newRecordParamConverter(ctx, ConversionValueIndex(paramindex), param, meta, girType)
	case *gir.Class:
		return newClassParamConverter(ctx, ConversionValueIndex(paramindex), param, meta, girType)
	case *gir.Alias:
		return newAliasParamConverter(meta, girType, param)
	case *gir.Enum:
		return newEnumParamConverter(meta, girType, param)
	case *gir.Bitfield:
		return newBitfieldParamConverter(meta, girType, param)
	}

	panic("unhandled case for parameter conversion: " + meta.CGoType + " " + meta.GoType)
}
