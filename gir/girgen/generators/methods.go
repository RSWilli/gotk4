package generators

import "github.com/diamondburned/gotk4/gir/girgen/file"

type MethodGeneratorList []MethodGenerator

type MethodGenerator interface {
	Generator

	GenerateInterfaceSignature(w file.File)
}

var _ Generator = (MethodGeneratorList)(nil)

// Generate implements Generator.
func (list MethodGeneratorList) Generate(w *file.Package) {
	for _, g := range list {
		if g == nil {
			continue
		}

		g.Generate(w)
	}
}

// GenerateInterfaceSignature implements MethodGenerator.
func (list MethodGeneratorList) GenerateInterfaceSignatures(w file.File) {
	for _, g := range list {
		if g == nil {
			continue
		}

		g.GenerateInterfaceSignature(w)
	}
}
