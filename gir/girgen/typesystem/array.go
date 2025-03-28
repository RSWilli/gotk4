package typesystem

import (
	"fmt"

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
	return a.Inner.CGoType()
}

// CType implements Type.
func (a *Array) CType() string {
	if a.cTypeOverride != "" {
		return a.cTypeOverride
	}
	return a.Inner.CType()
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

// getArrayType resolves the array type in the current env
func (e *env) getArrayType(arr *gir.Array) *Array {
	if arr.Length == nil && arr.FixedSize == 0 && !arr.IsZeroTerminated() {
		// this is an unbounded array, which requires some unsafe preconditions not
		// documented in GIR, must be handled manually
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
		e.logger.Debug("FIXME: array type is nil, needs special handling", "ctype", arr.CType)
		return nil
	}

	originalPointers := CountPointers(arr.CType)

	if arr.Type.Name == "utf8" || originalPointers > 1 {
		// this is a higher dimensional array, which needs to be implemented manually
		e.logger.Info("skipping high dimensional array type", "name", arr.Type.Name, "ctype", arr.CType)
		return nil
	}

	inner := e.findTypeByGIRName(arr.Type.Name)
	if inner == nil {
		e.logger.Warn("could not find array inner type", "name", arr.Type.Name)
		return nil
	}

	if _, ok := inner.(*PointerType); !ok {
		// inner type is always a pointer, because we need to do pointer arithmetics
		inner = IncreasePointers(inner, 1)
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

	return array
}
