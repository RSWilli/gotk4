package typesystem

import (
	"strings"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type CallableSignature struct {
	Identifier
	*Parameters
}

func DeclareFunction(e *env, v gir.CallableAttrs) *CallableSignature {
	if !v.IsIntrospectable() {
		return nil
	}

	e = e.sub("function", v.CIdentifier)

	if v.ShadowedBy != "" || v.MovedTo != "" {
		e.logger.Debug("skipping because shadowed or moved")
		return nil
	}

	if e.skip(nil, v) {
		return nil
	}

	params, _ := NewCallableParameters(e, v)

	if params == nil {
		return nil
	}

	return &CallableSignature{
		Identifier: &baseIdentifier{
			cIndentifier:   v.CIdentifier,
			cGoIndentifier: "C." + v.CIdentifier,
			goIndentifier:  strcases.SnakeToGo(true, v.Name),
		},
		Parameters: params,
	}
}

// PrefixedIdentifier is an identifier that prefixes the parent type, so that renaming the
// parent struct reflects to renaming the constructors / methods. It has a special case for
// constructors, which are prefixed with "New" and the parent type name.
type PrefixedIdentifier struct {
	Parent         Type
	Girname        string
	GirCIdentifier string
}

// CGoIndentifier implements Identifier.
func (c *PrefixedIdentifier) CGoIndentifier() string {
	return "C." + c.GirCIdentifier
}

// CIndentifier implements Identifier.
func (c *PrefixedIdentifier) CIndentifier() string {
	return c.GirCIdentifier
}

// GoIndentifier turns the girname into a hopefully unique function name
//
// e.g. BufferList.new_sized -> NewBufferListSized
func (c *PrefixedIdentifier) GoIndentifier() string {
	pascal := strcases.SnakeToGo(true, c.Girname)

	noNew, ok := strings.CutPrefix(pascal, "New")

	parentTypeName := c.Parent.GoType(0)

	if ok {
		return "New" + parentTypeName + noNew
	}

	return parentTypeName + pascal
}

var _ Identifier = &PrefixedIdentifier{}

func DeclarePrefixedFunction(e *env, parent Type, v gir.CallableAttrs) *CallableSignature {
	if !v.IsIntrospectable() {
		return nil
	}

	e = e.sub("function", v.CIdentifier)

	if v.ShadowedBy != "" || v.MovedTo != "" {
		e.logger.Debug("skipping because shadowed or moved")
		return nil
	}

	if e.skip(parent, v) {
		return nil
	}

	params, _ := NewCallableParameters(e, v)

	if params == nil {
		return nil
	}

	return &CallableSignature{
		Identifier: &PrefixedIdentifier{
			Parent:         parent,
			Girname:        v.Name,
			GirCIdentifier: v.CIdentifier,
		},
		Parameters: params,
	}
}

func DeclareMethod(e *env, parent Type, v gir.Method) *CallableSignature {
	if !v.IsIntrospectable() {
		return nil
	}

	e = e.sub("method", v.CIdentifier)

	if v.ShadowedBy != "" || v.MovedTo != "" {
		e.logger.Debug("skipping because shadowed or moved")
		return nil
	}

	if e.skip(parent, v) {
		return nil
	}

	params, _ := NewCallableParameters(e, v.CallableAttrs)

	if params == nil {
		return nil
	}

	return &CallableSignature{
		Identifier: &baseIdentifier{
			cIndentifier:   v.CIdentifier,
			cGoIndentifier: "C." + v.CIdentifier,
			goIndentifier:  strcases.SnakeToGo(true, v.Name),
		},
		Parameters: params,
	}
}
