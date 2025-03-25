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
func (s *RecordFieldSetterGenerator) Generate(w *file.Writer) {
	s.Doc.Generate(w.Go())

	// TODO: simplify, this is only for legacy compat, can be maybe be simplified:

	fmt.Fprintf(w.Go(), "func (%s *%s) %s(%s %s) {\n", s.ReceiverName, s.Parent.GoType(), s.GoSetterName, s.CGoIndentifier(), s.Type.GoType())
	fmt.Fprintf(w.Go(), "\tvalptr := &%s.native.%s\n", s.ReceiverName, s.CGoIndentifier())
	fmt.Fprintf(w.Go(), "\t*valptr = %s(%s)\n", s.Type.CGoType(), s.CGoIndentifier())
	fmt.Fprintf(w.Go(), "}\n\n")
}

func NewRecordFieldSetterGenerator(f *typesystem.Field) *RecordFieldSetterGenerator {
	if f.GoSetterName == "" {
		return nil
	}

	if !typesystem.IsPrimitive(f.Type) || typesystem.Pointers(f.Type) > 0 {
		// TODO: this can be done, but who owns the field if it is writable?
		return nil
	}

	if typesystem.Is(f.Type, typesystem.Utf8) {
		return nil // TODO
	}

	g := &RecordFieldSetterGenerator{
		Doc:          NewIdentifierGoDocGenerator(f),
		ReceiverName: strcases.ReceiverName(f.Parent.GoType()),
		Field:        f,
	}

	return g
}
