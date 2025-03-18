package typesystem

import (
	"log"

	"github.com/diamondburned/gotk4/gir"
)

type Interface struct {
	BaseType
	TypeStruct *Record

	gir gir.Interface

	Prerequesite []Type // Class or Interface

	Functions      []*CallableSignature
	Methods        []*CallableSignature
	VirtualMethods []*VirtualMethod
	Signals        []*Signal
}

func DeclareInterface(e *env, v gir.Interface) (*Interface, bool) {
	ctype := v.CType

	if ctype == "" {
		ctype = v.Name
	}

	i := &Interface{
		BaseType: BaseType{
			GirName: v.Name,
			GoTyp:   v.Name,
			CGoTyp:  "C." + ctype,
			CTyp:    ctype,

			GlibGetTypeFn: v.GLibGetType,
		},
		gir: v,
	}

	if v.GLibTypeStruct != "" {
		typeStructType := e.findType(&gir.Type{Name: v.GLibTypeStruct})

		if typeStructType == nil {
			return nil, false
		}

		typeStruct, ok := typeStructType.(*Record)

		if !ok {
			return nil, false
		}

		i.TypeStruct = typeStruct
	}

	return i, len(v.Prerequisites) > 0
}

func (in *Interface) resolve(e *env) bool {
	for _, prereq := range in.gir.Prerequisites {
		inter := e.findType(&gir.Type{Name: prereq.Name})

		if inter == nil {
			log.Printf("interface %s not found\n", prereq.Name)
			return false
		}

		if !IsInterface(inter) && !IsClass(inter) {
			log.Printf("prerequisite %s of interface %s is not class or interface, but %T instead\n", inter.GIRName(), in.gir.Name, UnderlyingType(inter))
			return false
		}

		in.Prerequesite = append(in.Prerequesite, inter)
	}

	return true
}

func (in *Interface) declareNested(e *env) {
	for _, v := range in.gir.Functions {
		if t := DeclareFunction(e, v); t != nil {
			in.Functions = append(in.Functions, t)
		}
	}

	for _, v := range in.gir.Methods {
		if t := NewMethod(e, v); t != nil {
			in.Methods = append(in.Methods, t)
		}
	}

	for _, v := range in.gir.Signals {
		if t := NewSignal(e, v); t != nil {
			in.Signals = append(in.Signals, t)
		}
	}

	if in.TypeStruct != nil {
		for _, v := range in.gir.VirtualMethods {
			if t := NewVirtualMethod(e, in, in.TypeStruct, v); t != nil {
				in.VirtualMethods = append(in.VirtualMethods, t)
			}
		}
	}
}

func IsInterface(t Type) bool {
	switch p := t.(type) {
	case *Interface:
		return true
	case *ForeignType:
		return IsInterface(p.Type)
	default:
		return false
	}
}
