package typesystem

import (
	"fmt"
)

type Type interface {
	GIRName() string
	GoType() string
	CGoType() string
	CType() string

	GLibGetType() string
	MarshalFuncName() string
}

var _ Type = (*PointerType)(nil)

func Is(t, other Type) bool {
	return UnderlyingType(t) == other
}

// UnderlyingType removes the [ForeignType] or [PointerType] if there is one. This is useful for type assertions
// on the underlying types, e.g. for parent classes
func UnderlyingType(t Type) Type {
	switch v := t.(type) {
	case *ForeignType:
		return UnderlyingType(v.Type)
	case *PointerType:
		return UnderlyingType(v.Base)
	default:
		return t
	}
}

type BaseType struct {
	GirName string
	GoTyp   string
	CGoTyp  string
	CTyp    string

	GlibGetTypeFn string
}

// GIRName implements Type.
func (b BaseType) GIRName() string {
	return b.GirName
}

// CGoType implements Type.
func (b BaseType) CGoType() string {
	return b.CGoTyp
}

// CType implements Type.
func (b BaseType) CType() string {
	return b.CTyp
}

// GoType implements Type.
func (b BaseType) GoType() string {
	return b.GoTyp
}

func (b BaseType) GLibGetType() string {
	return b.GlibGetTypeFn
}

func (b BaseType) MarshalFuncName() string {
	return fmt.Sprintf("marshal%s", b.GoTyp)
}

var _ Type = BaseType{}
