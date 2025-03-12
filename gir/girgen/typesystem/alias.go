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

func DeclareAlias(ns context, v gir.Alias) *Alias {
	if ns.skipType(v) {
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
			cGoType: "C." + v.CType,
			cType:   v.CType,
		},
		AliasedType: subtype,
		Doc:         NewDoc(&v.InfoAttrs, &v.InfoElements),
	}
}
