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
	// Nullable means that NULL can be returned
	Nullable  bool
	Direction string
	Scope     CallbackParamScope

	// CallerAllocates means that the out param allocation must be provided by the caller.
	//
	// in practise this means that we need one more pointer if the param direction is out and this is false
	CallerAllocates bool

	// Implicit declares that this param is referenced by another param, either through closure, destroy or
	// array size. It will be omitted in the go call, because it will get it's value from another source.
	Implicit bool

	// Closure is a pointer to the implicit param that takes the function pointer that will be called
	Closure *Param

	// Destroy is a pointer to the implicit param of the destroy notify callback.
	Destroy *Param

	// Optional signifies that an out or inout param can be NULL to ignore it. This is not useful
	// for moving the out params to go return values, so this is only here for completeness sake
	Optional bool
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

	// parameters contains the params that the gir declares. The actual CParameters need the InstanceParam prepended, hence this is private
	//
	// the parameter references (destroy, closure and array length) are relative indices in this list
	parameters ParamList

	// InstanceParam contains the C instance param, which will be used as a method receiver
	// for the go function. It is also always the first parameter for the c function call
	InstanceParam *Param

	// GoReturns containts the return values of the Go function. C Params that are declared as "out"
	// will also get moved here, so this may differ from Parameters, but must contain pointers to the same
	// objects.
	GoReturns ParamList

	// GoParameters will contain the parameters of the go function. C Params that are declared as "out"
	// will not be in this list.
	GoParameters ParamList
}

// CParameters returns the param list for the c call, since the instance param is always the first param if set
func (p *Parameters) CParameters() ParamList {
	if p.InstanceParam == nil {
		return p.parameters
	}

	params := make(ParamList, 0, len(p.parameters)+1)

	params = append(params, p.InstanceParam)
	params = append(params, p.parameters...)

	return params
}

// CGoReturn returns the CReturn if it is set and not void/none
func (p *Parameters) CGoReturn() *Param {
	if p.CReturn == nil {
		return nil
	}

	if p.CReturn.Type == Void {
		return nil
	}

	return p.CReturn
}

func NewCallableParameters(e *env, v gir.CallableAttrs) *Parameters {
	params := NewParameters(e, v.Parameters, v.ReturnValue, v.Throws)

	if params != nil {
		params.Doc = NewDoc(&v.InfoAttrs, &v.InfoElements)
	}

	return params
}

func NewParameters(e *env, girparams *gir.Parameters, ret *gir.ReturnValue, throws bool) *Parameters {
	params := &Parameters{}

	if girparams != nil {
		if girparams.InstanceParameter != nil {
			t := e.findAnyType(girparams.InstanceParameter.AnyType)

			if t == nil {
				return nil
			}

			params.InstanceParam = &Param{
				Doc: NewParamDoc(girparams.InstanceParameter.ParameterAttrs),

				CName:             "carg0",
				GoName:            "arg0", // TODO: find a better go name
				Type:              t,
				TransferOwnership: TransferNone,
				Skip:              false,
				Optional:          false,
				Nullable:          false,
				Direction:         "in",
				Scope:             "call",
				CallerAllocates:   false,
				Implicit:          false,
				Closure:           nil,
				Destroy:           nil,
			}
		}

		for i, p := range girparams.Parameters {
			if p.Direction == "inout" {
				log.Println("FIXME: skipping inout param")
				return nil
			}
			paramType := p.AnyType

			if p.Direction == "out" && !p.CallerAllocates && !canBeOutParamType(paramType) {
				log.Printf("ignoring out param type without enough pointers: %s", debugCTypeFromAnytype(paramType))
				return nil
			}

			if p.Direction == "out" && !p.CallerAllocates {
				// decrease the pointers, the last pointer will be added by the generator
				// when passing the value to the function
				paramType = decreaseAnyTypePointers(paramType)
			}

			t := e.findAnyType(paramType)

			if t == nil {
				return nil
			}

			direction := p.Direction

			if direction == "" {
				direction = "in"
			}

			scope := CallbackParamScope(p.Scope)

			if scope == "" {
				// https://gi.readthedocs.io/en/latest/annotations/giannotations.html
				scope = CallbackParamScopeCall
			}

			// https://gi.readthedocs.io/en/latest/annotations/giannotations.html#default-annotations
			transfer := TransferOwnership(p.TransferOwnership.TransferOwnership)

			if transfer == "" {
				switch direction {
				case "in":
					transfer = TransferFull
				case "out", "inout":
					if p.CallerAllocates {
						transfer = TransferNone
					}
				}
			}

			param := &Param{
				Doc:               NewParamDoc(p.ParameterAttrs),
				CName:             fmt.Sprintf("carg%d", i+1),
				GoName:            fmt.Sprintf("arg%d", i+1),
				Type:              t,
				TransferOwnership: transfer,
				Skip:              p.Skip,
				Optional:          p.Optional,
				Nullable:          p.Nullable,
				Direction:         direction,
				Scope:             scope,
				CallerAllocates:   p.CallerAllocates,
				Implicit:          false,
				Closure:           nil,
				Destroy:           nil,
			}

			params.parameters = append(params.parameters, param)

			if p.Direction == "out" {
				// value will be a return, must add to go params still to keep the index intact for skipping
				params.GoReturns = append(params.GoReturns, param)
			} else {
				params.GoParameters = append(params.GoParameters, param)
			}
		}

		// mark the implicit params. The idx is the index in c parameters, with a given instance param
		// a parameter may have multiple implicit params
		for i, p := range girparams.Parameters {
			param := params.parameters[i]
			if p.Closure != nil {
				param.Closure = params.parameters[*p.Closure]
				param.Closure.Implicit = true
			}
			if p.Destroy != nil {
				param.Destroy = params.parameters[*p.Destroy]
				param.Destroy.Implicit = true
			}
			if p.AnyType.Array != nil && p.AnyType.Array.Length != nil {
				// type must be an array type here, but can be a pointer to an array:

				switch param.Type.(type) {
				case *Array:
					param.Type.(*Array).Length = params.parameters[*p.Array.Length]
				case *PointerType:
					param.Type.(*PointerType).Base.(*Array).Length = params.parameters[*p.Array.Length]
				}

				params.parameters[*p.Array.Length].Implicit = true
			}
		}
	}

	if ret != nil {
		t := e.findAnyType(ret.AnyType)

		if t == nil {
			return nil
		}

		// https://gi.readthedocs.io/en/latest/annotations/giannotations.html#default-annotations
		transfer := TransferOwnership(ret.TransferOwnership.TransferOwnership)

		if transfer == "" {
			transfer = TransferFull
		}

		ret := &Param{
			Doc:               NewReturnDoc(ret),
			CName:             "cret",
			GoName:            "ret",
			Direction:         "return",
			Type:              t,
			TransferOwnership: transfer,
		}

		params.CReturn = ret

		if t.GIRName() != "none" {
			params.GoReturns = append(params.GoReturns, ret)
		}

	}

	if throws {
		// if a callable throws then it has a GError** as the last param, which is a nullable
		// out param
		throwType := e.findTypeByGIRName("GLib.Error")

		if throwType == nil {
			log.Println("error type not found, ignoring throwing function")
			return nil
		}

		throwType = IncreasePointers(throwType, 2)

		throwParam := &Param{
			Doc: ParamDoc{
				Name: "err",
				Doc:  "an error",
			},
			CName:             "_cerr",
			GoName:            "_goerr",
			Type:              throwType,
			Skip:              false,
			TransferOwnership: TransferFull,
			Optional:          true,
			Nullable:          true,
			Direction:         "out",
		}

		params.parameters = append(params.parameters, throwParam)
		params.GoReturns = append(params.GoReturns, throwParam)
	}

	e.sortGoParams(params.GoParameters)
	e.sortGoReturns(params.GoReturns)

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
		if p.Skip || p.Implicit {
			continue
		}

		decls = append(decls, p.GoDeclaration())
	}

	return strings.Join(decls, ", ")
}

func (pl ParamList) GoIdentifiers() string {
	decls := make([]string, 0, len(pl))

	for _, p := range pl {
		if p.Skip || p.Implicit {
			continue
		}

		decls = append(decls, p.GoName)
	}

	return strings.Join(decls, ", ")
}

func (pl ParamList) CIdentifiers() string {
	decls := make([]string, 0, len(pl))

	for _, p := range pl {
		decls = append(decls, p.GoName)
	}

	return strings.Join(decls, ", ")
}

func (pl ParamList) GoTypes() string {
	decls := make([]string, 0, len(pl))

	for _, p := range pl {
		if p.Skip || p.Implicit {
			continue
		}

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
