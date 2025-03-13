package typesystem

import (
	"fmt"
	"log"

	"github.com/diamondburned/gotk4/gir"
)

// Array is the type for array params. It has an inner type and may reference another [Param] for its length.
type Array struct {
	cTypeOverride   string
	cGoTypeOverride string
	goTypeOverride  string

	Inner          Type
	Length         *Param // length is filled out by [NewParameters]
	ZeroTerminated bool
	FixedSize      int
}

// GLibGetType implements Type.
func (a *Array) GLibGetType() string {
	panic("unimplemented")
}

// MarshalFuncName implements Type.
func (a *Array) MarshalFuncName() string {
	panic("unimplemented")
}

var _ Type = (*Array)(nil)

// CGoType implements Type.
func (a *Array) CGoType() string {
	if a.cGoTypeOverride != "" {
		return a.cGoTypeOverride
	}
	return fmt.Sprintf("*%s", a.Inner.CGoType())
}

// CType implements Type.
func (a *Array) CType() string {
	if a.cTypeOverride != "" {
		return a.cTypeOverride
	}
	return fmt.Sprintf("%s*", a.Inner.CType())
}

// GIRName implements Type.
func (a *Array) GIRName() string {
	if a.Inner == nil {
		return "array[unknown]"
	}
	return fmt.Sprintf("array[%s]", a.Inner.GIRName())
}

// GoType implements Type.
func (a *Array) GoType() string {
	if a.goTypeOverride != "" {
		return a.goTypeOverride
	}
	return fmt.Sprintf("[]%s", a.Inner.GoType())
}

func getArrayType(ns *Namespace, arr *gir.Array) *Array {
	if arr.Length == nil && arr.FixedSize == 0 && !arr.IsZeroTerminated() {
		// this is an unbounded array, which requires some unsafe preconditions not
		// documented in GIR, FIXME: can we even handle this?
		// log.Println("skipping unbounded array type")
		return nil
	}

	if arr.CType == "" {
		// this is true for some dummy fields, we just ignore them
		return nil
	}

	if arr.CType == "gpointer" || arr.CType == "gconstpointer" || arr.CType == "void*" {
		// this represents a bytes array, e.g. for g_bytes_get_data
		return &Array{
			cTypeOverride:   arr.CType,
			cGoTypeOverride: "C." + arr.CType,
			Inner:           prim("...", typeInvalid, typeInvalid, "byte"),
			FixedSize:       arr.FixedSize,
			ZeroTerminated:  arr.IsZeroTerminated(),
		}
	}

	cleanedCtype := cleanCType(arr.CType)

	if cleanedCtype == "gchar*" {
		// this is a string where the length is somehow given
		return &Array{
			cTypeOverride:   arr.CType, // may contain "const"
			cGoTypeOverride: "*C.gchar",
			goTypeOverride:  "string",
			Inner:           nil,
			FixedSize:       arr.FixedSize,
			ZeroTerminated:  arr.IsZeroTerminated(),
		}
	}

	if arr.Type == nil {
		log.Printf("FIXME: array type is nil, needs special handling for Ctype %s\n", arr.CType)
		return nil
	}

	originalPointers := CountPointers(arr.CType)

	if arr.Type.Name == "utf8" || originalPointers > 1 {
		// this is a higher dimensional array, and there is no
		// way in GIR that specifies the length of the inner array
		// log.Printf("skipping %s high dimensional array type %s\n", arr.Type.Name, arr.CType)
		return nil
	}

	inner := ns.findTypeByGIRName(arr.Type.Name)
	if inner == nil {
		log.Printf("could not find array inner type %s", arr.Type.Name)
		return nil
	}

	array := &Array{
		Inner:          inner,
		Length:         nil, // will be set by params if relevant
		ZeroTerminated: arr.IsZeroTerminated(),
		FixedSize:      arr.FixedSize,
	}

	typePointers := CountPointers(array.CType())

	missingPointers := originalPointers - typePointers

	if missingPointers < 0 {
		panic("too many pointers on type")
	}

	if missingPointers > 0 {
		array.Inner = &PointerType{
			Pointers: missingPointers,
			Base:     array.Inner,
		}
	}

	// if array.CType() != cleanedCtype {
	// 	log.Printf("array ctype wrong, expected %s got %s", cleanedCtype, array.CType())
	// }

	return array
}
