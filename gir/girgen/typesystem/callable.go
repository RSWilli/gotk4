package typesystem

import (
	"strings"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

// CallableIdentifier is an identifier that prefixes the parent type, so that renaming the
// parent struct reflects to renaming the constructors / methods. It has a special case for
// constructors, which are prefixed with "New" and the parent type name.
type CallableIdentifier struct {
	Parent         Type
	Girname        string
	GirCIdentifier string
}

// CGoIndentifier implements Identifier.
func (c *CallableIdentifier) CGoIndentifier() string {
	return "C." + c.GirCIdentifier
}

// CIndentifier implements Identifier.
func (c *CallableIdentifier) CIndentifier() string {
	return c.GirCIdentifier
}

// GoIndentifier turns the girname into a hopefully unique function name
//
// e.g. BufferList.new_sized -> NewBufferListSized
func (c *CallableIdentifier) GoIndentifier() string {
	pascal := strcases.SnakeToGo(true, c.Girname)

	pascal, hasNewPrefix := strings.CutPrefix(pascal, "New")
	pascal, hasNewSuffix := strings.CutSuffix(pascal, "New")

	var parentTypeName string

	if c.Parent != nil {
		parentTypeName = c.Parent.GoType(0)

		switch p := c.Parent.(type) {
		case *Class:
			parentTypeName = p.GoInterfaceName
		case *Interface:
			parentTypeName = p.GoInterfaceName
		}
	}

	if hasNewPrefix || hasNewSuffix {
		return "New" + parentTypeName + pascal
	}

	return parentTypeName + pascal
}

var _ Identifier = &CallableIdentifier{}

type CallableSignature struct {
	Identifier
	*Parameters
}

func DeclareFunction(e *env, v *gir.CallableAttrs) *CallableSignature {
	e = e.sub("function", v.CIdentifier)

	if !v.IsIntrospectable() {
		e.logger.Warn("skipping because not introspectable")
		return nil
	}

	if v.ShadowedBy != "" {
		e.logger.Warn("skipping because shadowed", "by", v.ShadowedBy)
		return nil
	}

	if v.MovedTo != "" {
		e.logger.Warn("skipping because moved", "to", v.MovedTo)
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
		Identifier: &CallableIdentifier{
			Parent:         nil,
			Girname:        v.Name,
			GirCIdentifier: v.CIdentifier,
		},
		Parameters: params,
	}
}

func DeclarePrefixedFunction(e *env, parent Type, v *gir.CallableAttrs) *CallableSignature {
	e = e.sub("function", v.CIdentifier)

	if !v.IsIntrospectable() {
		e.logger.Warn("skipping because not introspectable")
		return nil
	}

	if v.ShadowedBy != "" {
		e.logger.Warn("skipping because shadowed", "by", v.ShadowedBy)
		return nil
	}

	if v.MovedTo != "" {
		e.logger.Warn("skipping because moved", "to", v.MovedTo)
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
		Identifier: &CallableIdentifier{
			Parent:         parent,
			Girname:        v.Name,
			GirCIdentifier: v.CIdentifier,
		},
		Parameters: params,
	}
}

func DeclareMethod(e *env, parent Type, v *gir.Method) *CallableSignature {
	e = e.sub("method", v.CIdentifier)

	if !v.IsIntrospectable() {
		e.logger.Warn("skipping because not introspectable")
		return nil
	}

	if v.ShadowedBy != "" {
		e.logger.Warn("skipping because shadowed", "by", v.ShadowedBy)
		return nil
	}

	if v.MovedTo != "" {
		e.logger.Warn("skipping because moved", "to", v.MovedTo)
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
		Identifier: &CallableIdentifier{
			// Methods are scoped to the parent, so we don't need a parent prefix
			Parent:         nil,
			Girname:        v.Name,
			GirCIdentifier: v.CIdentifier,
		},
		Parameters: params,
	}
}
