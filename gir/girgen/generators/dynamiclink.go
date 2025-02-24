package generators

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
)

// WithDynamicLinking builds all generators needed for generated cgo dynamic linked code.
func WithDynamicLinking(ctx gencontext.GenerationContext, repos gir.Repositories) (gens []Generator) {
	for _, repo := range repos {
		for _, ns := range repo.Namespaces {
			gens = append(gens, NewNamespaceGenerator(
				ctx,
				&ns,
				repo.CIncludes,
				repo.Packages,
			))
		}
	}

	return
}
