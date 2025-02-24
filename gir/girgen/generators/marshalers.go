package generators

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
)

type MarshalGenerator struct {
	GoName       string
	WrapFunction string // may be equal to GoName for Enums and simple conversions
	// ValueFunction is the method on coregllib.Value to use
	ValueFunction string
}

func (g *MarshalGenerator) Generate(w *file.Writer) {
	w.GoImport("unsafe")

	fmt.Fprintf(w.Go(), "func marshal%s(p uintptr) (interface{}, error) {\n", g.GoName)
	fmt.Fprintf(w.Go(), "\treturn %s(coreglib.ValueFromNative(unsafe.Pointer(p)).%s()), nil\n", g.WrapFunction, g.ValueFunction)
	fmt.Fprintf(w.Go(), "}\n\n")
}

func NewMarshalEnumGenerator(goName string) *MarshalGenerator {
	return &MarshalGenerator{
		GoName:        goName,
		WrapFunction:  goName,
		ValueFunction: "Enum",
	}
}

func NewMarshalBifieldGenerator(goName string) *MarshalGenerator {
	return &MarshalGenerator{
		GoName:        goName,
		WrapFunction:  goName,
		ValueFunction: "Flags",
	}
}
