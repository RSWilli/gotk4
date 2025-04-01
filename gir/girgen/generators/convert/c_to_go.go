package convert

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

func NewCToGoConverter(p *typesystem.Param) Converter {
	if p.Nullable {
		if p.Type == typesystem.Utf8 {
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
	if _, ok := p.Type.(*typesystem.Primitive); ok {
		return newCToGoPrimitiveConverter(p)
	}

	if _, ok := p.Type.(*typesystem.PointerType); ok {
		return newCToGoPointerConverter(p)
	}

	switch t := typesystem.UnderlyingType(p.Type).(type) {
	// TODO: handle non pointer cases here for records, aliases, bitflags, enum etc

	case *typesystem.Primitive:
		// needed for manual primitives like GType:
		return newCToGoPrimitiveConverter(p)
	case *typesystem.Array:
	case *typesystem.Bitfield, *typesystem.Enum:
		return &CToGoCastingConverter{Param: p}
	case *typesystem.Callback:
	case *typesystem.Class:
	case *typesystem.Container:
	case *typesystem.Interface:
	case *typesystem.Record:
	case *typesystem.Union:
	case *typesystem.Alias:
		// tricky sub cases here, depending on the aliased type
		return &UnimplementedConverter{Param: p}
	default:
		// case typesystem.BaseType:
		// case *typesystem.ForeignType:
		panic(fmt.Sprintf("unexpected typesystem.Type: %T %s", t, t.CType()))
	}

	return &UnimplementedConverter{
		Param: p,
	}
}

func newCToGoPrimitiveConverter(p *typesystem.Param) Converter {
	switch p.Type {
	case typesystem.Utf8:
		return &CToGoStringConverter{Param: p}
	case typesystem.Gboolean:
		return &CToGoBooleanConverter{Param: p}
	default:
		return &CToGoCastingConverter{Param: p}
	}
}

func newCToGoPointerConverter(pointerparam *typesystem.Param) Converter {
	if pointerparam.Type.(*typesystem.PointerType).Pointers != 1 {
		return &UnimplementedConverter{
			Param: pointerparam,
		}
	}

	// basetype may be a foreign type
	basetype := pointerparam.Type.(*typesystem.PointerType).Base

	switch typesystem.UnderlyingType(basetype).(type) {
	case *typesystem.Record:
		return &CToGoRecordPointerConverter{Param: pointerparam, Record: basetype}
	case *typesystem.Class:
		return &CToGoClassPointerConverter{Param: pointerparam, Class: basetype}
	}

	return &UnimplementedConverter{
		Param: pointerparam,
	}
}
