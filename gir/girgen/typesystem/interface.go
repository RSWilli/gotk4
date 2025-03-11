package typesystem

import (
	"log"

	"github.com/diamondburned/gotk4/gir"
)

type Interface struct {
	baseType
	GetType    string
	TypeStruct *Record

	// Valid signifies that the interfaces prerequesites have been resolved correctly.
	// if this is false then the class generator must ignore this or the generation will
	// be wrong
	Valid bool

	Prerequesite []Type // Class or Interface

	Functions      []*CallableSignature
	Methods        []*CallableSignature
	VirtualMethods []*VirtualMethod
	Signals        []*Signal
}

func DeclareInterface(ns *Namespace, v gir.Interface) *Interface {
	ctype := v.CType

	if ctype == "" {
		ctype = v.Name
	}

	return &Interface{
		GetType: v.GLibGetType,
		baseType: baseType{
			girName: v.Name,
			goType:  v.Name,
			cGoType: "C." + ctype,
			cType:   ctype,
		},
	}
}

func (r *Interface) resolveNested(ns *Namespace, v gir.Interface) {
	for _, prereq := range v.Prerequisites {
		inter := ns.findType(&gir.Type{Name: prereq.Name})

		if inter == nil {
			log.Printf("interface %s not found\n", prereq.Name)
			return
		}

		switch underlying := UnderlyingType(inter).(type) {
		case *Class:
		case *Interface:
		default:
			log.Printf("prerequisite %s of interface %s is not class or interface, but %T instead\n", inter.GIRName(), v.Name, underlying)
			return
		}

		r.Prerequesite = append(r.Prerequesite, inter)
	}

	if v.GLibTypeStruct != "" {
		typeStructType := ns.findType(&gir.Type{Name: v.GLibTypeStruct})

		if typeStructType == nil {
			return
		}

		typeStruct, ok := typeStructType.(*Record)

		if !ok {
			return
		}

		r.TypeStruct = typeStruct

		for _, v := range v.VirtualMethods {
			if t := NewVirtualMethod(ns, r, typeStruct, v); t != nil {
				r.VirtualMethods = append(r.VirtualMethods, t)
			}
		}
	}

	// ------------- from here on the interface is valid and we will only omit invalid parts ----------------
	r.Valid = true

	for _, v := range v.Functions {
		if t := DeclareFunction(ns, v); t != nil {
			r.Functions = append(r.Functions, t)
		}
	}

	for _, v := range v.Methods {
		if t := NewMethod(ns, v); t != nil {
			r.Methods = append(r.Methods, t)
		}
	}

	for _, v := range v.Signals {
		if t := NewSignal(ns, v); t != nil {
			r.Signals = append(r.Signals, t)
		}
	}
}
