package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
)

type CallableSignature struct {
	GoName  string
	CName   string
	CGoName string
	*Parameters
}

func DeclareFunction(ns *Namespace, v gir.Function) *CallableSignature {
	if !v.IsIntrospectable() {
		return nil
	}

	params := NewCallableParameters(ns, v.CallableAttrs)

	if params == nil {
		// log.Printf("could not create parameters for function %s\n", v.CIdentifier)
		return nil
	}

	return &CallableSignature{
		GoName:     v.Name,
		CName:      v.CIdentifier,
		CGoName:    "C." + v.CIdentifier,
		Parameters: params,
	}
}

func NewMethod(ns *Namespace, v gir.Method) *CallableSignature {
	if !v.IsIntrospectable() {
		return nil
	}

	params := NewCallableParameters(ns, v.CallableAttrs)

	if params == nil {
		// log.Printf("could not create parameters for method %s\n", v.CIdentifier)
		return nil
	}

	return &CallableSignature{
		GoName:     v.Name,
		CName:      v.CIdentifier,
		CGoName:    "C." + v.CIdentifier,
		Parameters: params,
	}
}

func DeclareConstructor(ns *Namespace, v gir.Constructor) *CallableSignature {
	if !v.IsIntrospectable() {
		return nil
	}

	params := NewCallableParameters(ns, v.CallableAttrs)

	if params == nil {
		// log.Printf("could not create parameters for constructor %s\n", v.CIdentifier)
		return nil
	}

	return &CallableSignature{
		GoName:     v.Name,
		CName:      v.CIdentifier,
		CGoName:    "C." + v.CIdentifier,
		Parameters: params,
	}
}
