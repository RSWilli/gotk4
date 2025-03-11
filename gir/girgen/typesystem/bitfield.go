package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

// Bitfield is always a wrapper type for int in go
type Bitfield struct {
	baseType
	Doc Doc

	Members []*Member
}

func DeclareBitfield(ns *Namespace, v gir.Bitfield) *Bitfield {
	if !v.IsIntrospectable() {
		return nil
	}

	return &Bitfield{
		baseType: baseType{
			girName: v.Name,
			goType:  strcases.PascalToGo(v.Name),
			cGoType: "C." + v.CType,
			cType:   v.CType,
		},
		Doc:     NewDoc(&v.InfoAttrs, &v.InfoElements),
		Members: GetMembers(v.Members),
	}
}
