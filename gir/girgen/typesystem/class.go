package typesystem

import (
	"log"

	"github.com/diamondburned/gotk4/gir"
)

type Class struct {
	baseType

	Abstract bool

	// Valid signifies that the classes parent has been resolved correctly
	// if this is false then the class generator must ignore this or the generation will
	// be wrong
	Valid bool

	GetType string

	TypeStruct *Record
	Parent     Type
	// Implements contains the implemented interfaces. This also contains [*Class].
	Implements []Type

	Functions      []*Callable
	Methods        []*Callable
	VirtualMethods []*VirtualMethod
	Constructors   []*Callable
	Fields         []*Field
	Signals        []*Signal
}

func NewClass(ns *Namespace, v gir.Class) *Class {
	ctype := v.CType

	if ctype == "" {
		ctype = v.Name
	}

	return &Class{
		Abstract: v.Abstract,
		GetType:  v.GLibGetType,
		baseType: baseType{
			girName: v.Name,
			goType:  v.Name,
			cGoType: "C." + ctype,
			cType:   ctype,
		},
	}
}

func (r *Class) resolveNested(ns *Namespace, v gir.Class) {
	if v.Parent != "" {
		parent := ns.findType(&gir.Type{Name: v.Parent})

		if parent == nil {
			return
		}

		switch underlying := UnderlyingType(parent).(type) {
		case *Class:
		default:
			log.Printf("parent %s of class %s is not a class, but %T instead\n", parent.GIRName(), v.Name, underlying)
			return
		}

		r.Parent = parent
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
			if t := NewVirtualMethod(ns, typeStruct, v); t != nil {
				r.VirtualMethods = append(r.VirtualMethods, t)
			}
		}
	}

	// ------------- from here on the class is valid and we will only omit invalid parts ----------------
	r.Valid = true

	for _, impl := range v.Implements {
		inter := ns.findType(&gir.Type{Name: impl.Name})

		if inter == nil {
			log.Printf("interface %s not found\n", impl.Name)
			continue
		}

		r.Implements = append(r.Implements, inter)
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

	for _, v := range v.Signals {
		if t := NewSignal(ns, v); t != nil {
			r.Signals = append(r.Signals, t)
		}
	}

	for _, v := range v.Fields {
		if t := NewField(ns, v); t != nil {
			r.Fields = append(r.Fields, t)
		}
	}
}
