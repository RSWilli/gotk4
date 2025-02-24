package generators

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type RecordMethodGenerator struct {
	Doc    Generator
	GoName string
}

// Generate implements Generator.
func (r *RecordMethodGenerator) Generate(w *file.Writer) {

}

func NewRecordMethodGenerator(ctx gencontext.GenerationContext, parent *RecordGenerator, m gir.Method) *RecordMethodGenerator {
	if !m.IsIntrospectable() {
		return nil
	}

	g := &RecordMethodGenerator{
		Doc:    NewGoDocGenerator(m, 0),
		GoName: strcases.SnakeToGo(true, m.Name),
	}

	return g
}
