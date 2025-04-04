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

	ns, typ := e.findTypeByGIRName("GObject.Value")

	if typ == nil {
		e.logger.Warn("skipping because gvalue was not found")
		return nil
	}

	b := &Bitfield{
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

	b.Members = GetMembers(e, b, v.Members)

	return b
}

// pointersAllowed implements Type.
func (a *Bitfield) pointersAllowed(pointers int) bool {
	return pointers == 0
}
