package typesystem

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type Callback struct {
	BaseType

	TrampolineName string

	// Parameters can be nil, if so then the generator should refuse to
	// output the referencing function/struct whatever
	*Parameters
}

// DeclareCallback declares a new callback. This way the type can be resolved by others, but the referenced parameters
// have to be resolved later, because the callback params could be referencing other record types
func DeclareCallback(ctx context, v gir.Callback) *Callback {
	if ctx.skipType(v) {
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
		TrampolineName: fmt.Sprintf("_gotk4_%s%d_%s", ctx.goName(), ctx.majorVersion(), goType),
		Parameters:     nil, // resolved later
	}
}

func (cb *Callback) resolveParameters(ns context, v gir.Callback) {
	cb.Parameters = NewCallableParameters(ns, v.CallableAttrs)
}
