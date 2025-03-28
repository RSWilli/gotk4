package typesystem

import (
	"fmt"
	"slices"
	"strings"
)

// ForeignType describes a type that must be imported from another Namespace
type ForeignType struct {
	SourceNamespace *Namespace
	// Type Must not be a [ForeignType] nor a [PointerType]
	Type
}

// mkForeign wraps the given type in a [ForeignType] if the type is not already wrapped
func mkForeign(ns *Namespace, t Type) Type {
	switch t := t.(type) {
	case nil:
		return nil
	case *ForeignType:
		panic("trying to change foreign namespace")
	case *PointerType:
		// pointertypes must wrap foreign types
		return &PointerType{
			Pointers: t.Pointers,
			Base: &ForeignType{
				SourceNamespace: ns,
				Type:            t.Base,
			},
		}
	default:
		return &ForeignType{
			SourceNamespace: ns,
			Type:            t,
		}
	}
}

var nonNamespacedGoTypes = []string{
	"error",
}

// GoType implements Type.
func (b *ForeignType) GoType() string {
	subgotype := b.Type.GoType()

	if strings.Contains(subgotype, ".") {
		// already namespaced
		return subgotype
	}

	if slices.Contains(nonNamespacedGoTypes, subgotype) {
		return subgotype
	}

	return b.AddForeignNamespace(b.Type.GoType())
}

func (b *ForeignType) AddForeignNamespace(str string) string {
	if _, ok := b.Type.(*PointerType); ok {
		panic("foreign pointer type")
	}
	return fmt.Sprintf("%s.%s", b.SourceNamespace.GoName, str)
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
