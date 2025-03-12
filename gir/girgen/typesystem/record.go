package typesystem

import (
	"log"

	"github.com/diamondburned/gotk4/gir"
)

type Record struct {
	baseType

	Fields []*Field

	IsTypeStructFor Type

	Functions    []*CallableSignature
	Methods      []*CallableSignature
	Constructors []*CallableSignature

	// TODO:
	Unions     []*Union
	Properties []*struct{}
}

func DeclareRecord(ctx context, v gir.Record) *Record {
	if ctx.skipType(v) {
		return nil
	}

	r := &Record{
		baseType: baseType{
			girName: v.Name,
			goType:  v.Name,
			cGoType: "C." + v.CType,
			cType:   v.CType,
		},
	}

	return r
}

func (r *Record) resolveNested(ctx context, v gir.Record) {
	if v.GLibIsGTypeStructFor != "" {
		classType := ctx.findType(&gir.Type{Name: v.GLibIsGTypeStructFor})

		if classType == nil {
			return
		}

		switch classType.(type) {
		case *Class:
		case *Interface:
		default:
			log.Printf("%s should be a type struct for %s, but %s is %T and not class or interface\n", v.Name, v.GLibIsGTypeStructFor, v.GLibIsGTypeStructFor, classType)
			return
		}

		r.IsTypeStructFor = classType
	}

	for _, v := range v.Functions {
		if t := DeclareFunction(ctx, v); t != nil {
			r.Functions = append(r.Functions, t)
		}
	}

	for _, v := range v.Methods {
		if t := NewMethod(ctx, v); t != nil {
			r.Methods = append(r.Methods, t)
		}
	}

	for _, v := range v.Constructors {
		if t := DeclareConstructor(ctx, v); t != nil {
			r.Constructors = append(r.Constructors, t)
		}
	}

	if !v.Disguised {
		for _, v := range v.Fields {
			if t := NewField(ctx, v); t != nil {
				r.Fields = append(r.Fields, t)
			}
		}
	}
}
