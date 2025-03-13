package typesystem

import "github.com/diamondburned/gotk4/gir"

type Container struct {
	Outer Type
	Inner []Type
}

// GLibGetType implements Type.
func (c *Container) GLibGetType() string {
	return ""
}

// MarshalFuncName implements Type.
func (c *Container) MarshalFuncName() string {
	return ""
}

// CGoType implements Type.
func (c *Container) CGoType() string {
	return c.Outer.CGoType()
}

// CType implements Type.
func (c *Container) CType() string {
	return c.Outer.CType()
}

// GIRName implements Type.
func (c *Container) GIRName() string {
	return c.Outer.GIRName()
}

// GoType implements Type.
func (c *Container) GoType() string {
	return c.Outer.GoType()
}

var _ Type = &Container{}

func resolveInnerTypes(ns *Namespace, outer Type, t *gir.Type) Type {
	if len(t.Types) == 0 {
		return outer
	}

	c := &Container{
		Outer: outer,
	}

	for _, inner := range t.Types {
		innerTyp := ns.findTypeByGIRName(inner.Name)

		if innerTyp == nil {
			return nil
		}

		c.Inner = append(c.Inner, innerTyp)
	}

	return c
}
