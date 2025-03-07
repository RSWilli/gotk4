package typesystem

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir"
)

type CallbackParamScope string

const (
	CallbackParamScopeCall     CallbackParamScope = "call"
	CallbackParamScopeAsync    CallbackParamScope = "async"
	CallbackParamScopeNofified CallbackParamScope = "nofified"
	CallbackParamScopeForever  CallbackParamScope = "forever"
)

type TransferOwnership string

const (
	TransferNone      TransferOwnership = "none"
	TransferFull      TransferOwnership = "full"
	TransferContainer TransferOwnership = "container"
)

type Param struct {
	Doc ParamDoc

	CName  string
	GoName string
	Type   Type

	// Skip signifies that the parameter should be skipped in the go call
	// this happens with params that are only useful in C.
	Skip bool

	TransferOwnership TransferOwnership
	Optional          bool
	Nullable          bool
	Direction         string
	Scope             CallbackParamScope
	CallerAllocates   bool

	// Implicit declares that this param is referenced by another param, either through closure, destroy or
	// array size. It will be omitted in the go call, because it will get it's value from another source.
	Implicit bool

	Closure *Param
	Destroy *Param
}

type Parameters struct {
	Doc Doc

	CReturn     *Param
	CParameters []*Param

	GoReceiver   *Param
	GoReturns    []*Param
	GoParameters []*Param
}

func NewParameters(ns *Namespace, v gir.CallableAttrs) *Parameters {
	params := &Parameters{
		Doc: NewDoc(&v.InfoAttrs, &v.InfoElements),
	}

	if v.Parameters != nil {
		if v.Parameters.InstanceParameter != nil {
			t := ns.findAnyType(v.Parameters.InstanceParameter.AnyType)

			if t == nil {
				return nil
			}

			params.GoReceiver = &Param{
				Doc: NewParamDoc(v.Parameters.InstanceParameter.ParameterAttrs),

				CName:  "carg0",
				GoName: "arg0", // TODO: find a better go name
				Type:   t,
			}

			// instance param is always first C param
			params.CParameters = append(params.CParameters, params.GoReceiver)
		}

		for i, p := range v.Parameters.Parameters {
			t := ns.findAnyType(p.AnyType)

			if t == nil {
				return nil
			}

			param := &Param{
				Doc:               NewParamDoc(p.ParameterAttrs),
				CName:             fmt.Sprintf("carg%d", i+1),
				GoName:            fmt.Sprintf("arg%d", i+1), // TODO: find a better go name
				Type:              t,
				TransferOwnership: TransferOwnership(p.TransferOwnership.TransferOwnership),
				Skip:              p.Skip || p.Direction == "out",
				Optional:          p.Optional,
				Nullable:          p.Nullable,
				Direction:         p.Direction,
				Scope:             CallbackParamScope(p.Scope),
				CallerAllocates:   p.CallerAllocates,
				Implicit:          false,
				Closure:           nil,
				Destroy:           nil,
			}

			params.CParameters = append(params.CParameters, param)
			params.GoParameters = append(params.GoParameters, param)

			if p.Direction == "out" {
				// value will be a return, must add to go params still to keep the index intact for skipping
				params.GoReturns = append(params.GoReturns, param)
			}
		}

		// mark the implicit params. The idx is the index in c parameters, with a given instance param
		for i, p := range v.Parameters.Parameters {
			param := params.GoParameters[i]
			if p.Closure != nil {
				param.Closure = params.CParameters[*p.Closure]
				param.Closure.Implicit = true
				continue
			}
			if p.Destroy != nil {
				param.Destroy = params.CParameters[*p.Destroy]
				param.Destroy.Implicit = true
				continue
			}
			if p.AnyType.Array != nil && p.AnyType.Array.Length != nil {
				// type must be an array type here:
				param.Type.(*Array).Length = params.CParameters[*p.Array.Length]
				params.CParameters[*p.Array.Length].Implicit = true
				continue
			}
		}
	}

	if v.Throws {
		throwParam := &Param{
			Doc:               ParamDoc{},
			CName:             "_cerr",
			GoName:            "_goerr",
			Type:              TypeError,
			Skip:              false,
			TransferOwnership: TransferFull,
			Optional:          true,
			Nullable:          true,
			Direction:         "out",
		}

		params.CParameters = append(params.CParameters, throwParam)
		params.GoReturns = append(params.GoReturns, throwParam)
	}

	if v.ReturnValue != nil {
		t := ns.findAnyType(v.ReturnValue.AnyType)

		if t == nil {
			return nil
		}

		ret := &Param{
			Doc:    NewReturnDoc(v.ReturnValue),
			CName:  "cret",
			GoName: "ret",
			Type:   t,
		}

		params.CReturn = ret
		params.GoReturns = append(params.GoReturns, ret)
	}

	return params
}
