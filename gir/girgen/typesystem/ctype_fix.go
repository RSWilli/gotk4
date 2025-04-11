package typesystem

import "github.com/diamondburned/gotk4/gir"

type OverriddenCType interface {
	OriginalCGoType(pointers int) string
}

// cgoPrimitiveTypes contains edge cases for referencing C primitive types from
// CGo.
//
// See https://gist.github.com/zchee/b9c99695463d8902cd33.
var cgoPrimitiveTypes = map[string]string{
	"long long": "longlong",

	"unsigned char":      "uchar",
	"unsigned int":       "uint",
	"unsigned short":     "ushort",
	"unsigned long":      "ulong",
	"unsigned long long": "ulonglong",
}

func fixCType(resolved Type, requested gir.AnyType) Type {
	if requested.Type == nil {
		return resolved // may be array
	}

	req := requested.Type
	ctype := cleanCType(trimCTypePointers(req.CType))

	if goEquivalent, ok := cgoPrimitiveTypes[ctype]; ok {
		ctype = goEquivalent
	}

	if ctype == "" {
		return resolved // edge case, hopefully the type resolved by GIR is correct
	}

	if ctype == resolved.CType(0) {
		return resolved // no need to override
	}

	inter, ok := resolved.(ConvertibleType)
	if ok {
		return &OverriddenCTypeConvertible{
			Actual:  inter,
			Ctype:   ctype,
			Cgotype: "C." + ctype,
		}
	}

	castable, ok := resolved.(CastableType)

	if ok {
		return &OverriddenCTypeCastable{
			Actual:  castable,
			Ctype:   ctype,
			Cgotype: "C." + ctype,
		}
	}

	return &OverriddenCTypeSimple{
		Actual:  resolved,
		Ctype:   ctype,
		Cgotype: "C." + ctype,
	}
}
