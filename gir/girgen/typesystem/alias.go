package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type Alias struct {
	BaseType
	Doc

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
		BaseType: BaseType{
			GirName: v.Name,
			GoTyp:   strcases.PascalToGo(v.Name),
			CGoTyp:  "C." + v.CType,
			CTyp:    v.CType,

			// has no get type, but will be marshaled by calling the subtype marshaler.
			GlibGetTypeFn: "",
		},
		AliasedType: subtype,
		Doc:         NewDoc(&v.InfoAttrs, &v.InfoElements),
	}
}
