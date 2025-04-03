package typesystem

import (
	"fmt"
	"reflect"

	"github.com/diamondburned/gotk4/gir"
)

// Interface declares an interface implemented by a GObject.
type Interface struct {
	Doc
	BaseType
	gir gir.Interface

	TypeStruct *Record

	// Parent is always (foreign) GObject, because at runtime we will receive a GObject pointer
	// and wrap it. We look it up because we don't know the implementation here.
	Parent CouldBeForeign[*Class]

	GoWrapBaseClassFunction string

	GoInterfaceName string

	BaseConversions
	Marshaler

	Prerequesite []CouldBeForeign[Type]

	Functions      []*CallableSignature
	Methods        []*CallableSignature
	VirtualMethods []*VirtualMethod
	Signals        []*Signal
}

// GoType implements Type. Use the interface type if a pointer is needed
func (in *Interface) GoType(pointers int) string {
	if pointers == 0 {
		return in.BaseType.GoType(0)
	}

	return in.GoInterfaceName
}

// pointersAllowed implements Type.
func (a *Interface) pointersAllowed(pointers int) bool {
	return pointers == 1
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
			GoTyp:   v.Name + "Instance",
			CGoTyp:  "C." + ctype,
			CTyp:    ctype,
		},
		Marshaler:       newDefaultMarshaler(v.GLibGetType),
		GoInterfaceName: v.Name,

		GoWrapBaseClassFunction: fmt.Sprintf("unsafeWrap%s", v.Name),

		BaseConversions: newDefaultBaseConversions(v.Name),

		gir: v,
	}

	return i
}

func (in *Interface) resolve(e *env) bool {
	e = e.sub("interface", in.gir.CType)

	if in.gir.GLibTypeStruct != "" {
		ns, typeStructType := e.findTypeByGIRName(in.gir.GLibTypeStruct)

		if ns != nil {
			e.logger.Warn("type struct is foreign", "namespace", ns.Name)
			return false
		}

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

	ns, parent := e.findTypeByGIRName("GObject.Object")

	if parent == nil {
		e.logger.Error("GObject.Object not found")
		return false
	}

	if _, ok := parent.(*Class); !ok {
		return false
	}

	in.Parent = CouldBeForeign[*Class]{
		Namespace: ns,
		Type:      parent.(*Class),
	}

	for _, prereq := range in.gir.Prerequisites {
		ns, inter := e.findTypeByGIRName(prereq.Name)

		if inter == nil {
			e.logger.Info("interface prerequesite not found", "interface", prereq.Name)
			return false
		}

		switch inter.(type) {
		case *Class, *Interface:
		default:
			e.logger.Warn("prerequesite is not a class or an interface", "prerequesite", inter.GIRName(), "actual", reflect.TypeOf(parent).String())
			return false
		}

		in.Prerequesite = append(in.Prerequesite, CouldBeForeign[Type]{
			Namespace: ns,
			Type:      inter,
		})
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
