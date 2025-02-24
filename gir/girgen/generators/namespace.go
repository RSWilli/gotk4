package generators

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
	"github.com/diamondburned/gotk4/gir/girgen/file"
)

type NamespaceGenerator struct {
	ctx       gencontext.GenerationContext
	ns        *gir.Namespace
	cIncludes []string
	packages  []string

	GoPackageName  string
	GoPackageMajor int

	// Sub generators:
	GTypes     Generator
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
func (g *NamespaceGenerator) Generate(w *file.Writer) {
	w.SetGoPackageName(g.GoPackageName, g.GoPackageMajor)

	w.AddCFlag("-Wno-deprecated-declarations")

	for _, cIncl := range g.cIncludes {
		w.CInclude(cIncl)
		w.Exported.CInclude(cIncl)
	}

	for _, pkg := range g.packages {
		w.AddPackage(pkg)
	}

	// run sub generators
	GenerateAll(
		w,
		g.GTypes,
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
	ctx gencontext.GenerationContext,
	ns *gir.Namespace,
	cIncludes []gir.CInclude,
	packages []gir.Package,
) *NamespaceGenerator {
	meta := ctx.NamespaceMetadata(ns)
	gen := &NamespaceGenerator{
		ctx: ctx,
		ns:  ns,

		GoPackageName:  meta.GoPackageName,
		GoPackageMajor: meta.MajorVersion,

		GTypes: NewRegisterGTypeGenerator(ctx, ns),
	}

	// use a namespaced context for sub types lookup etc
	ctx = gencontext.Namespaced(meta, ctx)

	for _, cIncl := range cIncludes {
		gen.cIncludes = append(gen.cIncludes, cIncl.Name)
	}

	for _, pkg := range packages {
		gen.packages = append(gen.packages, pkg.Name)
	}

	for _, c := range ns.Constants {
		if cgen := NewConstantGenerator(ctx, c); cgen != nil {
			gen.Constants = append(gen.Constants, cgen)
		}
	}

	for _, a := range ns.Aliases {
		if agen := NewAliasGenerator(ctx, a); agen != nil {
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
		if cbgen := NewCallbackGenerator(ctx, cb); cbgen != nil {
			gen.Callbacks = append(gen.Callbacks, cbgen)
		}
	}
	// for _, v := range ns.Functions {
	// }
	// for _, v := range ns.Interfaces {
	// }
	// for _, v := range ns.Classes {
	// }
	for _, r := range ns.Records {
		if rgen := NewRecordGenerator(ctx, r); rgen != nil {
			gen.Records = append(gen.Records, rgen)
		}
	}
	// for _, v := range ns.Unions {
	// }

	return gen
}
