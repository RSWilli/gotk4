package generators

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
)

type GoDocGenerator struct {
}

func (n *GoDocGenerator) Generate(w *file.Writer) {
	// TODO:

	fmt.Fprintf(w.Go(), "// TODO godoc\n")
}

func NewGoDocGenerator(girWithDoc any, indent int) *GoDocGenerator {
	return &GoDocGenerator{}
}
