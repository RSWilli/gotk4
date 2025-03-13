package typesystem

import (
	"fmt"
	"log"
	"strings"

	"github.com/diamondburned/gotk4/gir"
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

// ForeignType describes a type that must be imported from another Namespace
type ForeignType struct {
	SourceNamespace *Namespace
	Type
}

// GoType implements Type.
func (b *ForeignType) GoType() string {
	if _, ok := b.Type.(*PointerType); ok {
		panic("foreign pointer type")
	}
	return fmt.Sprintf("%s.%s", b.SourceNamespace.GoName, b.Type.GoType())
}

func IsForeignType(t Type) bool {
	switch t := t.(type) {
	case *ForeignType:
		return true
	case *PointerType:
		return IsForeignType(t.Base)
	default:
		return false
	}
}

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

// GLibGetType implements Type.
func (b *PointerType) GLibGetType() string {
	return b.Base.GLibGetType()
}

// MarshalFuncName implements Type.
func (b *PointerType) MarshalFuncName() string {
	return b.Base.MarshalFuncName()
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

func Pointers(t Type) int {
	switch t := t.(type) {
	case *ForeignType:
		return Pointers(t.Type)
	case *PointerType:
		return t.Pointers
	default:
		return 0
	}
}
