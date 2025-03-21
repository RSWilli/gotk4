package typesystem

import (
	"fmt"
	"iter"
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

	// unsafe constructors names:
	GoUnsafeBorrowFunction       string
	GoUnsafeTransferFullFunction string
	GoUnsafeTransferNoneFunction string

	GoUnsafeToGlibNoneMethod string
	GoUnsafeToGlibFullMethod string

	Abstract bool

	// gir is used to resolve the class and it's nested definitions after it has been declared
	gir gir.Class

	TypeStruct *Record
	Parent     Type
	// Implements contains the implemented (foreign) interfaces
	Implements []Type

	Functions      []*CallableSignature
	Methods        []*CallableSignature
	VirtualMethods []*VirtualMethod
	Constructors   []*CallableSignature
	Fields         []*Field
	Signals        []*Signal
}

func DeclareClass(e *env, v gir.Class) *Class {
	if !v.IsIntrospectable() {
		return nil
	}

	if e.skip(nil, v) {
		return nil
	}

	if v.Parent == "" {
		// we can't handle non GObject child classes for now
		return nil
	}

	ctype := v.CType

	if ctype == "" {
		ctype = v.Name
	}

	c := &Class{
		Doc:             NewDoc(&v.InfoAttrs, &v.InfoElements),
		Abstract:        v.Abstract,
		GoInterfaceName: strcases.Interfacify(v.Name),

		GoUnsafeBorrowFunction:       fmt.Sprintf("Unsafe%sFromGlibBorrow", v.Name),
		GoUnsafeTransferNoneFunction: fmt.Sprintf("Unsafe%sFromGlibNone", v.Name),
		GoUnsafeTransferFullFunction: fmt.Sprintf("Unsafe%sFromGlibFull", v.Name),

		GoUnsafeToGlibNoneMethod: fmt.Sprintf("Unsafe%sToGlibNone", v.Name),
		GoUnsafeToGlibFullMethod: fmt.Sprintf("Unsafe%sToGlibFull", v.Name),
		BaseType: BaseType{
			GirName: v.Name,
			GoTyp:   v.Name,
			CGoTyp:  "C." + ctype,
			CTyp:    ctype,

			GlibGetTypeFn: v.GLibGetType,
		},
		gir: v,
	}

	if c.gir.GLibTypeStruct != "" {
		typeStructType := e.findType(&gir.Type{Name: c.gir.GLibTypeStruct})

		if typeStructType == nil {
			return nil
		}

		typeStruct, ok := typeStructType.(*Record)

		if !ok {
			log.Printf("type struct for %s is not a record but instead %T", c.gir.Name, typeStructType)
			return nil
		}

		c.TypeStruct = typeStruct
	}

	return c
}

func (c *Class) resolve(e *env) bool {
	parent := e.findType(&gir.Type{Name: c.gir.Parent})

	if parent == nil {
		return false
	}

	if !IsClass(parent) {
		log.Printf("parent %s of class %s is not a class, but %T instead\n", parent.GIRName(), c.gir.Name, UnderlyingType(parent))
		return false
	}

	c.Parent = parent

	for _, impl := range c.gir.Implements {
		inter := e.findType(&gir.Type{Name: impl.Name})

		if inter == nil {
			log.Printf("interface %s not found\n", impl.Name)
			continue
		}

		if !IsInterface(inter) {
			return false
		}

		c.Implements = append(c.Implements, inter)
	}

	return true
}

func (c *Class) declareNested(e *env) {
	for _, v := range c.gir.Functions {
		if t := DeclareFunction(e, c, v); t != nil {
			c.Functions = append(c.Functions, t)
		}
	}

	for _, v := range c.gir.Methods {
		if v.Name == "weak_ref" || v.Name == "weak_unref" {
			// we don't want the user to be able to weakly reference the object
			// as there are better tools for this and this will only cause problems
			continue
		}

		if v.Name == "ref" || v.Name == "unref" {
			// reffing will be done on the parent GObject, and we don't want such methods generated
			// because they will cause mem leaks
			continue
		}

		if t := NewMethod(e, c, v); t != nil {
			c.Methods = append(c.Methods, t)
		}
	}

	for _, v := range c.gir.Constructors {
		if t := DeclareConstructor(e, c, v); t != nil {
			c.Constructors = append(c.Constructors, t)
		}
	}

	for _, v := range c.gir.Signals {
		if t := NewSignal(e, c, v); t != nil {
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

// ImplementsConstructors returns the underlying type names and the borrow constructor name of the implemented
// interfaces of the class. We only need to borrow because the class parent constructor is already taking a reference
func (c *Class) ImplementsConstructors() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for _, inter := range c.Implements {
			typeName := UnderlyingType(inter).GoType()
			var constructorName string

			switch i := inter.(type) {
			case *Interface:
				constructorName = i.GoUnsafeBorrowFunction
			case *ForeignType:
				constructorName = i.AddForeignNamespace(i.Type.(*Interface).GoUnsafeBorrowFunction)
			default:
				panic("invalid implemented interface")
			}

			if !yield(typeName, constructorName) {
				return
			}
		}
	}
}

func (c *Class) ImplementedGoInterfaceNames() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, inter := range c.Implements {
			var name string

			switch i := inter.(type) {
			case *Interface:
				name = i.GoInterfaceName
			case *ForeignType:
				name = i.AddForeignNamespace(i.Type.(*Interface).GoInterfaceName)
			default:
				panic("invalid implemented interface")
			}

			if !yield(name) {
				return
			}
		}
	}
}

func (c *Class) ParentGoInterfaceName() string {
	switch p := c.Parent.(type) {
	case nil:
		panic("no parent")
	case *Class:
		return p.GoInterfaceName
	case *ForeignType:
		return p.AddForeignNamespace(p.Type.(*Class).GoInterfaceName)
	default:
		panic("invalid parent")
	}
}

func (c *Class) ParentGoUnsafeBorrowFunction() string {
	switch p := c.Parent.(type) {
	case nil:
		panic("no parent")
	case *Class:
		return p.GoUnsafeBorrowFunction
	case *ForeignType:
		return p.AddForeignNamespace(p.Type.(*Class).GoUnsafeBorrowFunction)
	default:
		panic("invalid parent")
	}
}

func (c *Class) ParentGoUnsafeTransferFullFunction() string {
	switch p := c.Parent.(type) {
	case nil:
		panic("no parent")
	case *Class:
		return p.GoUnsafeTransferFullFunction
	case *ForeignType:
		return p.AddForeignNamespace(p.Type.(*Class).GoUnsafeTransferFullFunction)
	default:
		panic("invalid parent")
	}
}

func (c *Class) ParentGoUnsafeTransferNoneFunction() string {
	switch p := c.Parent.(type) {
	case nil:
		panic("no parent")
	case *Class:
		return p.GoUnsafeTransferNoneFunction
	case *ForeignType:
		return p.AddForeignNamespace(p.Type.(*Class).GoUnsafeTransferNoneFunction)
	default:
		panic("invalid parent")
	}
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
