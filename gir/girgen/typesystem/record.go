package typesystem

import (
	"log"

	"github.com/diamondburned/gotk4/gir"
)

type Record struct {
	baseType

	Fields []*Field

	IsTypeStructFor Type

	Functions    []*Callable
	Methods      []*Callable
	Constructors []*Callable

	// TODO:
	Unions     []*Union
	Properties []*struct{}
}

func NewRecord(ns *Namespace, v gir.Record) *Record {
	if !v.IsIntrospectable() {
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

func (r *Record) resolveNested(ns *Namespace, v gir.Record) {
	if v.GLibIsGTypeStructFor != "" {
		classType := ns.findType(&gir.Type{Name: v.GLibIsGTypeStructFor})

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
		if t := NewFunction(ns, v); t != nil {
			r.Functions = append(r.Functions, t)
		}
	}

	for _, v := range v.Methods {
		if t := NewMethod(ns, v); t != nil {
			r.Methods = append(r.Methods, t)
		}
	}

	for _, v := range v.Constructors {
		if t := NewConstructor(ns, v); t != nil {
			r.Constructors = append(r.Constructors, t)
		}
	}

	if !v.Disguised {
		for _, v := range v.Fields {
			if t := NewField(ns, v); t != nil {
				r.Fields = append(r.Fields, t)
			}
		}
	}
}
