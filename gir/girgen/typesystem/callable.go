package typesystem

import (
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type CallableSignature struct {
	Identifier
	*Parameters
}

func DeclareFunction(e *env, parent Type, v gir.Function) *CallableSignature {
	if !v.IsIntrospectable() {
		return nil
	}

	if v.ShadowedBy != "" || v.MovedTo != "" {
		// log.Printf("skipping shadowed or moved function %s", v.CIdentifier)
		return nil
	}

	if e.skip(parent, v) {
		return nil
	}

	params := NewCallableParameters(e, v.CallableAttrs)

	if params == nil {
		// log.Printf("could not create parameters for function %s\n", v.CIdentifier)
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

func NewMethod(e *env, parent Type, v gir.Method) *CallableSignature {
	if !v.IsIntrospectable() {
		return nil
	}

	if v.ShadowedBy != "" || v.MovedTo != "" {
		// log.Printf("skipping shadowed or moved function %s", v.CIdentifier)
		return nil
	}

	if e.skip(parent, v) {
		return nil
	}

	params := NewCallableParameters(e, v.CallableAttrs)

	if params == nil {
		// log.Printf("could not create parameters for method %s\n", v.CIdentifier)
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

// constructorIdentifier is an identifier that references the parent type, so that renaming the
// parent struct reflects to renaming the constructor methods
type constructorIdentifier struct {
	parent         Type
	girname        string
	girCIdentifier string
}

// CGoIndentifier implements Identifier.
func (c *constructorIdentifier) CGoIndentifier() string {
	return "C." + c.girCIdentifier
}

// CIndentifier implements Identifier.
func (c *constructorIdentifier) CIndentifier() string {
	return c.girCIdentifier
}

// GoIndentifier turns the girname into a hopefully unique function name
//
// e.g. BufferList.new_sized -> NewBufferListSized
func (c *constructorIdentifier) GoIndentifier() string {
	pascal := strcases.SnakeToGo(true, c.girname)

	noNew, ok := strings.CutPrefix(pascal, "New")

	if ok {
		return "New" + c.parent.GoType() + noNew
	}

	return c.parent.GoType() + pascal
}

var _ Identifier = &constructorIdentifier{}

func DeclareConstructor(e *env, parent Type, v gir.Constructor) *CallableSignature {
	if !v.IsIntrospectable() {
		return nil
	}

	if v.ShadowedBy != "" || v.MovedTo != "" {
		// log.Printf("skipping shadowed or moved function %s", v.CIdentifier)
		return nil
	}

	if e.skip(parent, v) {
		return nil
	}

	params := NewCallableParameters(e, v.CallableAttrs)

	if params == nil {
		// log.Printf("could not create parameters for constructor %s\n", v.CIdentifier)
		return nil
	}

	return &CallableSignature{
		Identifier: &constructorIdentifier{
			parent:         parent,
			girname:        v.Name,
			girCIdentifier: v.CIdentifier,
		},
		Parameters: params,
	}
}

// GoSignature returns a string of the go function signature.
func (m *CallableSignature) GoSignature() string {
	var recv string
	if m.InstanceParam != nil {
		recv = fmt.Sprintf(" (%s %s)", m.InstanceParam.GoName, m.InstanceParam.Type.GoType())
	}

	var ret string
	if len(m.GoReturns) == 1 {
		ret = " " + m.GoReturns[0].Type.GoType()
	} else if len(m.GoReturns) > 1 {
		ret = "(" + m.GoReturns.GoTypes() + ")"
	}

	return fmt.Sprintf("func%s %s(%s)%s", recv, m.GoIndentifier(), m.GoParameters.GoDeclarations(), ret)
}

// GoSignature returns a string of the go function signature.
func (m *CallableSignature) GoInterfaceDeclaration() string {
	var ret string
	if len(m.GoReturns) == 1 {
		ret = " " + m.GoReturns[0].Type.GoType()
	} else if len(m.GoReturns) > 1 {
		ret = "(" + m.GoReturns.GoTypes() + ")"
	}

	return fmt.Sprintf("%s(%s)%s", m.GoIndentifier(), m.GoParameters.GoTypes(), ret)
}
