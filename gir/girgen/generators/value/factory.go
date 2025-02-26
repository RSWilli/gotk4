package value

import (
	"fmt"
	"log"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
	"github.com/diamondburned/gotk4/gir/girgen/types"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

func NewInstanceParamConverter(ctx gencontext.GenerationContext, param gir.InstanceParameter) Converter {
	panic("unimplemented")
}

func NewReturnConverter(ctx gencontext.GenerationContext, direction ConversionDirection, ret gir.ReturnValue) Converter {
	if ret.AnyType.Type == nil {
		log.Println("skipping array return type")
		return nil // TODO: handle array types
	}

	if ret.Type.Name == "none" {
		return NoopConverter{}
	}

	meta := ctx.LookupType(ret.Type.CType)

	if meta == nil {
		return nil // unknown type
	}

	inType := inType(meta, direction)
	outType := outType(meta, direction)

	defaultInIdent := inIdent(direction, "cret", ret.Type.Name)
	defaultOutIdent := outIdent(direction, "cret", ret.Type.Name)

	switch {
	case meta.CGoBaseType == "C.gboolean":
		return &BooleanConverter{
			Direction: direction,
			InIdent:   inIdent(direction, "cret", "ok"),
			InTyp:     inType,
			OutIdent:  outIdent(direction, "cret", "ok"),
			OutTyp:    outType,
		}
	case meta.GoType() == "string":
		return &StringConverter{
			Direction: direction,
			InIdent:   defaultInIdent,
			InTyp:     inType,
			OutIdent:  defaultOutIdent,
			OutTyp:    outType,
		}
	case meta.CGoBaseType == "C.gpointer":
		return nil // cannot convert gpointer
	}

	switch girType := meta.GirType.(type) {
	case *gir.Enum:
		return &EnumConverter{
			Direction: direction,
			InIdent:   defaultInIdent,
			InTyp:     inType,
			OutIdent:  defaultOutIdent,
			OutTyp:    outType,
		}
	case *gir.Record:
		_ = girType // TODO: the girType contains infos about the transfer mode and more
		return &RecordConverter{
			Direction: direction,
			InIdent:   defaultInIdent,
			InTyp:     inType,
			OutIdent:  defaultOutIdent,
			OutTyp:    outType,
		}
	}

	panic("unhandled case for returnvalue conversion: " + meta.CGoType() + " " + meta.GoType())
}

// NewParamConverter creates an appropriate converter for the given param. The index is used to allow the converter
// to create unique variable names for the conversion
func NewParamConverter(ctx gencontext.GenerationContext, direction ConversionDirection, paramindex int, param gir.Parameter) Converter {
	if param.AnyType.Type == nil {
		log.Println("skipping array param")
		return nil // TODO: handle array types
	}

	meta := ctx.LookupType(param.Type.CType)

	if meta == nil {
		return nil // unknown type
	}

	argName := fmt.Sprintf("arg%d", paramindex+1)
	outName := param.Name // FIXME: maybe this is better? fmt.Sprintf("_%s", param.Name)

	guessedParamDir := types.GuessParameterOutput(&param)

	if guessedParamDir == "out" {
		direction = direction.Switch()
		meta.GoPointers-- // one pointer less needed, because the converter will add it back

		argName, outName = outName, argName
	}

	inType := inType(meta, direction)
	outType := outType(meta, direction)

	// primitive cases
	switch {
	case meta.CGoBaseType == "C.gpointer" && param.Closure != nil && direction == ConvertCToGo:
		// we are converting a user_data param that should yield the function from a gbox, this is handled in the callback generator
		return NoopConverter{
			dir:    direction,
			inName: argName,
			inType: inType,
		}
	case meta.CGoBaseType == "C.gboolean":
		return &BooleanConverter{
			Direction: direction,
			InIdent:   argName,
			InTyp:     inType,
			OutIdent:  "ok",
			OutTyp:    outType,
		}
	case meta.CGoBaseType == "C.guint" ||
		meta.CGoBaseType == "C.gdouble" ||
		meta.CGoBaseType == "C.gssize" ||
		meta.CGoBaseType == "C.gsize" ||
		meta.CGoBaseType == "C.guint64" ||
		meta.CGoBaseType == "C.guint8":
		return &PrimitiveConverter{
			Direction: direction,
			InIdent:   argName,
			InTyp:     inType,
			OutIdent:  outName,
			OutTyp:    outType,
		}
	}

	if meta.CGoBaseType == "C.gpointer" {
		return nil // cannot handle gpointer if not a userdata arg
	}

	// cases that might be in other packages generated
	switch girType := meta.GirType.(type) {
	case *gir.Record:
		return NewRecordConverter(
			ctx,
			girType,
			direction,
			argName,
			inType,
			outName,
			outType,
		)
	case *gir.Class:
		return &ClassConverter{
			Direction: direction,
			InIdent:   argName,
			InTyp:     inType,
			OutIdent:  outName,
			OutTyp:    outType,
		}
	case *gir.Alias:
		return &AliasConverter{
			Direction: direction,
			InIdent:   argName,
			InTyp:     inType,
			OutIdent:  outName,
			OutTyp:    outType,
		}
	case *gir.Enum:
		return &EnumConverter{
			Direction: direction,
			InIdent:   argName,
			InTyp:     inType,
			OutIdent:  outName,
			OutTyp:    outType,
		}
	case *gir.Bitfield:
		return &BitfieldConverter{
			Direction: direction,
			InIdent:   argName,
			InTyp:     inType,
			OutIdent:  outName,
			OutTyp:    outType,
		}
	}

	panic("unhandled case for parameter conversion: " + meta.CGoBaseType + " " + meta.GoBaseType)
}

func inType(meta *typesystem.TypeMetadata, dir ConversionDirection) string {
	switch dir {
	case ConvertCToGo:
		return meta.CGoType()
	case ConvertGoToC:
		return meta.GoType()
	default:
		panic(fmt.Sprintf("unexpected value.ConversionDirection: %#v", dir))
	}
}

func outType(meta *typesystem.TypeMetadata, dir ConversionDirection) string {
	switch dir {
	case ConvertCToGo:
		return meta.GoType()
	case ConvertGoToC:
		return meta.CGoType()
	default:
		panic(fmt.Sprintf("unexpected value.ConversionDirection: %#v", dir))
	}
}

func inIdent(dir ConversionDirection, cIndent, goIndent string) string {
	switch dir {
	case ConvertCToGo:
		return cIndent
	case ConvertGoToC:
		return goIndent
	default:
		panic(fmt.Sprintf("unexpected value.ConversionDirection: %#v", dir))
	}
}

func outIdent(dir ConversionDirection, cIndent, goIndent string) string {
	switch dir {
	case ConvertCToGo:
		return goIndent
	case ConvertGoToC:
		return cIndent
	default:
		panic(fmt.Sprintf("unexpected value.ConversionDirection: %#v", dir))
	}
}
