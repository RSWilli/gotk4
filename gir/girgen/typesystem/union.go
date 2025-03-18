package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
)

type Union struct {
	BaseType
	gir gir.Union

	Doc Doc

	GetType string

	Functions    []*CallableSignature
	Methods      []*CallableSignature
	Constructors []*CallableSignature
	Fields       []*Field
}

func DeclareUnion(e *env, v gir.Union) *Union {
	if e.skipType(v) {
		return nil
	}

	return &Union{
		Doc:     NewDoc(&v.InfoAttrs, &v.InfoElements),
		GetType: v.GLibGetType,
		BaseType: BaseType{
			GirName: v.Name,
			GoTyp:   v.Name,
			CGoTyp:  "C." + v.CType,
			CTyp:    v.CType,

			GlibGetTypeFn: v.GLibGetType,
		},
		gir: v,
	}
}

func (u *Union) declareNested(e *env) resolvedState {
	for _, v := range u.gir.Functions {
		if t := DeclareFunction(e, v); t != nil {
			u.Functions = append(u.Functions, t)
		}
	}

	for _, v := range u.gir.Methods {
		if t := NewMethod(e, v); t != nil {
			u.Methods = append(u.Methods, t)
		}
	}

	for _, v := range u.gir.Constructors {
		if t := DeclareConstructor(e, u, v); t != nil {
			u.Constructors = append(u.Constructors, t)
		}
	}

	for _, v := range u.gir.Fields {
		if t := NewField(e, u, v); t != nil {
			u.Fields = append(u.Fields, t)
		}
	}

	// for range v.Records {
	// 	TODO
	// }

	return okResolved
}
