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
	versionedName versionedName
	includes      map[string]*namespaceWithIncludes

	repository *gir.Repository
	gir.Namespace
}

type repoWithIncludes struct {
	gir.Repository

	namespaces []*namespaceWithIncludes
}

type versionedName struct {
	name    string
	version gir.Version
}

func (v versionedName) String() string {
	return fmt.Sprintf("%s-%s", v.name, v.version)
}

func resolveNamespaceIncludes(repos gir.Repositories) []*repoWithIncludes {
	namespacesByName := make(map[versionedName]*namespaceWithIncludes)
	outRepos := make([]*repoWithIncludes, 0, len(repos))

	for _, repo := range repos {
		includes := repoPrefilledIncludes(repo.Repository)

		outRepo := &repoWithIncludes{
			Repository: repo.Repository,
		}

		for _, ns := range repo.Namespaces {
			versioned := versionedName{
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
			outRepo.namespaces = append(outRepo.namespaces, namespace)
		}

		outRepos = append(outRepos, outRepo)
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
	//
	// we are using the repos in the given order, which means we hopefully get all transitive imports
	// in a single loop, given all depended repos come before
	for _, repo := range outRepos {
		for _, ns := range repo.namespaces {
			for _, incl := range ns.includes {
				for name, transIncl := range incl.includes {
					ns.includes[name] = transIncl
				}
			}
		}
	}

	return outRepos
}

// repoPrefilledIncludes prefills the includes with namespaces name and version, to
// be able to later set it to the actual pointer to the foreign namespace
func repoPrefilledIncludes(r gir.Repository) map[string]*namespaceWithIncludes {
	m := make(map[string]*namespaceWithIncludes)
	for _, v := range r.Includes {
		m[v.Name] = &namespaceWithIncludes{
			versionedName: versionedName{
				name:    v.Name,
				version: v.Version,
			},
		}
	}
	return m
}
