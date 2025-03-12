package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type Enum struct {
	baseType
	Doc Doc

	Members []*Member
}

func DeclareEnum(ctx context, v gir.Enum) *Enum {
	if ctx.skipType(v) {
		return nil
	}

	return &Enum{
		baseType: baseType{
			girName: v.Name,
			goType:  strcases.PascalToGo(v.Name),
			cGoType: "C." + v.CType,
			cType:   v.CType,
		},
		Doc:     NewDoc(&v.InfoAttrs, &v.InfoElements),
		Members: GetMembers(v.Members),
	}
}
