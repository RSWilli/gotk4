package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
)

type Registry struct {
	Repositories []*Repository
}

// FromRepositories loads all repositories into the registry and resolves all type references.
//
// the repositories must be loaded in the correct order to be able to resolve all types
func FromRepositories(cfg Config, repos gir.Repositories) *Registry {
	r := &Registry{
		Repositories: make([]*Repository, 0, len(repos)),
	}

	withIncludes := resolveNamespaceIncludes(repos)

	for _, repoTmp := range withIncludes {
		repo := &Repository{
			Pkg: repoTmp.Pkg,
		}

		for _, nsTmp := range repoTmp.namespaces {

			ns := r.newNamespace(cfg, nsTmp)

			if ns == nil {
				continue
			}

			repo.Namespaces = append(repo.Namespaces, ns)
		}

		r.Repositories = append(r.Repositories, repo)
	}

	return r
}

func (r *Registry) findNS(v versionedName) *Namespace {
	for _, repo := range r.Repositories {
		for _, ns := range repo.Namespaces {
			if ns.v == v {
				return ns
			}
		}
	}

	return nil
}
