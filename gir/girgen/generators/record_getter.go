package generators

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type RecordFieldGetterGenerator struct {
	Doc    Generator
	GoName string

	Receiver string
}

// Generate implements Generator.
func (r *RecordFieldGetterGenerator) Generate(w *file.Writer) {
	r.Doc.Generate(w)
}

func NewRecordFieldGetterGenerator(ctx gencontext.GenerationContext, parent *RecordGenerator, f gir.Field) *RecordFieldGetterGenerator {
	if !f.IsReadable() || f.Private || f.Type == nil || f.Bits > 0 {
		return nil
	}

	meta := ctx.LookupType(f.Type.CType)

	if meta == nil {
		return nil
	}

	goName := strcases.SnakeToGo(true, f.Name)

	g := &RecordFieldGetterGenerator{
		Doc:      NewGoDocGenerator(f, 0),
		GoName:   goName,
		Receiver: strcases.FirstLetter(goName),
	}

	return g
}
