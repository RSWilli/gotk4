package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type Enum struct {
	BaseType
	Doc
	Marshaler

	Members []*Member
}

func DeclareEnum(e *env, v gir.Enum) *Enum {
	if !v.IsIntrospectable() {
		return nil
	}

	if e.skip(nil, v) {
		return nil
	}

	enum := &Enum{
		BaseType: BaseType{
			GirName: v.Name,
			GoTyp:   strcases.PascalToGo(v.Name),
			CGoTyp:  "C." + v.CType,
			CTyp:    v.CType,
		},
		Marshaler: newDefaultMarshaler(v.GLibGetType),
		Doc:       NewDoc(&v.InfoAttrs, &v.InfoElements),
	}

	enum.Members = GetMembers(e, enum, v.Members)

	return enum
}

// pointersAllowed implements Type.
func (a *Enum) pointersAllowed(pointers int) bool {
	return pointers == 0
}
