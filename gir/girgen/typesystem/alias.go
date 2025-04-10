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

func (a *Alias) resolve(e *env) resolvedState {
	e = e.sub("alias", a.gir.Name)

	ns, subtype := e.findType(&a.gir.Type)

	if subtype == nil {
		return maybeResolvable
	}

	if subtype == Void {
		return notResolvable
	}

	if TypeRequiresPointer(subtype) {
		e.logger.Warn("skipping alias that requires pointers")
		return notResolvable
	}

	a.AliasedType = CouldBeForeign[Type]{
		Namespace: ns,
		Type:      subtype,
	}

	return okResolved
}
