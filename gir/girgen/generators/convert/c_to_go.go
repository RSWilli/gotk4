package convert

import (
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

func NewCToGoConverter(p *typesystem.Param) Converter {
	if p.CallerAllocates {
		return &UnimplementedConverter{
			Param: p,
		}
	}

	if p.Nullable {
		if p.Type.Type == typesystem.Utf8 {
			return &CToGoNullableStringConverter{
				Param: p,
				SubConverter: &CToGoStringConverter{
					Param: p,
				},
			}
		}
		return &CToGoNullableConverter{
			Param:        p,
			SubConverter: newCToGoBasicConverter(p),
		}
	}

	return newCToGoBasicConverter(p)
}

func newCToGoBasicConverter(p *typesystem.Param) Converter {
	switch p.Type.Type {
	case typesystem.Utf8, typesystem.Filename:
		return &CToGoStringConverter{Param: p}
	case typesystem.Gboolean:
		return &CToGoBooleanConverter{Param: p}
	}

	switch p.Type.Type.(type) {
	case *typesystem.CastablePrimitive, *typesystem.Bitfield, *typesystem.Enum:
		return &CToGoCastingConverter{
			Param: p,
		}
	case *typesystem.Alias:
		return newCToGoAliasedConverter(p)
	}

	if conv, ok := p.Type.Type.(typesystem.ConvertFromGlibFullType); p.TransferOwnership == typesystem.TransferFull && ok {
		return &CToGoConvertibleConverter{
			Param:       p,
			ConvertFunc: conv.GoUnsafeFromGlibFullFunction(),
		}
	}

	if conv, ok := p.Type.Type.(typesystem.ConvertFromGlibNoneType); p.TransferOwnership == typesystem.TransferNone && ok {
		return &CToGoConvertibleConverter{
			Param:       p,
			ConvertFunc: conv.GoUnsafeFromGlibNoneFunction(),
		}
	}

	if conv, ok := p.Type.Type.(typesystem.ConvertFromGlibBorrowType); p.TransferOwnership == typesystem.TransferBorrow && ok {
		return &CToGoConvertibleConverter{
			Param:       p,
			ConvertFunc: conv.GoUnsafeFromGlibBorrowFunction(),
		}
	}

	return &UnimplementedConverter{Param: p}
}

// tricky sub cases here, depending on the aliased type
func newCToGoAliasedConverter(p *typesystem.Param) Converter {
	subtype := p.Type.Type.(*typesystem.Alias).AliasedType

	if _, ok := subtype.Type.(*typesystem.CastablePrimitive); ok {
		return &AliasConverter{
			SubConverter: &CToGoCastingConverter{
				Param: p,
			},
		}
	}

	return &UnimplementedConverter{Param: p}
}
