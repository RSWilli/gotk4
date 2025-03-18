package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type Enum struct {
	BaseType
	Doc

	Members []*Member
}

func DeclareEnum(e *env, v gir.Enum) *Enum {
	if e.skipType(v) {
		return nil
	}

	return &Enum{
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
