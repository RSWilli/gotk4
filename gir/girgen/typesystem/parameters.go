package typesystem

import (
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type CallbackParamScope string

const (
	CallbackParamScopeCall  CallbackParamScope = "call"
	CallbackParamScopeAsync CallbackParamScope = "async"

	// CallbackParamScopeNotified must be accompanied by a Destroy parameter.
	CallbackParamScopeNotified CallbackParamScope = "notified"
	CallbackParamScopeForever  CallbackParamScope = "forever"
)

type TransferOwnership string

const (
	TransferNone      TransferOwnership = "none"
	TransferFull      TransferOwnership = "full"
	TransferBorrow    TransferOwnership = "borrow"
	TransferContainer TransferOwnership = "container"
)

type Param struct {
	Doc ParamDoc

	CName  string
	GoName string

	// Type is the type of the parameter. It is missing a pointer if this param is an out param that isn't CallerAllocates
	Type Type

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

	// BorrowFrom is a reference to the (instance) param that this wants to borrow from. In code generation terms this means
	// that we need to connect this (return) param to the instance param so that the GC wont clean it up early
	BorrowFrom *Param

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

func (p *Param) GoParamDeclaration() string {
	return fmt.Sprintf("%s %s", p.GoName, p.GoParamType())
}

func (p *Param) GoParamType() string {
	switch UnderlyingType(p.Type).(type) {
	case *Class:
		return ClassGoInterfaceName(p.Type)
	case *Interface:
		return InterfaceGoInterfaceName(p.Type)
	default:
		return p.Type.GoType()
	}
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

func validParam(param *Param, currentType Type) bool {
	if param.Implicit || param.Skip {
		return true
	}

	switch t := currentType.(type) {
	case *ForeignType:
		return validParam(param, t.Type)
	case *PointerType:
		return t.Pointers == 1 && validPointerParam(param, t.Base)
	case *Record, *Class:
		return false // needs one pointer
	case *Alias:
		return validParam(param, t.AliasedType)
	case *Callback:
		// closure must exist, and if notified then destroy must exist
		return param.Closure != nil && (param.Scope != CallbackParamScopeNotified || param.Destroy != nil)
	default:
		return true
	}
}

func validPointerParam(param *Param, currentType Type) bool {
	switch t := currentType.(type) {
	case *PointerType:
		panic("pointer to pointer type?")
	case *ForeignType:
		return validPointerParam(param, t.Type)
	case *Callback:
		return false
	case *Record, *Class:
		return true
	case *Alias:
		return validPointerParam(param, t.AliasedType)
	default:
		return true
	}
}

// validForGoBindings checks some preconditions to determine if we can convert all arguments to
func (p *Parameters) validForGoBindings() bool {
	for _, p := range p.CParameters() {
		if !validParam(p, p.Type) {
			return false
		}
	}

	return true
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
	params := &Parameters{
		Doc: NewDoc(&v.InfoAttrs, &v.InfoElements),
	}

	if v.Parameters != nil {
		if v.Parameters.InstanceParameter != nil {
			t := e.findAnyType(v.Parameters.InstanceParameter.AnyType)

			if t == nil {
				e.logger.Warn("instance param type not found", "ctype", debugCTypeFromAnytype(v.Parameters.InstanceParameter.AnyType))
				return nil
			}

			params.InstanceParam = &Param{
				Doc: NewParamDoc(v.Parameters.InstanceParameter.ParameterAttrs),

				CName:             "carg0",
				GoName:            strcases.ParamNameToGo(v.Parameters.InstanceParameter.Name),
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

		for i, p := range v.Parameters.Parameters {
			if p.Direction == "inout" {
				e.logger.Warn("FIXME: skipping inout param")
				return nil
			}
			paramType := p.AnyType

			if p.Direction == "out" && !p.CallerAllocates && !canBeOutParamType(paramType) {
				e.logger.Warn("ignoring out param type without enough pointers", "ctype", debugCTypeFromAnytype(paramType))
				return nil
			}

			if p.Direction == "out" && !p.CallerAllocates {
				// decrease the pointers, the last pointer will be added by the generator
				// when passing the value to the function
				paramType = decreaseAnyTypePointers(paramType)
			}

			t := e.findAnyType(paramType)

			if t == nil {
				e.logger.Warn("type not found", "ctype", debugCTypeFromAnytype(paramType))
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
				GoName:            strcases.ParamNameToGo(p.Name),
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
		for i, p := range v.Parameters.Parameters {
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

	if v.ReturnValue != nil {
		t := e.findAnyType(v.ReturnValue.AnyType)

		if t == nil {
			e.logger.Warn("return type not found", "ctype", debugCTypeFromAnytype(v.ReturnValue.AnyType))
			return nil
		}

		// https://gi.readthedocs.io/en/latest/annotations/giannotations.html#default-annotations
		transfer := TransferOwnership(v.ReturnValue.TransferOwnership.TransferOwnership)

		if transfer == "" {
			transfer = TransferFull
		}

		ret := &Param{
			Doc:               NewReturnDoc(v.ReturnValue),
			CName:             "cret",
			GoName:            "ret",
			Direction:         "return",
			Type:              t,
			TransferOwnership: transfer,
		}

		if transfer == TransferBorrow {
			if params.InstanceParam == nil {
				e.logger.Error("can't borrow without an instance param")
				return nil
			}

			ret.BorrowFrom = params.InstanceParam
		}

		params.CReturn = ret

		if t.GIRName() != "none" {
			params.GoReturns = append(params.GoReturns, ret)
		}

	}

	if v.Throws {
		// if a callable throws then it has a GError** as the last param, which is a nullable
		// out param
		throwType := e.findTypeByGIRName("GLib.Error")

		if throwType == nil {
			e.logger.Warn("GLib.Error type not found, ignoring throwing function")
			return nil
		}

		throwType = IncreasePointers(throwType, 1)

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

	if !params.validForGoBindings() {
		e.logger.Warn("parameters not valid for go bindings")
		return nil
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

// GoParamTypes differs from GoTypes in that class and interface params don't return the
// pointer to the struct type, but instead the go interface name
func (pl ParamList) GoParamTypes() string {
	decls := make([]string, 0, len(pl))

	for _, p := range pl {
		if p.Skip || p.Implicit {
			continue
		}

		decls = append(decls, p.GoParamType())
	}

	return strings.Join(decls, ", ")
}

// GoParamDeclarations differs from GoDeclarations in that class and interface params don't return the
// pointer to the struct type, but instead the go interface name
func (pl ParamList) GoParamDeclarations() string {
	decls := make([]string, 0, len(pl))

	for _, p := range pl {
		if p.Skip || p.Implicit {
			continue
		}

		decls = append(decls, p.GoParamDeclaration())
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
