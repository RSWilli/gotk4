package generators

import "github.com/diamondburned/gotk4/gir/girgen/file"

type Generator interface {
	Generate(*file.Writer)
}

type GeneratorList []Generator

var _ Generator = GeneratorList{}

func (list GeneratorList) Generate(w *file.Writer) {
	for _, g := range list {
		if g == nil {
			continue
		}

		g.Generate(w)
	}
}

type NoopGenerator struct{}

func (list NoopGenerator) Generate(w *file.Writer) {}

func GenerateAll(w *file.Writer, gens ...Generator) {
	GeneratorList(gens).Generate(w)
}
