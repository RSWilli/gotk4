package typesystem

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type Callback struct {
	baseType

	TrampolineName string

	// Parameters can be nil, if so then the generator should refuse to
	// output the referencing function/struct whatever
	*Parameters
}

// NewCallback declares a new callback. This way the type can be resolved by others, but the referenced parameters
// have to be resolved later, because the callback params could be referencing other record types
func NewCallback(ns *Namespace, v gir.Callback) *Callback {
	if !v.IsIntrospectable() {
		return nil
	}

	return &Callback{
		baseType: baseType{
			girName: v.Name,
			goType:  strcases.PascalToGo(v.Name),
			cGoType: "C." + v.CType,
			cType:   v.CType,
		},
		TrampolineName: fmt.Sprintf("_gotk4_%s%d_%s", ns.GoName, ns.v.majorVersion, v.Name),
		Parameters:     nil, // resolved later
	}
}

func (cb *Callback) resolveParameters(ns *Namespace, v gir.Callback) {
	cb.Parameters = NewParameters(ns, v.CallableAttrs)
}
