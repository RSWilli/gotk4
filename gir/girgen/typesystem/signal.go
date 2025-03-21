package typesystem

import (
	"log"

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

	*Parameters
}

func NewSignal(e *env, parent Type, v gir.Signal) *Signal {
	if e.skip(parent, v) {
		return nil
	}

	params := NewParameters(e, v.Parameters, v.ReturnValue, false)

	if params == nil {
		log.Printf("could not create parameters for signal %s", v.Name)
		return nil
	}

	return &Signal{
		Doc:        NewDoc(nil, &v.InfoElements),
		Name:       v.Name,
		Detailed:   v.Detailed,
		When:       SignalWhen(v.When),
		Action:     v.Action,
		NoHooks:    v.NoHooks,
		NoRecurse:  v.NoRecurse,
		Parameters: params,
	}
}
