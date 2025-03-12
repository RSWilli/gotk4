package typesystem

import (
	"fmt"
	"log"

	"github.com/diamondburned/gotk4/gir"
)

// namespaceWithIncludes wraps a gir.Namespace and contains resolved includes
//
// it is used as a preprocessing step before resolving all the types in the namespace
type namespaceWithIncludes struct {
	versionedName versionedNamespace
	includes      map[string]*namespaceWithIncludes

	repository *gir.Repository
	gir.Namespace
}

type versionedNamespace struct {
	name    string
	version gir.Version
}

func (v versionedNamespace) String() string {
	return fmt.Sprintf("%s-%s", v.name, v.version)
}

func resolveNamespaceIncludes(repos gir.Repositories) []*namespaceWithIncludes {
	namespacesByName := make(map[versionedNamespace]*namespaceWithIncludes)
	namespaces := make([]*namespaceWithIncludes, 0, len(repos))

	for _, repo := range repos {
		includes := repoPrefilledIncludes(repo.Repository)

		for _, ns := range repo.Namespaces {
			versioned := versionedNamespace{
				name:    ns.Name,
				version: ns.Version,
			}

			namespace := &namespaceWithIncludes{
				versionedName: versioned,
				includes:      includes,

				repository: &repo.Repository,
				Namespace:  ns,
			}

			namespacesByName[versioned] = namespace
			namespaces = append(namespaces, namespace)
		}
	}

	// we need to loop again to resolve the includes correctly:
	for _, ns := range namespacesByName {
		for name, i := range ns.includes {
			included, ok := namespacesByName[i.versionedName]

			if !ok {
				log.Printf("%s included namespace %s which wasn't found, ignoring for now\n", ns.versionedName, i.versionedName)
				delete(ns.includes, name)
			} else {
				ns.includes[name] = included
			}
		}
	}

	// and one more time for transitive includes:
	for _, ns := range namespaces {
		for _, incl := range ns.includes {
			for name, transIncl := range incl.includes {
				ns.includes[name] = transIncl
			}
		}
	}

	return namespaces
}

// repoPrefilledIncludes prefills the includes with namespaces name and version, to
// be able to later
func repoPrefilledIncludes(r gir.Repository) map[string]*namespaceWithIncludes {
	m := make(map[string]*namespaceWithIncludes)
	for _, v := range r.Includes {
		m[v.Name] = &namespaceWithIncludes{
			versionedName: versionedNamespace{
				name:    v.Name,
				version: v.Version,
			},
		}
	}
	return m
}
