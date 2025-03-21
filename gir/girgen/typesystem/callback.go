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
}

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

			GlibGetTypeFn: "",
		},
		// e.g. _gotk4_gtk4_AssistantPageFunc
		TrampolineName: fmt.Sprintf("%s_%s", e.trampolinePrefix(), goType),
		Parameters:     nil,
		gir:            v,
	}
}

func (cb *Callback) resolveParameters(e *env) bool {
	params := NewCallableParameters(e, cb.gir.CallableAttrs)

	if params == nil {
		return false
	}

	cb.Parameters = params

	return true
}
