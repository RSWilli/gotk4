package generators

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type MarshalGenerator struct {
	Type typesystem.Marshalable

	WrapFunction string // may be equal to GoName for Enums and simple conversions
	// ValueFunction is the method on coregllib.Value to use
	ValueFunction string
}

func (g *MarshalGenerator) Generate(w *file.Package) {
	w.GoImport("unsafe")

	// TODO: import gobject

	fmt.Fprintf(w.Go(), "func marshal%s(p uintptr) (interface{}, error) {\n", g.Type.GoType(0))
	fmt.Fprintf(w.Go(), "\treturn %s(gobject.ValueFromNative(unsafe.Pointer(p)).%s()), nil\n", g.WrapFunction, g.ValueFunction)
	fmt.Fprintf(w.Go(), "}\n")
}

func NewMarshalEnumGenerator(typ typesystem.Marshalable) *MarshalGenerator {
	return &MarshalGenerator{
		Type:          typ,
		WrapFunction:  typ.GoType(0),
		ValueFunction: "Enum",
	}
}

func NewMarshalBifieldGenerator(typ typesystem.Marshalable) *MarshalGenerator {
	return &MarshalGenerator{
		Type:          typ,
		WrapFunction:  typ.GoType(0),
		ValueFunction: "Flags",
	}
}

func NewMarshalObjectGenerator(typ typesystem.Marshalable, wrapCoreObjectName string) *MarshalGenerator {
	return &MarshalGenerator{
		Type:          typ,
		WrapFunction:  wrapCoreObjectName,
		ValueFunction: "Object",
	}
}
