package generators

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type RecordFieldSetterGenerator struct {
	Doc    Generator
	GoName string
}

// Generate implements Generator.
func (r *RecordFieldSetterGenerator) Generate(w *file.Writer) {

}

func NewRecordFieldSetterGenerator(ctx gencontext.GenerationContext, parent *RecordGenerator, f gir.Field) *RecordFieldSetterGenerator {
	if !f.Writable || f.Private || f.Type == nil || f.Bits > 0 {
		return nil
	}

	meta := ctx.LookupType(f.Type.CType)

	if meta == nil {
		return nil
	}

	if meta.CGoPointers > 0 {
		// cannot allocate
		return nil
	}

	if meta.CGoBaseType == "guintptr" || meta.GoBaseType == "string" {
		return nil
	}

	g := &RecordFieldSetterGenerator{
		Doc:    NewGoDocGenerator(f, 0),
		GoName: strcases.SnakeToGo(true, f.Name),
	}

	return g
}
