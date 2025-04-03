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

	AliasedType CouldBeForeign[Type]
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
		},
		Doc: NewDoc(&v.InfoAttrs, &v.InfoElements),
		gir: v,
	}

	return a
}

func (a *Alias) resolve(e *env) bool {
	e = e.sub("alias", a.gir.Type)

	ns, subtype := e.findType(&a.gir.Type)

	if subtype == nil {
		return false
	}

	a.AliasedType = CouldBeForeign[Type]{
		Namespace: ns,
		Type:      subtype,
	}

	return true
}

// pointersAllowed implements Type.
func (a *Alias) pointersAllowed(pointers int) bool {
	return pointers == 0 && a.AliasedType.Type.pointersAllowed(0)
}
