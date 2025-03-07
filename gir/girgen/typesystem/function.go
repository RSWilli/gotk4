package typesystem

import "github.com/diamondburned/gotk4/gir"

type Callable struct {
	GoName  string
	CGoName string
	*Parameters
}

func NewFunction(ns *Namespace, v gir.Function) *Callable {
	if !v.IsIntrospectable() {
		return nil
	}

	params := NewParameters(ns, v.CallableAttrs)

	if params == nil {
		return nil
	}

	return &Callable{
		GoName:     v.Name,
		CGoName:    v.CIdentifier,
		Parameters: params,
	}
}

func NewMethod(ns *Namespace, v gir.Method) *Callable {
	if !v.IsIntrospectable() {
		return nil
	}

	params := NewParameters(ns, v.CallableAttrs)

	if params == nil {
		return nil
	}

	return &Callable{
		GoName:     v.Name,
		CGoName:    v.CIdentifier,
		Parameters: params,
	}
}

func NewConstructor(ns *Namespace, v gir.Constructor) *Callable {
	if !v.IsIntrospectable() {
		return nil
	}

	params := NewParameters(ns, v.CallableAttrs)

	if params == nil {
		return nil
	}

	return &Callable{
		GoName:     v.Name,
		CGoName:    v.CIdentifier,
		Parameters: params,
	}
}
