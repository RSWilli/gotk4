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

// GoKeywords includes Go keywords. This is primarily to prevent collisions with
// meaningful Go words.
var GoKeywords = map[string]string{
	// Keywords.
	"break":       "",
	"default":     "",
	"func":        "fn",
	"interface":   "iface",
	"select":      "sel",
	"case":        "",
	"defer":       "",
	"go":          "",
	"map":         "",
	"struct":      "",
	"chan":        "ch",
	"else":        "",
	"goto":        "",
	"package":     "pkg",
	"switch":      "",
	"const":       "",
	"fallthrough": "",
	"if":          "",
	"range":       "",
	"type":        "typ",
	"continue":    "",
	"for":         "",
	"import":      "",
	"return":      "ret",
	"var":         "",
}
