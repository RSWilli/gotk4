package typesystem

import (
	"fmt"
	"log"

	"github.com/diamondburned/gotk4/gir"
)

type VirtualMethod struct {
	TrampolineName string

	Invoker *Field

	*Parameters
}

func NewVirtualMethod(ns *Namespace, parent Type, typestruct *Record, v gir.VirtualMethod) *VirtualMethod {
	if !v.IsIntrospectable() {
		return nil
	}

	// e.g. _gotk4_gtk4_AccessibleText_virtual_get_contents
	tramp := fmt.Sprintf("_gotk4_%s%d_%s_%s", ns.GoName, ns.v.majorVersion, parent.GoType(), v.Name)

	params := NewCallableParameters(ns, v.CallableAttrs)

	if params == nil {
		// log.Printf("could not create parameters for virtual method %s\n", tramp)
		return nil
	}

	field := findTypeStructField(v, typestruct)

	if field == nil {
		log.Printf("could not find type struct field name for %s\n", tramp)
		return nil
	}

	return &VirtualMethod{
		TrampolineName: tramp,

		Invoker:    field,
		Parameters: params,
	}
}

func findTypeStructField(virtual gir.VirtualMethod, ts *Record) *Field {
	name := virtual.Name
	// if virtual.Invoker != "" {
	// 	name = virtual.Invoker
	// }

	for _, field := range ts.Fields {
		if field.CName == name {
			return field
		}
	}

	return nil
}
