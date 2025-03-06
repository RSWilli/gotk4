package typesystem

import (
	"log"
	"strings"

	"github.com/diamondburned/gotk4/gir"
)

type Registry struct {
	namespaces map[VersionedNamespace]*Namespace
}

type VersionedNamespace struct {
	Name         string
	MajorVersion int
}

func versionedNsFromString(s string) VersionedNamespace {
	parts := strings.Split(s, "-")

	if len(parts) > 2 {
		panic("got invalid versioned namespace string " + s)
	}

	return VersionedNamespace{
		Name:         parts[0],
		MajorVersion: parseMajorVersion(parts[1]),
	}
}

// FromRepositories loads all repositories into the registry and checks whether all includes are present
func FromRepositories(repos gir.Repositories, modulePath string, importOverrides map[string]string) *Registry {
	r := &Registry{
		namespaces: make(map[VersionedNamespace]*Namespace, len(repos)),
	}

	for _, repo := range repos {
		includes := repoPrefilledIncludes(repo.Repository)

		for _, ns := range repo.Namespaces {
			versioned := VersionedNamespace{
				Name:         ns.Name,
				MajorVersion: parseMajorVersion(ns.Version),
			}

			r.namespaces[versioned] = &Namespace{
				VersionedName: versioned,

				GoPackageName: goPackageName(ns.Name),
				goImportPath:  goImportPath(modulePath, ns.Name, ns.Version),

				includes: includes,

				aliasesByName:    nsAliases(ns),
				classesByName:    nsClasses(ns),
				interfacesByName: nsInterfaces(ns),
				recordsByName:    nsRecords(ns),
				enumsByName:      nsEnums(ns),
				functionsByName:  nsFunctions(ns),
				unionsByName:     nsUnions(ns),
				bitfieldsByName:  nsBitfiels(ns),
				callbacksByName:  nsCallbacks(ns),
				constantsByName:  nsConstant(ns),
			}
		}
	}

	// we need to loop again to resolve the includes correctly:

	for _, ns := range r.namespaces {
		for name, i := range ns.includes {
			included, ok := r.namespaces[i.VersionedName]

			if !ok {
				log.Printf("typesystem: could not find included namespace %s-%d, ignoring for now\n", i.VersionedName.Name, i.VersionedName.MajorVersion)
			}

			ns.includes[name] = included
		}
	}

	for namespace, importOverride := range importOverrides {
		ns, ok := r.namespaces[versionedNsFromString(namespace)]

		if !ok {
			continue
		}

		ns.goImportPath = importOverride
	}

	return r
}

type Namespace struct {
	VersionedName VersionedNamespace

	GoPackageName string
	goImportPath  string

	includes map[string]*Namespace

	aliasesByName    map[string]gir.Alias
	classesByName    map[string]gir.Class
	interfacesByName map[string]gir.Interface
	recordsByName    map[string]gir.Record
	enumsByName      map[string]gir.Enum
	functionsByName  map[string]gir.Function
	unionsByName     map[string]gir.Union
	bitfieldsByName  map[string]gir.Bitfield
	callbacksByName  map[string]gir.Callback
	constantsByName  map[string]gir.Constant
}

func nsAliases(ns gir.Namespace) map[string]gir.Alias {
	m := make(map[string]gir.Alias)
	for _, v := range ns.Aliases {
		m[v.Name] = v
	}
	return m
}

func nsClasses(ns gir.Namespace) map[string]gir.Class {
	m := make(map[string]gir.Class)
	for _, v := range ns.Classes {
		m[v.Name] = v
	}
	return m
}

func nsInterfaces(ns gir.Namespace) map[string]gir.Interface {
	m := make(map[string]gir.Interface)
	for _, v := range ns.Interfaces {
		m[v.Name] = v
	}
	return m
}

func nsRecords(ns gir.Namespace) map[string]gir.Record {
	m := make(map[string]gir.Record)
	for _, v := range ns.Records {
		m[v.Name] = v
	}
	return m
}

func nsEnums(ns gir.Namespace) map[string]gir.Enum {
	m := make(map[string]gir.Enum)
	for _, v := range ns.Enums {
		m[v.Name] = v
	}
	return m
}

func nsFunctions(ns gir.Namespace) map[string]gir.Function {
	m := make(map[string]gir.Function)
	for _, v := range ns.Functions {
		m[v.Name] = v
	}
	return m
}

func nsUnions(ns gir.Namespace) map[string]gir.Union {
	m := make(map[string]gir.Union)
	for _, v := range ns.Unions {
		m[v.Name] = v
	}
	return m
}

func nsBitfiels(ns gir.Namespace) map[string]gir.Bitfield {
	m := make(map[string]gir.Bitfield)
	for _, v := range ns.Bitfields {
		m[v.Name] = v
	}
	return m
}

func nsCallbacks(ns gir.Namespace) map[string]gir.Callback {
	m := make(map[string]gir.Callback)
	for _, v := range ns.Callbacks {
		m[v.Name] = v
	}
	return m
}

func nsConstant(ns gir.Namespace) map[string]gir.Constant {
	m := make(map[string]gir.Constant)
	for _, v := range ns.Constants {
		m[v.CType] = v
	}
	return m
}
