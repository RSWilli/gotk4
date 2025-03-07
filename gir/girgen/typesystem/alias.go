package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type Alias struct {
	baseType
	Doc Doc

	AliasedType Type
}

func NewAlias(ns *Namespace, v gir.Alias) *Alias {
	if !v.IsIntrospectable() {
		return nil
	}

	subtype := ns.findType(&v.Type)

	if subtype == nil {
		return nil
	}

	return &Alias{
		baseType: baseType{
			girName: v.Name,
			goType:  strcases.PascalToGo(v.Name),
			cGoType: "C." + v.Name,
			cType:   v.Name,
		},
		AliasedType: subtype,
		Doc:         NewDoc(&v.InfoAttrs, &v.InfoElements),
	}
}
