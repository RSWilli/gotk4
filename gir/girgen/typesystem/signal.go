package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
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
		for _, param := range v.Parameters.Parameters {
			ns, t := e.findAnyType(param.AnyType)

			if t == nil {
				e.logger.Warn("type not found", "type", debugCTypeFromAnytype(param.AnyType))
				return nil
			}

			pointers := CountCTypePointers(CTypeFromAnytype(param.AnyType))

			transfer := TransferOwnership(param.TransferOwnership.TransferOwnership)

			if transfer == "" {
				transfer = TransferNone
			}

			s.Parameters = append(s.Parameters, &Param{
				Doc:    NewParamDoc(param.ParameterAttrs),
				GoName: strcases.SnakeToGo(false, param.Name),
				Type: CouldBeForeign[Type]{
					Namespace: ns,
					Type:      t,
				},
				TransferOwnership: transfer,
				CTypePointers:     pointers,
			})
		}
	}

	return s
}
