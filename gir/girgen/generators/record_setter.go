package generators

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type RecordFieldSetterGenerator struct {
	Doc SubGenerator

	ReceiverName string

	*typesystem.Field
}

// Generate implements Generator.
func (s *RecordFieldSetterGenerator) Generate(w *file.Package) {
	s.Doc.Generate(w.Go())

	// TODO: simplify, this is only for legacy compat, can be maybe be simplified:

	fmt.Fprintf(w.Go(), "func (%s *%s) %s(%s %s) {\n", s.ReceiverName, s.Parent.GoType(0), s.GoSetterName, s.CGoIndentifier(), s.Type.Type.GoType(0))
	fmt.Fprintf(w.Go(), "\tvalptr := &%s.native.%s\n", s.ReceiverName, s.CGoIndentifier())
	fmt.Fprintf(w.Go(), "\t*valptr = %s(%s)\n", s.Type.Type.CGoType(0), s.CGoIndentifier())
	fmt.Fprintf(w.Go(), "}\n\n")
}

func NewRecordFieldSetterGenerator(f *typesystem.Field) *RecordFieldSetterGenerator {
	if f.GoSetterName == "" {
		return nil
	}

	if _, ok := f.Type.Type.(*typesystem.CastablePrimitive); !ok || f.CTypePointers > 0 {
		return nil
	}

	if f.Type.Type == typesystem.Gpointer {
		return nil // not supported
	}

	if f.Type.Type == typesystem.Utf8 {
		return nil // TODO
	}

	g := &RecordFieldSetterGenerator{
		Doc:          NewIdentifierGoDocGenerator(f),
		ReceiverName: strcases.ReceiverName(f.Parent.GoType(0)),
		Field:        f,
	}

	return g
}
