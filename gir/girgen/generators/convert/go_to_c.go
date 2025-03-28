package convert

import (
	"fmt"

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

	if _, ok := p.Type.(*typesystem.Primitive); ok {
		return newGoToCPrimitiveConverter(p)
	}

	if _, ok := p.Type.(*typesystem.PointerType); ok {
		return newGoToCPointerConverter(p)
	}

	switch t := typesystem.UnderlyingType(p.Type).(type) {
	// TODO: handle non pointer cases here for records, aliases, bitflags, enum etc

	case *typesystem.Primitive:
		// needed for manual primitives like GType:
		return newGoToCPrimitiveConverter(p)
	case *typesystem.Array:
	case *typesystem.Bitfield, *typesystem.Enum:
		return &GoToCCastingConverter{Param: p}
	case *typesystem.Callback:
	case *typesystem.Class:
	case *typesystem.Container:
	case *typesystem.Interface:
	case *typesystem.PointerType:
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

func newGoToCPrimitiveConverter(p *typesystem.Param) Converter {
	switch p.Type {
	case typesystem.Utf8:
		return &UnimplementedConverter{Param: p}
	case typesystem.Gboolean:
		return &GoToCBooleanConverter{Param: p}
	default:
		return &GoToCCastingConverter{Param: p}
	}
}

// newGoToCPointerConverter must be called on a PointerType Param.
func newGoToCPointerConverter(pointerparam *typesystem.Param) Converter {
	if pointerparam.Type.(*typesystem.PointerType).Pointers != 1 {
		return &UnimplementedConverter{
			Param: pointerparam,
		}
	}

	// basetype may be a foreign type
	basetype := pointerparam.Type.(*typesystem.PointerType).Base

	switch typesystem.UnderlyingType(basetype).(type) {
	case *typesystem.Record:
		return &GoToCRecordPointerConverter{Param: pointerparam, Record: basetype}
	}

	return &UnimplementedConverter{
		Param: pointerparam,
	}
}
