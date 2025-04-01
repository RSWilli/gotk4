package typesystem

import (
	"fmt"
	"iter"
	"reflect"

	"github.com/diamondburned/gotk4/gir"
)

// Class is a type that extends GObject
type Class struct {
	BaseType

	Doc

	// GoInterfaceName is the name of the interface that describes the class. Every extending class will implement this
	// interface. Constructors and methods on the class will use the interface (e.g. MyClasser) name instead of the pointer type
	// (e.g. *MyClass) to allow easy passing of child class types.
	GoInterfaceName string

	GoWrapBaseClassFunction string

	// unsafe constructors names:
	GoUnsafeFromGlibBorrowFunction string
	GoUnsafeFromGlibFullFunction   string
	GoUnsafeFromGlibNoneFunction   string

	GoPrivateUpcastMethod string

	GoUnsafeToGlibNoneFunction string
	GoUnsafeToGlibFullFunction string

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
		// FIXME: this should instead register a fundamental type in the namespace
		return nil
	}

	ctype := v.CType

	if ctype == "" {
		ctype = v.Name
	}

	c := &Class{
		Doc:             NewDoc(&v.InfoAttrs, &v.InfoElements),
		Abstract:        v.Abstract,
		GoInterfaceName: v.Name + "Like",

		GoWrapBaseClassFunction: fmt.Sprintf("unsafeWrap%s", v.Name),
		GoPrivateUpcastMethod:   fmt.Sprintf("upcastTo%s", v.CType), // use cidentifier to not shadow parent methods

		GoUnsafeFromGlibBorrowFunction: fmt.Sprintf("Unsafe%sFromGlibBorrow", v.Name),
		GoUnsafeFromGlibNoneFunction:   fmt.Sprintf("Unsafe%sFromGlibNone", v.Name),
		GoUnsafeFromGlibFullFunction:   fmt.Sprintf("Unsafe%sFromGlibFull", v.Name),

		GoUnsafeToGlibNoneFunction: fmt.Sprintf("Unsafe%sToGlibNone", v.Name),
		GoUnsafeToGlibFullFunction: fmt.Sprintf("Unsafe%sToGlibFull", v.Name),
		BaseType: BaseType{
			GirName: v.Name,
			GoTyp:   v.Name,
			CGoTyp:  "C." + ctype,
			CTyp:    ctype,

			GlibGetTypeFn: v.GLibGetType,
		},
		gir: v,
	}

	return c
}

func (c *Class) resolve(e *env) bool {
	e = e.sub("class", c.gir.CType)

	if c.gir.GLibTypeStruct != "" {
		typeStructType := e.findTypeByGIRName(c.gir.GLibTypeStruct)

		if typeStructType == nil {
			return false
		}

		typeStruct, ok := typeStructType.(*Record)

		if !ok {
			e.logger.Warn("type struct is not a record", "actual", reflect.TypeOf(typeStructType).String())
			return false
		}

		c.TypeStruct = typeStruct
	}

	parent := e.findTypeByGIRName(c.gir.Parent)

	if parent == nil {
		return false
	}

	if !IsClass(parent) {
		e.logger.Warn("parent is not a class", "parent", parent.GIRName(), "actual", reflect.TypeOf(UnderlyingType(parent)).String())
		return false
	}

	c.Parent = parent

	for _, impl := range c.gir.Implements {
		inter := e.findTypeByGIRName(impl.Name)

		if inter == nil {
			e.logger.Info("implemented interface not found", "interface", impl.Name)
			continue
		}

		if !IsInterface(inter) {
			return false
		}

		if c.redundantImplements(inter) {
			e.logger.Debug("skipping redundant interface, as it is already implemented by parents", "interface", impl.Name)
			continue
		}

		c.Implements = append(c.Implements, inter)
	}

	return true
}

func (c *Class) redundantImplements(inter Type) bool {
	for _, impl := range c.Implements {
		if UnderlyingType(impl) == UnderlyingType(inter) {
			return true
		}
	}

	if c.Parent != nil {
		parent := UnderlyingType(c.Parent).(*Class)
		return parent.redundantImplements(inter)
	}

	return false
}

func (c *Class) declareNested(e *env) {
	e = e.sub("class", c.gir.CType)

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
			// reffing will be done on the base GObject, and we don't want such methods generated
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

func (c *Class) BaseClassGoUnsafeFromGlibBorrowFunction() string {
	switch p := c.BaseClass().(type) {
	case *Class:
		return p.GoUnsafeFromGlibBorrowFunction
	case *ForeignType:
		return p.AddForeignNamespace(p.Type.(*Class).GoUnsafeFromGlibBorrowFunction)
	default:
		panic("invalid base class")
	}
}

func (c *Class) BaseClassGoUnsafeFromGlibFullFunction() string {
	switch p := c.BaseClass().(type) {
	case *Class:
		return p.GoUnsafeFromGlibFullFunction
	case *ForeignType:
		return p.AddForeignNamespace(p.Type.(*Class).GoUnsafeFromGlibFullFunction)
	default:
		panic("invalid base class")
	}
}

func (c *Class) BaseClassGoUnsafeFromGlibNoneFunction() string {
	switch p := c.BaseClass().(type) {
	case *Class:
		return p.GoUnsafeFromGlibNoneFunction
	case *ForeignType:
		return p.AddForeignNamespace(p.Type.(*Class).GoUnsafeFromGlibNoneFunction)
	default:
		panic("invalid base class")
	}
}

func (c *Class) BaseClassGoUnsafeToGlibFullFunction() string {
	switch p := c.BaseClass().(type) {
	case *Class:
		return p.GoUnsafeToGlibFullFunction
	case *ForeignType:
		return p.AddForeignNamespace(p.Type.(*Class).GoUnsafeToGlibFullFunction)
	default:
		panic("invalid base class")
	}
}

func (c *Class) BaseClassGoUnsafeToGlibNoneFunction() string {
	switch p := c.BaseClass().(type) {
	case *Class:
		return p.GoUnsafeToGlibNoneFunction
	case *ForeignType:
		return p.AddForeignNamespace(p.Type.(*Class).GoUnsafeToGlibNoneFunction)
	default:
		panic("invalid base class")
	}
}

// BaseClass returns the base class from the view of the namespace of c
func (c *Class) BaseClass() Type {
	parents := c.AllParents()

	if len(parents) == 0 {
		return c
	}

	return parents[len(parents)-1]
}

// AllParents returns a list of parents. Note that the list is relative to the current namespace,
// meaning that foreign types will be labeled as such.
//
// This also requires that the following chain can happen:
//
//	C1 -> ForeignA[C2] -> C3 -> ForeignB[C4] -> C5
//
// From the view of C1 the classes C3 and C5 are also foreign in there respective namespaces. C5 is the base
// class so it will be used as a pointer. They will get returned as:
//
//	[ForeignA[C2], ForeignA[C3], ForeignB[C4], ForeignB[C5]]
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
			panic(fmt.Sprintf("unexpected class parent %T", parent))
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

func GetClassParent(t Type) Type {
	switch p := t.(type) {
	case *Class:
		return p.Parent
	case *ForeignType:
		return GetClassParent(p.Type)
	default:
		panic("invalid type received")
	}
}

func ClassGoInterfaceName(t Type) string {
	switch p := t.(type) {
	case *PointerType:
		return ClassGoInterfaceName(p.Base)
	case *Class:
		return p.GoInterfaceName
	case *ForeignType:
		return p.AddForeignNamespace(ClassGoInterfaceName(p.Type))
	default:
		panic("invalid type received")
	}
}
