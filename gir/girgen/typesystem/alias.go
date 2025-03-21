package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type Alias struct {
	BaseType
	Doc

	// gir is used to resolve the aliased type after it has been declared
	gir gir.Alias

	AliasedType Type
}

func DeclareAlias(e *env, v gir.Alias) *Alias {
	if !v.IsIntrospectable() {
		return nil
	}

	if e.skip(nil, v) {
		return nil
	}

	a := &Alias{
		BaseType: BaseType{
			GirName: v.Name,
			GoTyp:   strcases.PascalToGo(v.Name),
			CGoTyp:  "C." + v.CType,
			CTyp:    v.CType,

			// has no get type, but will be marshaled by calling the subtype marshaler.
			GlibGetTypeFn: "",
		},
		AliasedType: nil, // lazily set
		Doc:         NewDoc(&v.InfoAttrs, &v.InfoElements),
		gir:         v,
	}

	return a
}

func (a *Alias) resolve(e *env) bool {
	subtype := e.findType(&a.gir.Type)

	if subtype == nil {
		return false
	}

	a.AliasedType = subtype

	return true
}
