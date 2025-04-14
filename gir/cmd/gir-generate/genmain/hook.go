package genmain

import "github.com/diamondburned/gotk4/gir/girgen/generators"

type Generators struct {
	Namespaces []*generators.NamespaceGenerator
}

type GeneratorHook func(*Generators)

func AddGeneratorToPackage(goname string, newgen generators.Generator) GeneratorHook {
	return func(gens *Generators) {
		for _, generator := range gens.Namespaces {
			if generator.Namespace.GoName == goname {
				generator.SubGenerators = append(generator.SubGenerators, newgen)
				return
			}
		}
	}
}
