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

	// gir is used to resolve the class and it's nested definitions after it has been declared
	gir gir.Class

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

// DeclareClass declares a new class, it returns the class type and whether this instance needs additional
// resolving, because it has dependencies to other types
func DeclareClass(e *env, v gir.Class) (*Class, bool) {
	ctype := v.CType

	if ctype == "" {
		ctype = v.Name
	}

	c := &Class{
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

	if c.gir.GLibTypeStruct != "" {
		typeStructType := e.findType(&gir.Type{Name: c.gir.GLibTypeStruct})

		if typeStructType == nil {
			return nil, false
		}

		typeStruct, ok := typeStructType.(*Record)

		if !ok {
			log.Printf("type struct for %s is not a record but instead %T", c.gir.Name, typeStructType)
			return nil, false
		}

		c.TypeStruct = typeStruct
	}

	return c, v.Parent != "" || len(v.Implements) > 0
}

func (c *Class) resolve(e *env) bool {
	if c.gir.Parent != "" {
		parent := e.findType(&gir.Type{Name: c.gir.Parent})

		if parent == nil {
			return false
		}

		if !IsClass(parent) {
			log.Printf("parent %s of class %s is not a class, but %T instead\n", parent.GIRName(), c.gir.Name, UnderlyingType(parent))
			return false
		}

		c.Parent = parent
	}

	for _, impl := range c.gir.Implements {
		inter := e.findType(&gir.Type{Name: impl.Name})

		if inter == nil {
			log.Printf("interface %s not found\n", impl.Name)
			continue
		}

		if !IsInterface(inter) && !IsClass(inter) {
			return false
		}

		c.Implements = append(c.Implements, inter)
	}

	return true
}

func (c *Class) declareNested(e *env) {
	for _, v := range c.gir.Functions {
		if t := DeclareFunction(e, v); t != nil {
			c.Functions = append(c.Functions, t)
		}
	}

	for _, v := range c.gir.Methods {
		// TODO: handle ref and unref like the records do

		if t := NewMethod(e, v); t != nil {
			c.Methods = append(c.Methods, t)
		}
	}

	for _, v := range c.gir.Constructors {
		if t := DeclareConstructor(e, c, v); t != nil {
			c.Constructors = append(c.Constructors, t)
		}
	}

	for _, v := range c.gir.Signals {
		if t := NewSignal(e, v); t != nil {
			c.Signals = append(c.Signals, t)
		}
	}

	for _, v := range c.gir.Fields {
		if t := NewField(e, c, v); t != nil {
			c.Fields = append(c.Fields, t)
		}
	}

	if c.TypeStruct != nil {
		for _, v := range c.gir.VirtualMethods {
			if t := NewVirtualMethod(e, c, c.TypeStruct, v); t != nil {
				c.VirtualMethods = append(c.VirtualMethods, t)
			}
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
