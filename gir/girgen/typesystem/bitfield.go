package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

// Bitfield is always a wrapper type for int in go
type Bitfield struct {
	BaseType
	Doc

	Members []*Member
}

func DeclareBitfield(e *env, v gir.Bitfield) *Bitfield {
	if e.skipType(v) {
		return nil
	}

	return &Bitfield{
		BaseType: BaseType{
			GirName: v.Name,
			GoTyp:   strcases.PascalToGo(v.Name),
			CGoTyp:  "C." + v.CType,
			CTyp:    v.CType,

			GlibGetTypeFn: v.GLibGetType,
		},
		Doc:     NewDoc(&v.InfoAttrs, &v.InfoElements),
		Members: GetMembers(v.Members),
	}
}
