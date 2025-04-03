package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
)

// Container describes a container type, which must be manually implemented.
type Container struct {
	BaseType

	// The conversions need to accept more parameters, one constructor for each generic subtype
	BaseConversions

	GenericParams int
}

var _ Type = &Container{}

type ContainerInstance struct {
	*Container

	InnerTypes []CouldBeForeign[Type]
}

// GIRName implements Type.
func (c *ContainerInstance) GIRName() string {
	return "instance of" + c.Container.GIRName()
}

// GoType implements Type.
func (c *ContainerInstance) GoType(pointers int) string {
	return c.GoType(pointers)
}

// GoType implements Type.
func (c *ContainerInstance) pointersAllowed(pointers int) bool {
	return pointers == 1
}

var _ Type = &ContainerInstance{}

func (e *env) resolveInnerTypes(outer Type, t *gir.Type) Type {
	if len(t.Types) == 0 {
		return outer
	}

	c, ok := outer.(*Container)
	if !ok {
		e.logger.Warn("skipping type because parent is not generic but has nested types", "outer", outer.GoType(0))
		return nil
	}

	if len(t.Types) != c.GenericParams {
		e.logger.Warn("skipping type it has more nested types than the container can support", "outer", outer.GoType(0))
		return nil
	}

	instance := &ContainerInstance{
		Container: c,
	}

	for _, inner := range t.Types {
		ns, innerTyp := e.findTypeByGIRName(inner.Name)

		if innerTyp == nil {
			return nil
		}

		instance.InnerTypes = append(instance.InnerTypes, CouldBeForeign[Type]{
			Namespace: ns,
			Type:      innerTyp,
		})
	}

	return instance
}

// pointersAllowed implements Type.
func (a *Container) pointersAllowed(pointers int) bool {
	return pointers == 1
}
