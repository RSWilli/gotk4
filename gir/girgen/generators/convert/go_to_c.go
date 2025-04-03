package convert

import (
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

func NewGoToCConverter(p *typesystem.Param) Converter {
	if p.Implicit {
		return &ImplicitConverter{
			Param: p,
		}
	}
	if p.Skip {
		return &SkippedConverter{
			Param: p,
		}
	}

	if p.CallerAllocates {
		return &UnimplementedConverter{
			Param: p,
		}
	}

	if p.Direction == "out" {
		return &UnimplementedConverter{
			Param: p,
		}
	}

	if p.Nullable {
		if p.Type.Type == typesystem.Utf8 {
			return &GoToCNullableStringConverter{
				Param: p,
				SubConverter: &GoToCStringConverter{
					Param: p,
				},
			}
		}
		return &GoToCNullableConverter{
			Param:        p,
			SubConverter: newGoToCBasicConverter(p),
		}
	}

	return newGoToCBasicConverter(p)
}

func newGoToCBasicConverter(p *typesystem.Param) Converter {
	switch p.Type.Type {
	case typesystem.Utf8, typesystem.Filename:
		return &GoToCStringConverter{Param: p}
	case typesystem.Gboolean:
		return &GoToCBooleanConverter{Param: p}
	}

	switch p.Type.Type.(type) {
	case *typesystem.CastablePrimitive, *typesystem.Bitfield, *typesystem.Enum:
		return &GoToCCastingConverter{
			Param: p,
		}
	case *typesystem.Callback:
		return &GoToCCallbackConverter{
			Param: p,
		}
	case *typesystem.Alias:
		return newGoToCAliasedConverter(p)
	}

	if conv, ok := p.Type.Type.(typesystem.ConvertToGlibFullType); p.TransferOwnership == typesystem.TransferFull && ok {
		return &GoToCConvertibleConverter{
			Param:       p,
			ConvertFunc: conv.GoUnsafeToGlibFullFunction(),
		}
	}

	if conv, ok := p.Type.Type.(typesystem.ConvertToGlibNoneType); p.TransferOwnership == typesystem.TransferNone && ok {
		return &GoToCConvertibleConverter{
			Param:       p,
			ConvertFunc: conv.GoUnsafeToGlibNoneFunction(),
		}
	}

	return &UnimplementedConverter{Param: p}
}

// tricky sub cases here, depending on the aliased type
func newGoToCAliasedConverter(p *typesystem.Param) Converter {
	subtype := p.Type.Type.(*typesystem.Alias).AliasedType

	if _, ok := subtype.Type.(*typesystem.CastablePrimitive); ok {
		return &AliasConverter{
			SubConverter: &GoToCCastingConverter{
				Param: p,
			},
		}
	}

	return &UnimplementedConverter{Param: p}
}
