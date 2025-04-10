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
	e = e.sub("bitfield", v.CType)

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
		Marshaler: e.newDefaultMarshaler(v.GLibGetType, v.Name),
		Doc:       NewDoc(&v.InfoAttrs, &v.InfoElements),
	}

	b.Members = GetMembers(e, b, v.Members)

	return b
}

func (a *Bitfield) maxPointersAllowed() int {
	return 0
}
