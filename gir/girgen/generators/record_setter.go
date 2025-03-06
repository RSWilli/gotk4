package generators

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type RecordFieldSetterGenerator struct {
	Doc     Generator
	GoName  string
	CGoName string

	GoType  string
	CgoType string

	GoImports []string

	parent *RecordGenerator
}

// Generate implements Generator.
func (s *RecordFieldSetterGenerator) Generate(w *file.Writer) {
	s.Doc.Generate(w)

	for _, pkg := range s.GoImports {
		w.GoImport(pkg)
	}

	// TODO: simplify, this is only for legacy compat, can be maybe be simplified:

	fmt.Fprintf(w.Go(), "func (%s *%s) %s(%s %s) {\n", s.parent.ReceiverName, s.parent.GoName, s.GoName, s.CGoName, s.GoType)
	fmt.Fprintf(w.Go(), "\tvalptr := &%s.native.%s\n", s.parent.ReceiverName, s.CGoName)
	fmt.Fprintf(w.Go(), "\t*valptr = %s(%s)\n", s.CgoType, s.CGoName)
	fmt.Fprintf(w.Go(), "}\n\n")
}

func NewRecordFieldSetterGenerator(ctx gencontext.GenerationContext, parent *RecordGenerator, f gir.Field) *RecordFieldSetterGenerator {
	if !f.Writable || !f.IsReadable() || f.Private || f.Type == nil || f.Bits > 0 {
		return nil
	}

	meta := ctx.LookupType(f.Type.Name, f.Type.CType)

	if meta == nil {
		return nil
	}

	if meta.CGoPointers > 0 {
		// cannot allocate
		return nil
	}

	if meta.CGoBaseType == "guintptr" || !meta.IsCastable {
		return nil
	}

	// TODO: use value converter instead of casting

	// TODO: check if this getter collides with any method
	goName := strcases.SnakeToGo(true, "set_"+f.Name)

	g := &RecordFieldSetterGenerator{
		Doc:       NewGoDocGenerator(goName, f, 0),
		GoName:    goName,
		CGoName:   DodgeReservedFieldName(f.Name),
		parent:    parent,
		GoImports: meta.RequiredImports,
		CgoType:   meta.CGoType(),
		GoType:    meta.GoType(),
	}

	return g
}

// go
func DodgeReservedFieldName(s string) string {
	if s, ok := GoKeywords[s]; ok {
		return "_" + s
	}

	return s
}
