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
func (g *RecordFieldGetterGenerator) Generate(w *file.Writer) {
	g.Doc.Generate(w.Go())

	// TODO: simplify, this is only for legacy compat, can be maybe be simplified:

	fmt.Fprintf(w.Go(), "func (%s *%s) %s() %s {\n", g.ReceiverName, g.Parent.GoType(), g.GoGetterName, g.Type.GoType())
	fmt.Fprintf(w.Go(), "\tvalptr := &%s.native.%s\n", g.ReceiverName, g.CGoIndentifier())
	fmt.Fprintf(w.Go(), "\tvar _v %s // out\n", g.Type.GoType())
	fmt.Fprintf(w.Go(), "\t_v = %s(*valptr)\n", g.Type.GoType())
	fmt.Fprintf(w.Go(), "\treturn _v\n")
	fmt.Fprintf(w.Go(), "}\n\n")
}

func NewRecordFieldGetterGenerator(f *typesystem.Field) *RecordFieldGetterGenerator {
	if f.GoGetterName == "" {
		return nil
	}

	if !typesystem.IsPrimitive(f.Type) || typesystem.Pointers(f.Type) > 0 {
		// TODO: this can be done, but remember that the returned value is only borrowed, so it has to keep the record alive
		// if we are wrapping a pointer into the record.
		return nil
	}

	if typesystem.Is(f.Type, typesystem.Utf8) {
		return nil // TODO
	}

	g := &RecordFieldGetterGenerator{
		Doc: NewIdentifierGoDocGenerator(f),

		ReceiverName: strcases.ReceiverName(f.Parent.GoType()),

		Field: f,
	}

	return g
}
