package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type Enum struct {
	BaseType
	Doc
	Marshaler

	Members Members
}

func DeclareEnum(e *env, v gir.Enum) *Enum {
	e = e.sub("bitfield", v.CType)

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
		Marshaler: e.newDefaultMarshaler(v.GLibGetType, v.Name),
		Doc:       NewDoc(&v.InfoAttrs, &v.InfoElements),
	}

	enum.Members = GetMembers(e, enum, v.Members)

	return enum
}

// maxPointersAllowed implements maxPointerConstrainedType.
func (a *Enum) maxPointersAllowed() int {
	return 0
}
