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

	if req.CType == "" {
		return resolved // edge case, hopefully the type resolved by GIR is correct
	}

	ctypePointers := CountCTypePointers(req.CType)
	cleanType := cleanCType(req.CType)
	baseCtype := trimCTypePointers(cleanType)

	if goEquivalent, ok := cgoPrimitiveTypes[baseCtype]; ok {
		baseCtype = goEquivalent
	}

	if cleanType == resolved.CType(ctypePointers) {
		return resolved // no need to override
	}

	inter, ok := resolved.(ConvertibleType)
	if ok {
		return &OverriddenCTypeConvertible{
			Actual:  inter,
			Ctype:   baseCtype,
			Cgotype: "C." + baseCtype,
		}
	}

	castable, ok := resolved.(CastableType)

	if ok {
		return &OverriddenCTypeCastable{
			Actual:  castable,
			Ctype:   baseCtype,
			Cgotype: "C." + baseCtype,
		}
	}

	if resolved == Utf8 || resolved == Filename {
		return &OverriddenCTypeString{
			Actual:  resolved.(*StringPrimitive),
			Ctype:   baseCtype + GetPointers(ctypePointers),
			Cgotype: GetPointers(ctypePointers) + "C." + baseCtype,
		}
	}

	return &OverriddenCTypeSimple{
		Actual:  resolved,
		Ctype:   baseCtype,
		Cgotype: "C." + baseCtype,
	}
}
