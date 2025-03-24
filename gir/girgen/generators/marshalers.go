package generators

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type MarshalGenerator struct {
	typesystem.Type

	WrapFunction string // may be equal to GoName for Enums and simple conversions
	// ValueFunction is the method on coregllib.Value to use
	ValueFunction string
}

func (g *MarshalGenerator) Generate(w *file.Writer) {
	w.GoImport("unsafe")

	fmt.Fprintf(w.Go(), "func %s(p uintptr) (interface{}, error) {\n", g.MarshalFuncName())
	fmt.Fprintf(w.Go(), "\treturn %s(coreglib.ValueFromNative(unsafe.Pointer(p)).%s()), nil\n", g.WrapFunction, g.ValueFunction)
	fmt.Fprintf(w.Go(), "}\n")
}

func NewMarshalEnumGenerator(typ typesystem.Type) *MarshalGenerator {
	return &MarshalGenerator{
		Type:          typ,
		WrapFunction:  typ.GoType(),
		ValueFunction: "Enum",
	}
}

func NewMarshalBifieldGenerator(typ typesystem.Type) *MarshalGenerator {
	return &MarshalGenerator{
		Type:          typ,
		WrapFunction:  typ.GoType(),
		ValueFunction: "Flags",
	}
}

func NewMarshalObjectGenerator(typ typesystem.Type, wrapCoreObjectName string) *MarshalGenerator {
	return &MarshalGenerator{
		Type:          typ,
		WrapFunction:  wrapCoreObjectName,
		ValueFunction: "Object",
	}
}
