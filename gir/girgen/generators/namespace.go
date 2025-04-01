package generators

import (
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type NamespaceGenerator struct {
	ns *typesystem.Namespace

	// Sub generators:
	Constants  GeneratorList
	Aliases    GeneratorList
	Enums      GeneratorList
	Bitfields  GeneratorList
	Callbacks  GeneratorList
	Functions  GeneratorList
	Interfaces GeneratorList
	Classes    GeneratorList
	Records    GeneratorList
	Unions     GeneratorList
}

// Generate implements Generator.
func (g *NamespaceGenerator) Generate(w *file.Package) {
	w.SetNamespace(g.ns)

	// run sub generators
	GenerateAll(
		w,
		g.Constants,
		g.Aliases,
		g.Enums,
		g.Bitfields,
		g.Callbacks,
		g.Functions,
		g.Interfaces,
		g.Classes,
		g.Records,
		g.Unions,
	)
}

func NewNamespaceGenerator(
	ns *typesystem.Namespace,
) *NamespaceGenerator {
	// namespace := ctx.Namespace(ns)
	gen := &NamespaceGenerator{
		ns: ns,
	}

	for _, c := range ns.Constants {
		if cgen := NewConstantGenerator(c); cgen != nil {
			gen.Constants = append(gen.Constants, cgen)
		}
	}

	for _, a := range ns.Aliases {
		if agen := NewAliasGenerator(a); agen != nil {
			gen.Aliases = append(gen.Aliases, agen)
		}
	}
	for _, e := range ns.Enums {
		if egen := NewEnumGenerator(e); egen != nil {
			gen.Enums = append(gen.Enums, egen)
		}
	}
	for _, b := range ns.Bitfields {
		if bgen := NewBitfieldGenerator(b); bgen != nil {
			gen.Bitfields = append(gen.Bitfields, bgen)
		}
	}
	for _, cb := range ns.Callbacks {
		if cbgen := NewCallbackGenerator(cb); cbgen != nil {
			gen.Callbacks = append(gen.Callbacks, cbgen)
		}
	}
	for _, f := range ns.Functions {
		if fgen := NewCallableGenerator(f); fgen != nil {
			gen.Functions = append(gen.Functions, fgen)
		}
	}
	for _, inter := range ns.Interfaces {
		if intergen := NewInterfaceGenerator(inter); intergen != nil {
			gen.Interfaces = append(gen.Interfaces, intergen)
		}
	}
	for _, class := range ns.Classes {
		if classgen := NewClassGenerator(class); classgen != nil {
			gen.Classes = append(gen.Classes, classgen)
		}
	}
	for _, r := range ns.Records {
		if rgen := NewRecordGenerator(r); rgen != nil {
			gen.Records = append(gen.Records, rgen)
		}
	}
	// for _, v := range ns.Unions {
	// }

	return gen
}
