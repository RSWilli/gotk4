package typesystem

import (
	"strings"

	"github.com/diamondburned/gotk4/gir"
)

type Type interface {
	GIRName() string
	GoType() string
	CGoType() string
	CType() string
}

// ForeignType describes a type that must be imported from another Namespace
type ForeignType struct {
	SourceNamespace *Namespace
	Type
}

// UnderlyingType removes the [ForeignType] if there is one. This is useful for type assertions
// on the underlying types, e.g. for parent classes
func UnderlyingType(t Type) Type {
	switch v := t.(type) {
	case *ForeignType:
		return v.Type
	default:
		return t
	}
}

type baseType struct {
	girName string
	goType  string
	cGoType string
	cType   string
}

// GIRName implements Type.
func (b baseType) GIRName() string {
	return b.girName
}

// CGoType implements Type.
func (b baseType) CGoType() string {
	return b.cGoType
}

// CType implements Type.
func (b baseType) CType() string {
	return b.cType
}

// GoType implements Type.
func (b baseType) GoType() string {
	return b.goType
}

var _ Type = baseType{}

// WithPointers optionally wraps the found type in a PointerType if the pointercount does not match the
// ctype's pointer count specified in the gir type
func WithPointers(girType *gir.Type, t Type) Type {
	if t == nil {
		return nil
	}

	if girType.CType == t.CType() {
		return t // fast path
	}

	girPointers := strings.Count(girType.CType, "*")
	typPointers := strings.Count(t.CType(), "*")

	missing := girPointers - typPointers

	if missing < 0 {
		panic("too many pointers on type")
	}

	if missing == 0 {
		return t
	}

	return &PointerType{
		Pointers: strings.Repeat("*", missing),
		Base:     t,
	}
}

type PointerType struct {
	Pointers string
	Base     Type
}

// GIRName implements Type.
func (b PointerType) GIRName() string {
	return b.Base.GIRName()
}

// CGoType implements Type.
func (b PointerType) CGoType() string {
	return b.Pointers + b.Base.CGoType()
}

// CType implements Type.
func (b PointerType) CType() string {
	return b.Base.CType() + b.Pointers
}

// GoType implements Type.
func (b PointerType) GoType() string {
	return b.Pointers + b.Base.GoType()
}

var _ Type = PointerType{}
