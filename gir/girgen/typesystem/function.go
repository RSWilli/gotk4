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

func DeclareFunction(ctx context, v gir.Function) *CallableSignature {
	if ctx.skipType(v) {
		return nil
	}

	params := NewCallableParameters(ctx, v.CallableAttrs)

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

func NewMethod(ctx context, v gir.Method) *CallableSignature {
	if ctx.skipType(v) {
		return nil
	}

	params := NewCallableParameters(ctx, v.CallableAttrs)

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

func DeclareConstructor(ctx context, v gir.Constructor) *CallableSignature {
	if ctx.skipType(v) {
		return nil
	}

	params := NewCallableParameters(ctx, v.CallableAttrs)

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
