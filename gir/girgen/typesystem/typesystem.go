package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
)

type Registry struct {
	Namespaces []*Namespace
}

// FromRepositories loads all repositories into the registry and resolves all type references.
//
// the repositories must be loaded in the correct order to be able to resolve all types
func FromRepositories(cfg Config, repos gir.Repositories) *Registry {
	r := &Registry{
		Namespaces: make([]*Namespace, 0, len(repos)),
	}

	withIncludes := resolveNamespaceIncludes(repos)

	skipFuncs := cfg.getSkipFuncs(withIncludes)

	for _, nsTmp := range withIncludes {

		ns := r.newNamespace(skipFuncs[nsTmp.versionedName], nsTmp)

		if ns == nil {
			continue
		}

		r.Namespaces = append(r.Namespaces, ns)
	}

	return r
}

func (r *Registry) findNS(v versionedNamespace) *Namespace {
	for _, ns := range r.Namespaces {
		if ns.v == v {
			return ns
		}
	}

	return nil
}
