package typesystem

import (
	"fmt"
	"log"
	"strings"

	"github.com/diamondburned/gotk4/gir"
)

type CallbackParamScope string

const (
	CallbackParamScopeCall  CallbackParamScope = "call"
	CallbackParamScopeAsync CallbackParamScope = "async"

	// CallbackParamScopeNofified must be accompanied by a Destroy parameter.
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

	// CallerAllocates means that the out param allocation must be provided by the caller.
	//
	// in practise this means that we need one more pointer if the param direction is out and this is false
	CallerAllocates bool

	// Implicit declares that this param is referenced by another param, either through closure, destroy or
	// array size. It will be omitted in the go call, because it will get it's value from another source.
	Implicit bool

	Closure *Param
	Destroy *Param
}

func (p *Param) CDeclaration() string {
	return fmt.Sprintf("%s %s", p.CName, p.Type.CType())
}

func (p *Param) CGoDeclaration() string {
	return fmt.Sprintf("%s %s", p.CName, p.Type.CGoType())
}

func (p *Param) GoDeclaration() string {
	return fmt.Sprintf("%s %s", p.GoName, p.Type.GoType())
}

type Parameters struct {
	Doc

	// CReturn contains the param that the c function returns
	CReturn *Param

	// CParameters contains the params that the c function requires
	CParameters ParamList

	// GoReceiver contains the C instance param, which will be used as a method receiver
	// for the go function.
	GoReceiver *Param

	// GoReturns containts the return values of the Go function. C Params that are declared as "out"
	// will also get moved here, so this may differ from CParameters, but must contain pointers to the same
	// objects.
	GoReturns ParamList

	// GoParameters will contain the parameters of the go function. C Params that are declared as "out"
	// will not be in this list.
	GoParameters ParamList
}

func NewCallableParameters(ns context, v gir.CallableAttrs) *Parameters {
	params := NewParameters(ns, v.Parameters, v.ReturnValue, v.Throws)

	if params != nil {
		params.Doc = NewDoc(&v.InfoAttrs, &v.InfoElements)
	}

	return params
}

func NewParameters(ctx context, girparams *gir.Parameters, ret *gir.ReturnValue, throws bool) *Parameters {
	params := &Parameters{}

	if girparams != nil {
		if girparams.InstanceParameter != nil {
			t := ctx.findAnyType(girparams.InstanceParameter.AnyType)

			if t == nil {
				return nil
			}

			params.GoReceiver = &Param{
				Doc: NewParamDoc(girparams.InstanceParameter.ParameterAttrs),

				CName:  "carg0",
				GoName: "arg0", // TODO: find a better go name
				Type:   t,
			}

			// instance param is always first C param
			params.CParameters = append(params.CParameters, params.GoReceiver)
		}

		for i, p := range girparams.Parameters {
			paramType := p.AnyType

			if p.Direction == "out" && !p.CallerAllocates && !canBeOutParamType(paramType) {
				log.Printf("ignoring out param type without enough pointers: %s", debugCTypeFromAnytype(paramType))
				return nil
			}

			if p.Direction == "out" && !p.CallerAllocates {
				// decrease the pointers, the last pointer will be added by the generator
				// when passing the value to the function
				paramType = decreasePointers(paramType)
			}

			t := ctx.findAnyType(paramType)

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
		for i, p := range girparams.Parameters {
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

	if throws {
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

	if ret != nil {
		t := ctx.findAnyType(ret.AnyType)

		if t == nil {
			return nil
		}

		if t.GIRName() != "none" {
			ret := &Param{
				Doc:    NewReturnDoc(ret),
				CName:  "cret",
				GoName: "ret",
				Type:   t,
			}

			params.CReturn = ret
			params.GoReturns = append(params.GoReturns, ret)
		}

	}

	ctx.sortGoParams(params.GoParameters)
	ctx.sortGoReturns(params.GoReturns)

	return params
}

// canBeOutParamType returns true if the AnyType has enough pointers to be an "out" parameter
//
// if this is false then it's most likely that the documentation is wrong
func canBeOutParamType(t gir.AnyType) bool {
	switch {
	case t.Array != nil:
		// needs a pointer to the array, resulting in at least 2 pointers
		return CountPointers(t.Array.CType) >= 2
	case t.Type != nil:
		// needs at least a pointer to the value.
		return CountPointers(t.Type.CType) >= 1
	default:
		panic("invalid anytype")
	}
}

type ParamList []*Param

func (pl ParamList) GoDeclarations() string {
	decls := make([]string, 0, len(pl))

	for _, p := range pl {
		decls = append(decls, p.GoDeclaration())
	}

	return strings.Join(decls, ", ")
}

func (pl ParamList) GoIdentifiers() string {
	decls := make([]string, 0, len(pl))

	for _, p := range pl {
		decls = append(decls, p.GoName)
	}

	return strings.Join(decls, ", ")
}

func (pl ParamList) GoTypes() string {
	decls := make([]string, 0, len(pl))

	for _, p := range pl {
		decls = append(decls, p.Type.GoType())
	}

	return strings.Join(decls, ", ")
}

func (pl ParamList) CGoDeclarations() string {
	decls := make([]string, 0, len(pl))

	for _, p := range pl {
		decls = append(decls, p.CGoDeclaration())
	}

	return strings.Join(decls, ", ")
}
