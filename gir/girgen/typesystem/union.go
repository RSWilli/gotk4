package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
)

type Union struct {
	BaseType

	Doc Doc

	GetType string

	Functions    []*CallableSignature
	Methods      []*CallableSignature
	Constructors []*CallableSignature
	Fields       []*Field
}

func DeclareUnion(ctx context, v gir.Union) *Union {
	if ctx.skipType(v) {
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
	}
}

func (u *Union) resolveNested(ns context, v gir.Union) {
	for _, v := range v.Functions {
		if t := DeclareFunction(ns, v); t != nil {
			u.Functions = append(u.Functions, t)
		}
	}

	for _, v := range v.Methods {
		if t := NewMethod(ns, v); t != nil {
			u.Methods = append(u.Methods, t)
		}
	}

	for _, v := range v.Constructors {
		if t := DeclareConstructor(ns, u, v); t != nil {
			u.Constructors = append(u.Constructors, t)
		}
	}

	for _, v := range v.Fields {
		if t := NewField(ns, u, v); t != nil {
			u.Fields = append(u.Fields, t)
		}
	}

	// for range v.Records {
	// 	TODO
	// }
}
