package generators

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type RecordFieldGetterGenerator struct {
	Doc SubGenerator

	ReceiverName string

	*typesystem.Field
}

// Generate implements Generator.
func (g *RecordFieldGetterGenerator) Generate(w *file.Package) {
	g.Doc.Generate(w.Go())

	// TODO: simplify, this is only for legacy compat, can be maybe be simplified:

	fmt.Fprintf(w.Go(), "func (%s *%s) %s() %s {\n", g.ReceiverName, g.Parent.GoType(0), g.GoGetterName, g.Type.NamespacedGoType(0))
	fmt.Fprintf(w.Go(), "\tvalptr := &%s.native.%s\n", g.ReceiverName, g.CGoIndentifier())
	fmt.Fprintf(w.Go(), "\tvar _v %s\n", g.Type.NamespacedGoType(0))
	fmt.Fprintf(w.Go(), "\t_v = %s(*valptr)\n", g.Type.NamespacedGoType(0))
	fmt.Fprintf(w.Go(), "\treturn _v\n")
	fmt.Fprintf(w.Go(), "}\n\n")
}

func NewRecordFieldGetterGenerator(f *typesystem.Field) *RecordFieldGetterGenerator {
	if f.GoGetterName == "" {
		return nil
	}

	if _, ok := f.Type.Type.(*typesystem.CastablePrimitive); !ok || f.CTypePointers > 0 {
		return nil
	}

	if f.Type.Type == typesystem.Utf8 {
		return nil // TODO
	}

	g := &RecordFieldGetterGenerator{
		Doc: NewIdentifierGoDocGenerator(f),

		ReceiverName: strcases.ReceiverName(f.Parent.GoType(0)),

		Field: f,
	}

	return g
}
