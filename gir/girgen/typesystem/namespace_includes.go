package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
	"golang.org/x/exp/maps"
)

// namespaceIncludes wraps a gir.Namespace and contains resolved includes
//
// it is used as a preprocessing step before resolving all the types in the namespace
type namespaceIncludes struct {
	versionedName VersionedNamespace
	includes      map[string]*namespaceIncludes

	repository *gir.Repository
	gir.Namespace
}

func resolveNamespaceIncludes(repos gir.Repositories) []*namespaceIncludes {
	namespaces := make(map[VersionedNamespace]*namespaceIncludes)

	for _, repo := range repos {
		includes := repoPrefilledIncludes(repo.Repository)

		for _, ns := range repo.Namespaces {
			versioned := VersionedNamespace{
				Name:         ns.Name,
				MajorVersion: parseMajorVersion(ns.Version),
			}

			namespaces[versioned] = &namespaceIncludes{
				versionedName: versioned,
				includes:      includes,

				repository: &repo.Repository,
				Namespace:  ns,
			}
		}
	}

	// we need to loop again to resolve the includes correctly:
	for _, ns := range namespaces {
		for name, i := range ns.includes {
			included, ok := namespaces[i.versionedName]

			if !ok {
				delete(ns.includes, name)
			} else {
				ns.includes[name] = included
			}
		}
	}

	return maps.Values(namespaces)
}

// repoPrefilledIncludes prefills the includes with namespaces name and version, to
// be able to later
func repoPrefilledIncludes(r gir.Repository) map[string]*namespaceIncludes {
	m := make(map[string]*namespaceIncludes)
	for _, v := range r.Includes {
		m[v.Name] = &namespaceIncludes{
			versionedName: VersionedNamespace{
				Name:         v.Name,
				MajorVersion: parseMajorVersion(v.Version),
			},
		}
	}
	return m
}
