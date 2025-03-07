package typesystem

import "github.com/diamondburned/gotk4/gir"

// Array is the type for array params. It has an inner type and may reference another [Param] for its length.
type Array struct {
	Inner          Type
	Length         *Param
	ZeroTerminated bool
	FixedSize      int
}

// CGoType implements Type.
func (a *Array) CGoType() string {
	panic("unimplemented")
}

// CType implements Type.
func (a *Array) CType() string {
	panic("unimplemented")
}

// GIRName implements Type.
func (a *Array) GIRName() string {
	panic("unimplemented")
}

// GoType implements Type.
func (a *Array) GoType() string {
	panic("unimplemented")
}

func getArrayType(ns *Namespace, arr *gir.Array) *Array {
	if !arr.Introspectable {
		return nil
	}

	inner := ns.findType(arr.Type)

	if inner == nil {
		return nil
	}

	return &Array{
		Inner:          inner,
		Length:         nil, // will be set by params if relevant
		ZeroTerminated: arr.IsZeroTerminated(),
		FixedSize:      arr.FixedSize,
	}
}
