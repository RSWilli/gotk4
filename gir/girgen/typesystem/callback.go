package typesystem

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type Callback struct {
	BaseType

	TrampolineName string

	// gir is used to resolve the parameters after the callback has been declared
	gir gir.Callback

	*Parameters

	UserdataParam *Param
}

var _ checkedParameterType = (*Callback)(nil)

// DeclareCallback declares a new callback. This way the type can be resolved by others, but the referenced parameters
// have to be resolved later, because the callback params could be referencing other record types
func DeclareCallback(e *env, v gir.Callback) *Callback {
	if !v.IsIntrospectable() {
		return nil
	}

	if e.skip(nil, v) {
		return nil
	}

	goType := strcases.PascalToGo(v.Name)

	return &Callback{
		BaseType: BaseType{
			GirName: v.Name,
			GoTyp:   goType,
			CGoTyp:  "C." + v.CType,
			CTyp:    v.CType,
		},
		// e.g. _gotk4_gtk4_AssistantPageFunc
		TrampolineName: fmt.Sprintf("%s_%s", e.trampolinePrefix(), goType),
		Parameters:     nil,
		gir:            v,
	}
}

func (cb *Callback) resolveParameters(e *env) resolvedState {
	e = e.sub("callback", cb.gir.CType)

	params, state := NewCallableParameters(e, cb.gir.CallableAttrs)

	if state == notResolvable {
		return notResolvable
	}

	if state == maybeResolvable {
		return maybeResolvable
	}

	if params == nil {
		panic("nil params received even though valid")
	}

	for _, p := range params.CParameters() {
		if p.IsUserData {
			cb.UserdataParam = p
			break
		}
	}

	if cb.UserdataParam == nil {
		e.logger.Warn("skipping callback without user data closure")
		return notResolvable
	}

	cb.Parameters = params

	return okResolved
}

// maxPointersAllowed implements maxPointerConstrainedType.
func (a *Callback) maxPointersAllowed() int {
	return 0
}

// allowedTypeForParam implements allowedParameter.
func (cb *Callback) allowedTypeForParam(param *Param) bool {
	return param.CTypePointers == 0 && param.Closure != nil && (param.Scope != CallbackParamScopeNotified || param.Destroy != nil)
}
