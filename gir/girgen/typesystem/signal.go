package typesystem

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir"
)

type SignalWhen string

const (
	SignalWhenFirst   = "first"
	SignalWhenLast    = "last"
	SignalWhenCleanup = "cleanup"
)

type Signal struct {
	Doc       Doc
	Name      string
	Detailed  bool
	When      SignalWhen
	Action    bool
	NoHooks   bool
	NoRecurse bool

	Parameters ParamList
}

func NewSignal(e *env, parent Type, v gir.Signal) *Signal {
	if e.skip(parent, v) {
		return nil
	}

	return nil

	e = e.sub("signal", v.Name)

	s := &Signal{
		Doc:       NewDoc(nil, &v.InfoElements),
		Name:      v.Name,
		Detailed:  v.Detailed,
		When:      SignalWhen(v.When),
		Action:    v.Action,
		NoHooks:   v.NoHooks,
		NoRecurse: v.NoRecurse,
	}

	// Signal params often don't specify a Ctype, especially for records and classes etc. So we
	// cannot use the normal CallableParameter resolution here

	if v.Parameters != nil {
		for i, param := range v.Parameters.Parameters {
			t := e.findAnyType(param.AnyType)

			if t == nil {
				e.logger.Warn("type not found", "type", debugCTypeFromAnytype(param.AnyType))
				return nil
			}

			if isPointerMandatory(t) {
				t = SetPointers(t, 1)
			}

			transfer := TransferOwnership(param.TransferOwnership.TransferOwnership)

			if transfer == "" {
				transfer = TransferNone
			}

			s.Parameters = append(s.Parameters, &Param{
				Doc:               NewParamDoc(param.ParameterAttrs),
				GoName:            fmt.Sprintf("arg%d", i),
				Type:              t,
				TransferOwnership: transfer,
			})
		}
	}

	return s
}

// addPointerIfManatory adds a pointer to a type where we don't know from the ctype how many pointers are needed
// but some types always need a pointer
func isPointerMandatory(t Type) bool {
	switch t := t.(type) {
	case *PointerType:
		return false // not mandatory, because we already have one
	case *ForeignType:
		return isPointerMandatory(t.Type)
	case *Record, *Class:
		return true
	default:
		return false
	}
}
