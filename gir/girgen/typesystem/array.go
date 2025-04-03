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

	Inner         CouldBeForeign[Type]
	InnerPointers int

	Length         *Param // length is filled out by [NewParameters]
	ZeroTerminated bool
	FixedSize      int
}

var _ Type = (*Array)(nil)

func (a *Array) CGoType(pointers int) string {
	return "array"
}

// CType implements Type.
func (a *Array) CType(pointers int) string {
	return "array"
}

// GoType implements Type.
func (a *Array) GoType(pointers int) string {
	return "array"
}

// GIRName implements Type.
func (a *Array) GIRName() string {
	if a.Inner.Type == nil {
		return "array[unknown]"
	}
	return fmt.Sprintf("array[%s]", a.Inner.Type.GIRName())
}

// pointersAllowed implements Type.
func (a *Array) pointersAllowed(pointers int) bool {
	return true //pointers == 0 && (a.Inner.Type == nil || a.Inner.Type.pointersAllowed(a.InnerPointers))
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
			Inner: CouldBeForeign[Type]{
				Type: prim("...", typeInvalid, typeInvalid, "byte"),
			},
			FixedSize:      arr.FixedSize,
			ZeroTerminated: arr.IsZeroTerminated(),
		}
	}

	cleanedCtype := cleanCType(arr.CType)

	if cleanedCtype == "gchar*" {
		// this is a string where the length is somehow given
		return &Array{
			cTypeOverride:   arr.CType, // may contain "const"
			cGoTypeOverride: "*C.gchar",
			goTypeOverride:  "string",
			Inner:           CouldBeForeign[Type]{}, // no inner type
			FixedSize:       arr.FixedSize,
			ZeroTerminated:  arr.IsZeroTerminated(),
		}
	}

	if arr.Type == nil {
		e.logger.Debug("FIXME: array type is nil, needs special handling", "ctype", arr.CType)
		return nil
	}

	ns, inner := e.findTypeByGIRName(arr.Type.Name)
	if inner == nil {
		e.logger.Warn("could not find array inner type", "name", arr.Type.Name)
		return nil
	}

	array := &Array{
		Inner: CouldBeForeign[Type]{
			Namespace: ns,
			Type:      inner,
		},
		Length:         nil, // will be set by params if relevant
		ZeroTerminated: arr.IsZeroTerminated(),
		FixedSize:      arr.FixedSize,
	}

	return array
}
