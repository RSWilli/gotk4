package typesystem

import (
	"log"
	"slices"
	"strings"

	"github.com/diamondburned/gotk4/gir"
)

func CountPointers(ctype string) int {
	pointers := strings.Count(ctype, "*")

	if strings.HasPrefix(ctype, "gpointer") || strings.HasPrefix(ctype, "gconstpointer") {
		pointers += 1
	}

	return pointers
}

// WithPointers optionally wraps the found type in a PointerType if the pointercount does not match the
// ctype's pointer count specified in the gir type
func WithPointers(girType *gir.Type, t Type) Type {
	if t == nil {
		panic("cannot wrap nil type with pointers")
	}

	if t == Gpointer {
		return t // there is an unknown amount of internal pointers, we don't need another one
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

var goInternalPointers = []string{
	"error",
}

// GoType implements Type.
func (b PointerType) GoType() string {
	subtype := b.Base.GoType()

	if slices.Contains(goInternalPointers, subtype) {
		// if the type already contains a pointer the go type can have one less pointer
		return strings.Repeat("*", b.Pointers-1) + subtype
	}

	return strings.Repeat("*", b.Pointers) + subtype
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

func IncreasePointers(t Type, n int) Type {
	switch t := t.(type) {
	case *PointerType:
		return &PointerType{
			Pointers: t.Pointers + n,
			Base:     t.Base,
		}
	default:
		return &PointerType{
			Pointers: n,
			Base:     t,
		}
	}
}

func DecreasePointers(t Type, n int) Type {
	switch t := t.(type) {
	case *PointerType:
		switch t.Pointers {
		case 0:
			panic("can't decrease pointers further")
		case 1:
			return t.Base
		default:
			return &PointerType{
				Pointers: t.Pointers - n,
				Base:     t.Base,
			}
		}
	default:
		panic("received non pointer type in decrease pointers")
	}
}
