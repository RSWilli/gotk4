package typesystem

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type VirtualMethod struct {
	// Parent is the class or interface that this virtual method belongs to.
	Parent ConvertibleType

	// TrampolineName is the name of the trampoline function that needs to be
	// called when the virtual function was overridden.
	TrampolineName string

	// ParentTrampolineName is the C function name that is used to call the C function pointer
	// of the virtual method of the parent class. This is needed because we cannot cast c function pointers
	// to callable functions.
	ParentTrampolineName string

	Invoker *Field

	// GoName is the name of the override in the Overrides struct.
	GoName string

	// ParentName is the name of the method on the instance that calls the default implementation
	// on the parent class.
	ParentName string

	*Parameters
}

func NewVirtualMethod(e *env, parent ConvertibleType, typestruct *Record, v gir.VirtualMethod) *VirtualMethod {
	if !v.IsIntrospectable() {
		return nil
	}

	if e.skip(parent, v) {
		return nil
	}

	e = e.sub("virtual method", v.Name)

	// e.g. _gotk4_gtk4_AccessibleText_get_contents
	trampoline := fmt.Sprintf("%s_%s_%s", e.trampolinePrefix(), parent.GoType(1), v.Name)

	// e.g. _gotk4_gtk4_AccessibleText_virtual_get_contents
	parentTrampoline := fmt.Sprintf("%s_%s_virtual_%s", e.trampolinePrefix(), parent.GoType(1), v.Name)

	params, _ := NewCallableParameters(e, v.CallableAttrs)

	if params == nil {
		return nil
	}

	field := findTypeStructField(v, typestruct)

	if field == nil {
		e.logger.Warn("could not find type struct field name")
		return nil
	}

	goname := strcases.SnakeToGo(true, field.CIndentifier())

	return &VirtualMethod{
		Parent:               parent,
		TrampolineName:       trampoline,
		ParentTrampolineName: parentTrampoline,

		Invoker:    field,
		GoName:     goname,
		ParentName: fmt.Sprintf("Parent%s", goname),
		Parameters: params,
	}
}

func findTypeStructField(virtual gir.VirtualMethod, ts *Record) *Field {
	name := virtual.Name
	// if virtual.Invoker != "" {
	// 	name = virtual.Invoker
	// }

	for _, field := range ts.Fields {
		if field.CIndentifier() == name {
			return field
		}
	}

	return nil
}
