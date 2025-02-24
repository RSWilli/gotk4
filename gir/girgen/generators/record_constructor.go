package generators

import (
	"strings"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type RecordConstructorGenerator struct {
	Doc    Generator
	GoName string
}

// Generate implements Generator.
func (r *RecordConstructorGenerator) Generate(w *file.Writer) {

}

func NewRecordConstructorGenerator(ctx gencontext.GenerationContext, parent *RecordGenerator, c gir.Constructor) *RecordConstructorGenerator {
	if !c.IsIntrospectable() {
		return nil
	}

	g := &RecordConstructorGenerator{
		Doc:    NewGoDocGenerator(c, 0),
		GoName: goConstructorName(c.Name, parent.GoName),
	}

	return g
}

// goConstructorName turns the girname into a hopefully unique function name
//
// e.g. BufferList.new_sized -> NewBufferListSized
func goConstructorName(girname string, recordName string) string {
	pascal := strcases.SnakeToGo(true, girname)

	noNew := strings.TrimPrefix(pascal, "New")

	return "New" + recordName + noNew
}
