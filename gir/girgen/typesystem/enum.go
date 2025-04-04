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
	e = e.sub("bitfield", v.CType)

	if !v.IsIntrospectable() {
		return nil
	}

	if e.skip(nil, v) {
		return nil
	}

	ns, typ := e.findTypeByGIRName("GObject.Value")

	if typ == nil {
		e.logger.Warn("skipping enum because gvalue was not found", "enum", v.Name)
		return nil
	}

	enum := &Enum{
		BaseType: BaseType{
			GirName: v.Name,
			GoTyp:   strcases.PascalToGo(v.Name),
			CGoTyp:  "C." + v.CType,
			CTyp:    v.CType,
		},
		Marshaler: newDefaultMarshaler(v.GLibGetType, CouldBeForeign[*Record]{
			Namespace: ns,
			Type:      typ.(*Record),
		}),
		Doc: NewDoc(&v.InfoAttrs, &v.InfoElements),
	}

	enum.Members = GetMembers(e, enum, v.Members)

	return enum
}

// pointersAllowed implements Type.
func (a *Enum) pointersAllowed(pointers int) bool {
	return pointers == 0
}
