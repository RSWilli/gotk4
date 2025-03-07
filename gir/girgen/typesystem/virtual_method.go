package typesystem

import "github.com/diamondburned/gotk4/gir"

type VirtualMethod struct {
	TrampolineName string

	Invoker *Field

	// Parameters can be nil, if so then the generator should refuse to
	// output the referencing function/struct whatever
	*Parameters
}

func NewVirtualMethod(ns *Namespace, typestruct *Record, v gir.VirtualMethod) *VirtualMethod {
	if !v.IsIntrospectable() {
		return nil
	}

	params := NewParameters(ns, v.CallableAttrs)

	if params == nil {
		return nil
	}

	field := findTypeStructField(v, typestruct)

	if field == nil {
		return nil
	}

	return &VirtualMethod{
		Invoker:    field,
		Parameters: params,
	}
}

func findTypeStructField(virtual gir.VirtualMethod, ts *Record) *Field {
	name := virtual.Name
	if virtual.Invoker != "" {
		name = virtual.Invoker
	}

	for _, field := range ts.Fields {
		if field.CName == name {
			return field
		}
	}

	return nil
}
