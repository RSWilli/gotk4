package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

// Bitfield is always a wrapper type for int in go
type Bitfield struct {
	BaseType
	Doc

	Marshaler

	Members []*Member
}

func DeclareBitfield(e *env, v gir.Bitfield) *Bitfield {
	if !v.IsIntrospectable() {
		return nil
	}

	if e.skip(nil, v) {
		return nil
	}

	b := &Bitfield{
		BaseType: BaseType{
			GirName: v.Name,
			GoTyp:   strcases.PascalToGo(v.Name),
			CGoTyp:  "C." + v.CType,
			CTyp:    v.CType,
		},
		Marshaler: newDefaultMarshaler(v.GLibGetType),
		Doc:       NewDoc(&v.InfoAttrs, &v.InfoElements),
	}

	b.Members = GetMembers(e, b, v.Members)

	return b
}

// pointersAllowed implements Type.
func (a *Bitfield) pointersAllowed(pointers int) bool {
	return pointers == 0
}
