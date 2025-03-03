package generators

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type RecordFieldGetterGenerator struct {
	Doc     Generator
	GoName  string
	CGoName string

	GoType  string
	CgoType string

	GoImports []string

	parent *RecordGenerator
}

// Generate implements Generator.
func (g *RecordFieldGetterGenerator) Generate(w *file.Writer) {
	g.Doc.Generate(w)

	for _, pkg := range g.GoImports {
		w.GoImport(pkg)
	}

	// TODO: simplify, this is only for legacy compat, can be maybe be simplified:

	fmt.Fprintf(w.Go(), "func (%s *%s) %s() %s {\n", g.parent.ReceiverName, g.parent.GoName, g.GoName, g.GoType)
	fmt.Fprintf(w.Go(), "\tvalptr := &%s.native.%s\n", g.parent.ReceiverName, g.CGoName)
	fmt.Fprintf(w.Go(), "\tvar _v %s // out\n", g.GoType)
	fmt.Fprintf(w.Go(), "\t_v = %s(*valptr)\n", g.GoType)
	fmt.Fprintf(w.Go(), "\treturn _v\n")
	fmt.Fprintf(w.Go(), "}\n\n")
}

func NewRecordFieldGetterGenerator(ctx gencontext.GenerationContext, parent *RecordGenerator, f gir.Field) *RecordFieldGetterGenerator {
	if !f.IsReadable() || f.Private || f.Type == nil || f.Bits > 0 {
		return nil
	}

	meta := ctx.LookupType(f.Type.CType)

	if meta == nil {
		return nil
	}

	if meta.CGoPointers > 0 || meta.GoPointers > 0 || !meta.IsCastable {
		return nil
	}

	// TODO: use value converter instead of casting. Remember that the returned value is only borrowed, so it has to keep the record alive
	// if we are wrapping a pointer into the record.

	// TODO: check if this getter collides with any method
	goName := strcases.SnakeToGo(true, f.Name)

	g := &RecordFieldGetterGenerator{
		Doc:     NewGoDocGenerator(goName, f, 0),
		GoName:  goName,
		CGoName: DodgeReservedFieldName(f.Name),
		parent:  parent,

		GoImports: meta.RequiredImports,

		GoType:  meta.GoType(),
		CgoType: meta.CGoType(),
	}

	return g
}
