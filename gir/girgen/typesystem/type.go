package typesystem

import (
	"log"
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

func CountPointers(ctype string) int {
	pointers := strings.Count(ctype, "*")

	if strings.HasPrefix(ctype, "gpointer") || strings.HasPrefix(ctype, "gconstpointer") {
		return pointers + 1
	}

	return pointers
}

// WithPointers optionally wraps the found type in a PointerType if the pointercount does not match the
// ctype's pointer count specified in the gir type
func WithPointers(girType *gir.Type, t Type) Type {
	if t == nil {
		panic("cannot wrap nil type with pointers")
	}

	if girType.CType == "" {
		return t
	}

	cleanedCType := cleanCType(girType.CType)

	if cleanedCType == t.CType() {
		return t // fast path
	}

	girPointers := CountPointers(cleanedCType)
	typPointers := CountPointers(t.CType())

	missing := girPointers - typPointers

	if missing < 0 {
		log.Printf("too many pointers on type %s vs %s", cleanedCType, t.CType())
		return nil
	}

	if missing == 0 {
		return t
	}

	// log.Printf("adding missing pointers to %s to match %s\n", t.CType(), cleanedCType)

	return &PointerType{
		Pointers: missing,
		Base:     t,
	}
}

type PointerType struct {
	Pointers int
	Base     Type
}

// GIRName implements Type.
func (b PointerType) GIRName() string {
	return b.Base.GIRName()
}

// CGoType implements Type.
func (b PointerType) CGoType() string {
	return strings.Repeat("*", b.Pointers) + b.Base.CGoType()
}

// CType implements Type.
func (b PointerType) CType() string {
	return b.Base.CType() + strings.Repeat("*", b.Pointers)
}

// GoType implements Type.
func (b PointerType) GoType() string {
	return strings.Repeat("*", b.Pointers) + b.Base.GoType()
}

var _ Type = PointerType{}
