package typesystem

import (
	"fmt"
	"iter"
	"reflect"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

// Interface declares an interface implemented by a GObject.
type Interface struct {
	Doc
	BaseType
	gir gir.Interface

	TypeStruct *Record

	// Parent is always (foreign) GObject, because at runtime we will receive a GObject pointer
	// and wrap it. We look it up because we don't know the implementation here.
	Parent Type

	GoWrapBaseClassFunction string

	GoInterfaceName string

	// unsafe constructors names:
	GoUnsafeBorrowFunction       string
	GoUnsafeTransferFullFunction string
	GoUnsafeTransferNoneFunction string

	GoUnsafeToGlibNoneMethod string
	GoUnsafeToGlibFullMethod string

	Prerequesite []Type // Class or Interface

	Functions      []*CallableSignature
	Methods        []*CallableSignature
	VirtualMethods []*VirtualMethod
	Signals        []*Signal
}

func DeclareInterface(e *env, v gir.Interface) *Interface {
	if !v.IsIntrospectable() {
		return nil
	}

	if e.skip(nil, v) {
		return nil
	}

	ctype := v.CType

	if ctype == "" {
		ctype = v.Name
	}

	i := &Interface{
		Doc: NewDoc(&v.InfoAttrs, &v.InfoElements),
		BaseType: BaseType{
			GirName: v.Name,
			GoTyp:   v.Name,
			CGoTyp:  "C." + ctype,
			CTyp:    ctype,

			GlibGetTypeFn: v.GLibGetType,
		},
		GoInterfaceName: strcases.Interfacify(v.Name),

		GoWrapBaseClassFunction: fmt.Sprintf("unsafeWrap%s", v.Name),

		GoUnsafeBorrowFunction:       fmt.Sprintf("Unsafe%sFromGlibBorrow", v.Name),
		GoUnsafeTransferNoneFunction: fmt.Sprintf("Unsafe%sFromGlibNone", v.Name),
		GoUnsafeTransferFullFunction: fmt.Sprintf("Unsafe%sFromGlibFull", v.Name),

		GoUnsafeToGlibNoneMethod: fmt.Sprintf("Unsafe%sToGlibNone", v.Name),
		GoUnsafeToGlibFullMethod: fmt.Sprintf("Unsafe%sToGlibFull", v.Name),
		gir:                      v,
	}

	return i
}

func (in *Interface) resolve(e *env) bool {
	e = e.sub("interface", in.gir.CType)

	if in.gir.GLibTypeStruct != "" {
		typeStructType := e.findTypeByGIRName(in.gir.GLibTypeStruct)

		if typeStructType == nil {
			return false
		}

		typeStruct, ok := typeStructType.(*Record)

		if !ok {
			e.logger.Warn("type struct is not a record", "actual", reflect.TypeOf(typeStructType).String())
			return false
		}

		in.TypeStruct = typeStruct
	}

	parent := e.findTypeByGIRName("GObject.Object")

	if parent == nil {
		e.logger.Error("GObject.Object not found")
		return false
	}

	if !IsClass(parent) {
		return false
	}

	in.Parent = parent

	for _, prereq := range in.gir.Prerequisites {
		inter := e.findTypeByGIRName(prereq.Name)

		if inter == nil {
			e.logger.Info("interface prerequesite not found", "interface", prereq.Name)
			return false
		}

		if !IsInterface(inter) && !IsClass(inter) {
			e.logger.Warn("prerequesite is not a class or an interface", "prerequesite", inter.GIRName(), "actual", reflect.TypeOf(UnderlyingType(parent)).String())
			return false
		}

		in.Prerequesite = append(in.Prerequesite, inter)
	}

	return true
}

func (in *Interface) declareNested(e *env) {
	e = e.sub("interface", in.gir.CType)

	for _, v := range in.gir.Functions {
		if t := DeclareFunction(e, in, v); t != nil {
			in.Functions = append(in.Functions, t)
		}
	}

	for _, v := range in.gir.Methods {
		if t := NewMethod(e, in, v); t != nil {
			in.Methods = append(in.Methods, t)
		}
	}

	for _, v := range in.gir.Signals {
		if t := NewSignal(e, in, v); t != nil {
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

func (in *Interface) ParentGoInterfaceName() string {
	switch p := in.Parent.(type) {
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

func (in *Interface) ParentGoUnsafeBorrowFunction() string {
	switch p := in.Parent.(type) {
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

func (c *Interface) ParentGoUnsafeTransferFullFunction() string {
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

func (c *Interface) ParentGoUnsafeTransferNoneFunction() string {
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

// PrerequesiteConstructors returns the underlying type names and the borrow constructor name of the
// interface's prerequesites. We only need to borrow because the class parent constructor is already taking a reference
func (c *Interface) PrerequesiteConstructors() iter.Seq2[string, string] {
	return func(yield func(string, string) bool) {
		for _, inter := range c.Prerequesite {
			typeName := UnderlyingType(inter).GoType()
			var constructorName string

			switch i := inter.(type) {
			case *Class:
				constructorName = i.GoUnsafeBorrowFunction
			case *Interface:
				constructorName = i.GoUnsafeBorrowFunction
			case *ForeignType:
				switch t := i.Type.(type) {
				case *Class:
					constructorName = i.AddForeignNamespace(t.GoUnsafeBorrowFunction)
				case *Interface:
					constructorName = i.AddForeignNamespace(t.GoUnsafeBorrowFunction)
				default:
					panic("invalid interface prerequesite")
				}
			default:
				panic("invalid interface prerequesite")
			}

			if !yield(typeName, constructorName) {
				return
			}
		}
	}
}

func (c *Interface) PrerequesitesGoInterfaceNames() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, inter := range c.Prerequesite {
			var name string

			switch i := inter.(type) {
			case *Class:
				name = i.GoInterfaceName
			case *Interface:
				name = i.GoInterfaceName
			case *ForeignType:
				switch t := i.Type.(type) {
				case *Class:
					name = i.AddForeignNamespace(t.GoInterfaceName)
				case *Interface:
					name = i.AddForeignNamespace(t.GoInterfaceName)
				default:
					panic("invalid interface prerequesite")
				}
			default:
				panic("invalid interface prerequesite")
			}

			if !yield(name) {
				return
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
