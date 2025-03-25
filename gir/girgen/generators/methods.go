package generators

import "github.com/diamondburned/gotk4/gir/girgen/file"

type MethodGeneratorList []MethodGenerator

type MethodGenerator interface {
	Generator

	GenerateInterfaceSignature(w file.CodeWriter)
}

var _ Generator = (MethodGeneratorList)(nil)

// Generate implements Generator.
func (list MethodGeneratorList) Generate(w *file.Writer) {
	for _, g := range list {
		if g == nil {
			continue
		}

		g.Generate(w)
	}
}

// GenerateInterfaceSignature implements MethodGenerator.
func (list MethodGeneratorList) GenerateInterfaceSignatures(w file.CodeWriter) {
	for _, g := range list {
		if g == nil {
			continue
		}

		g.GenerateInterfaceSignature(w)
	}
}
