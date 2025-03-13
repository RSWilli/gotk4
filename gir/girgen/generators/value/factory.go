package value

// import (
// 	"fmt"
// 	"log"

// 	"github.com/diamondburned/gotk4/gir"
// 	"github.com/diamondburned/gotk4/gir/gencontext"
// 	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
// )

// func NewInstanceParamConverter(ctx gencontext.GenerationContext, direction ConversionDirection, param gir.InstanceParameter, cIdent, goIdent string) Converter {
// 	return NewParamConverter(ctx, direction, gir.ParameterAttrs{
// 		Closure:  param.Closure,
// 		Destroy:  param.Destroy,
// 		Scope:    param.Scope,
// 		Skip:     param.Skip,
// 		Nullable: param.Nullable,
// 		Optional: param.Skip,

// 		TransferOwnership: param.TransferOwnership,
// 		AnyType:           param.AnyType,
// 		Doc:               param.Doc,
// 	}, cIdent, goIdent)
// }

// func NewReturnConverter(ctx gencontext.GenerationContext, direction ConversionDirection, ret gir.ReturnValue, cIdent, goIdent string) Converter {
// 	return NewParamConverter(ctx, direction, gir.ParameterAttrs{
// 		Closure:  ret.Closure,
// 		Destroy:  ret.Destroy,
// 		Scope:    ret.Scope,
// 		Skip:     ret.Skip,
// 		Nullable: ret.Nullable,
// 		Optional: ret.Skip,

// 		TransferOwnership: ret.TransferOwnership,
// 		AnyType:           ret.AnyType,
// 		Doc:               ret.Doc,
// 	}, cIdent, goIdent)
// }

// func NewThrowConverter(ctx gencontext.GenerationContext, direction ConversionDirection, cIdent, goIdent string) Converter {
// 	return NewParamConverter(ctx, direction, gir.ParameterAttrs{
// 		TransferOwnership: gir.TransferOwnership{
// 			TransferOwnership: "full",
// 		},
// 		AnyType: gir.AnyType{
// 			Type: &gir.Type{
// 				Name: "GLib.Error",
// 				// Function parameter type is technically a double-pointer
// 				// here.
// 				CType: "GError**",
// 			},
// 		},
// 		Optional:  true,
// 		Nullable:  true,
// 		Direction: "out",
// 	}, cIdent, goIdent)
// }

// // NewParamConverter creates an appropriate converter for the given param. The index is used to allow the converter
// // to create unique variable names for the conversion
// func NewParamConverter(ctx gencontext.GenerationContext, direction ConversionDirection, param gir.ParameterAttrs, cIdent, goIdent string) Converter {
// 	if !isValidGoIndent(cIdent) || !isValidGoIndent(goIdent) {
// 		log.Println("skipping because identifier contains invalid chars")
// 		return nil
// 	}

// 	if param.AnyType.Type == nil {
// 		log.Println("skipping array param")
// 		return nil // TODO: handle array types
// 	}

// 	meta := ctx.LookupType(param.Type.Name, param.Type.CType)

// 	if meta == nil {
// 		return nil // unknown type
// 	}

// 	guessedParamDir := guessParameterOutput(param)

// 	if guessedParamDir == "inout" {
// 		log.Println("skipping inout parameter")
// 		return nil
// 	}

// 	if guessedParamDir == "out" && meta.CGoPointers < 1 {
// 		log.Println("skipping out param without enough pointers")
// 		return nil
// 	}

// 	if guessedParamDir == "out" {
// 		direction = direction.Switch()
// 		meta.GoPointers-- // one pointer less needed, because the converter will add it back
// 	}

// 	inIdent := inIdentByDir(direction, cIdent, goIdent)
// 	outIdent := outIdentByDir(direction, cIdent, goIdent)

// 	inType := inType(meta, direction)
// 	outType := outType(meta, direction)

// 	// primitive cases
// 	switch {
// 	case meta.CGoType() == "*C.void" && param.Closure != nil && direction == ConvertCToGo:
// 		fallthrough
// 	case meta.CGoBaseType == "C.gpointer" && param.Closure != nil && direction == ConvertCToGo:
// 		// we are converting a user_data param that should yield the function from a gbox, this is handled in the callback generator
// 		return NoopConverter{
// 			dir:    direction,
// 			inName: inIdent,
// 			inType: inType,
// 		}
// 	case meta.CGoBaseType == "C.gboolean":
// 		return &BooleanConverter{
// 			Direction: direction,
// 			InIdent:   inIdent,
// 			InTyp:     inType,
// 			OutIdent:  outIdent,
// 			OutTyp:    outType,
// 		}
// 	case meta.CGoBaseType == "C.guint" ||
// 		meta.CGoBaseType == "C.guint8" ||
// 		meta.CGoBaseType == "C.guint16" ||
// 		meta.CGoBaseType == "C.guint32" ||
// 		meta.CGoBaseType == "C.guint64" ||
// 		meta.CGoBaseType == "C.gint" ||
// 		meta.CGoBaseType == "C.gint8" ||
// 		meta.CGoBaseType == "C.gint16" ||
// 		meta.CGoBaseType == "C.gint32" ||
// 		meta.CGoBaseType == "C.gint64" ||
// 		meta.CGoBaseType == "C.gsize" ||
// 		meta.CGoBaseType == "C.gssize" ||
// 		meta.CGoBaseType == "C.gdouble" ||
// 		meta.CGoBaseType == "C.gfloat" ||
// 		meta.CGoBaseType == "C.glong" ||
// 		meta.CGoBaseType == "C.gulong" ||
// 		meta.CGoBaseType == "C.int" ||
// 		meta.CGoBaseType == "C.uint":
// 		return &PrimitiveConverter{
// 			Direction: direction,
// 			InIdent:   inIdent,
// 			InTyp:     inType,
// 			OutIdent:  outIdent,
// 			OutTyp:    outType,
// 		}
// 	case meta.GoType() == "string":
// 		return &StringConverter{
// 			Direction: direction,
// 			InIdent:   inIdent,
// 			InTyp:     inType,
// 			OutIdent:  outIdent,
// 			OutTyp:    outType,
// 		}
// 	}

// 	if meta.GoBaseType == "gobject.Value" {
// 		return nil // TODO: use coreglib for this
// 	}

// 	if meta.GoBaseType == "unsafe.Pointer" || meta.GoBaseType == "uintptr" {
// 		return nil // will not handle unsafe pointers if not a userdata arg
// 	}

// 	// cases that might be in other packages generated
// 	switch girType := meta.GirType.(type) {
// 	case *gir.Record:
// 		return NewRecordConverter(
// 			ctx,
// 			girType,
// 			direction,
// 			inIdent,
// 			inType,
// 			outIdent,
// 			outType,
// 		)
// 	case *gir.Class:
// 		return &ClassConverter{
// 			Direction: direction,
// 			InIdent:   inIdent,
// 			InTyp:     inType,
// 			OutIdent:  outIdent,
// 			OutTyp:    outType,
// 		}
// 	case *gir.Alias:
// 		return &AliasConverter{
// 			Direction: direction,
// 			InIdent:   inIdent,
// 			InTyp:     inType,
// 			OutIdent:  outIdent,
// 			OutTyp:    outType,
// 		}
// 	case *gir.Enum:
// 		return &EnumConverter{
// 			Direction: direction,
// 			InIdent:   inIdent,
// 			InTyp:     inType,
// 			OutIdent:  outIdent,
// 			OutTyp:    outType,
// 		}
// 	case *gir.Bitfield:
// 		return &BitfieldConverter{
// 			Direction: direction,
// 			InIdent:   inIdent,
// 			InTyp:     inType,
// 			OutIdent:  outIdent,
// 			OutTyp:    outType,
// 		}
// 	case *gir.Callback:
// 		return &CallbackConverter{
// 			Direction: direction,
// 			InIdent:   inIdent,
// 			InTyp:     inType,
// 			OutIdent:  outIdent,
// 			OutTyp:    outType,
// 		}
// 	default:
// 		panic("unhandled case for conversion: " + meta.CGoBaseType + " " + meta.GoBaseType)
// 	}
// }

// func inType(meta *typesystem.TypeMetadata, dir ConversionDirection) string {
// 	switch dir {
// 	case ConvertCToGo:
// 		return meta.CGoType()
// 	case ConvertGoToC:
// 		return meta.GoType()
// 	default:
// 		panic(fmt.Sprintf("unexpected value.ConversionDirection: %#v", dir))
// 	}
// }

// func outType(meta *typesystem.TypeMetadata, dir ConversionDirection) string {
// 	switch dir {
// 	case ConvertCToGo:
// 		return meta.GoType()
// 	case ConvertGoToC:
// 		return meta.CGoType()
// 	default:
// 		panic(fmt.Sprintf("unexpected value.ConversionDirection: %#v", dir))
// 	}
// }

// func inIdentByDir(dir ConversionDirection, cIndent, goIndent string) string {
// 	switch dir {
// 	case ConvertCToGo:
// 		return cIndent
// 	case ConvertGoToC:
// 		return goIndent
// 	default:
// 		panic(fmt.Sprintf("unexpected value.ConversionDirection: %#v", dir))
// 	}
// }

// func outIdentByDir(dir ConversionDirection, cIndent, goIndent string) string {
// 	switch dir {
// 	case ConvertCToGo:
// 		return goIndent
// 	case ConvertGoToC:
// 		return cIndent
// 	default:
// 		panic(fmt.Sprintf("unexpected value.ConversionDirection: %#v", dir))
// 	}
// }
