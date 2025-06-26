package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type Enum struct {
	BaseType
	Doc
	Marshaler

	gir *gir.Enum

	Members Members

	Functions []*CallableSignature
}

func DeclareEnum(e *env, v *gir.Enum) *Enum {
	e = e.sub("enum", v.Name)

	if !v.IsIntrospectable() {
		e.logger.Warn("skipping because not introspectable")
		return nil
	}

	if e.skip(nil, v) {
		return nil
	}

	enum := &Enum{
		gir: v,
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

func (enum *Enum) declareNested(e *env) {
	for _, v := range enum.gir.Functions {
		if t := DeclarePrefixedFunction(e, enum, v.CallableAttrs); t != nil {
			enum.Functions = append(enum.Functions, t)
		}
	}
}

// maxPointersAllowed implements maxPointerConstrainedType.
func (a *Enum) maxPointersAllowed() int {
	return 0
}
