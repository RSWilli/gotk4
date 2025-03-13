package typesystem

import (
	"fmt"
	"log"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type Class struct {
	BaseType

	Doc

	// GoInterfaceName is the name of the interface that describes the class. Every extending class will implement this
	// interface. Constructors and methods on the class will use the interface (e.g. MyClasser) name instead of the pointer type
	// (e.g. *MyClass) to allow easy passing of child class types.
	GoInterfaceName string

	// GoWrapFunctionName is the name of the function that turns a C pointer into the go struct. This function must construct
	// the go struct with all the parent structs embedded into it
	GoWrapFunctionName string

	Abstract bool

	// Valid signifies that the classes parent has been resolved correctly
	// if this is false then the class generator must ignore this or the generation will
	// be wrong
	Valid bool

	TypeStruct *Record
	Parent     Type
	// Implements contains the implemented interfaces and parent [Class].
	Implements []Type

	Functions      []*CallableSignature
	Methods        []*CallableSignature
	VirtualMethods []*VirtualMethod
	Constructors   []*CallableSignature
	Fields         []*Field
	Signals        []*Signal
}

func NewClass(ns context, v gir.Class) *Class {
	ctype := v.CType

	if ctype == "" {
		ctype = v.Name
	}

	return &Class{
		Doc:                NewDoc(&v.InfoAttrs, &v.InfoElements),
		Abstract:           v.Abstract,
		GoInterfaceName:    strcases.Interfacify(v.Name),
		GoWrapFunctionName: fmt.Sprintf("wrap%s", v.Name),
		BaseType: BaseType{
			GirName: v.Name,
			GoTyp:   v.Name,
			CGoTyp:  "C." + ctype,
			CTyp:    ctype,

			GlibGetTypeFn: v.GLibGetType,
		},
	}
}

func (r *Class) resolveNested(ns context, v gir.Class) {
	if v.Parent != "" {
		parent := ns.findType(&gir.Type{Name: v.Parent})

		if parent == nil {
			return
		}

		if !IsClass(parent) {
			log.Printf("parent %s of class %s is not a class, but %T instead\n", parent.GIRName(), v.Name, UnderlyingType(parent))
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
			log.Printf("type struct for %s is not a record but instead %T", v.Name, typeStructType)
			return
		}

		r.TypeStruct = typeStruct

		for _, v := range v.VirtualMethods {
			if t := NewVirtualMethod(ns, r, typeStruct, v); t != nil {
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
		if t := DeclareFunction(ns, v); t != nil {
			r.Functions = append(r.Functions, t)
		}
	}

	for _, v := range v.Methods {
		// TODO: handle ref and unref like the records do

		if t := NewMethod(ns, v); t != nil {
			r.Methods = append(r.Methods, t)
		}
	}

	for _, v := range v.Constructors {
		if t := DeclareConstructor(ns, r, v); t != nil {
			r.Constructors = append(r.Constructors, t)
		}
	}

	for _, v := range v.Signals {
		if t := NewSignal(ns, v); t != nil {
			r.Signals = append(r.Signals, t)
		}
	}

	for _, v := range v.Fields {
		if t := NewField(ns, r, v); t != nil {
			r.Fields = append(r.Fields, t)
		}
	}
}

// AllParents returns a list of parents. Note that the list is relative to the current namespace,
// meaning that foreign types will be labeled as such.
//
// This also requires that the following chain can happen:
//
//	C1 -> ForeignA[C2] -> C3 -> ForeignB[C4] -> C5
//
// From the view of C1 the classes C3 and C5 are also foreign, so they will get returned as:
//
//	[ForeignA[C2], ForeignA[C3] -> ForeignB[C4] -> ForeignB[C5]]
func (c *Class) AllParents() []Type {
	parents := make([]Type, 0, 10) // abitrary cap

	var currentNs *Namespace
	var currentType = c.Parent
loop:
	for {
		switch parent := currentType.(type) {
		case nil:
			break loop
		case *ForeignType:
			currentNs = parent.SourceNamespace
			currentType = parent.Type
		case *Class:
			currentType = parent.Parent

			if currentNs == nil {
				parents = append(parents, parent)
				continue loop
			}

			parents = append(parents, &ForeignType{
				SourceNamespace: currentNs,
				Type:            parent,
			})

		default:
			log.Panicf("unexpected class parent %T", parent)
		}
	}

	return parents
}

func IsClass(t Type) bool {
	switch p := t.(type) {
	case *Class:
		return true
	case *ForeignType:
		return IsClass(p.Type)
	default:
		return false
	}
}
