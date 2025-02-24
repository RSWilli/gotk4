package typesystem

import (
	"log"
	"strings"

	"github.com/diamondburned/gotk4/gir"
)

type Registry struct {
	namespaces map[versionedNamespace]*namespace
}

type versionedNamespace struct {
	name         string
	majorversion string
}

func versionedNsFromString(s string) versionedNamespace {
	parts := strings.Split(s, "-")

	if len(parts) > 2 {
		panic("got invalid versioned namespace string " + s)
	}

	return versionedNamespace{
		name:         parts[0],
		majorversion: gir.MajorVersion(parts[1]),
	}
}

// FromRepositories loads all repositories into the registry and checks whether all includes are present
func FromRepositories(repos gir.Repositories, modulePath string, importOverrides map[string]string) *Registry {
	r := &Registry{
		namespaces: make(map[versionedNamespace]*namespace, len(repos)),
	}

	for _, repo := range repos {
		includes := repoPrefilledIncludes(repo.Repository)

		for _, ns := range repo.Namespaces {
			r.namespaces[versionedNamespace{
				name:         ns.Name,
				majorversion: gir.MajorVersion(ns.Version),
			}] = &namespace{
				name:    ns.Name,
				version: ns.Version,

				goPackageName: goPackageName(ns.Name),
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
			included, ok := r.namespaces[versionedNamespace{name: i.name, majorversion: gir.MajorVersion(i.version)}]

			if !ok {
				log.Printf("typesystem: could not find included namespace %s-%s, ignoring for now\n", i.name, i.version)
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

type namespace struct {
	name    string
	version string

	goPackageName string
	goImportPath  string

	includes map[string]*namespace

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

// repoPrefilledIncludes prefills the includes with namespaces name and version, to
// be able to later
func repoPrefilledIncludes(r gir.Repository) map[string]*namespace {
	m := make(map[string]*namespace)
	for _, v := range r.Includes {
		m[v.Name] = &namespace{
			name:    v.Name,
			version: v.Version,
		}
	}
	return m
}

func nsAliases(ns gir.Namespace) map[string]gir.Alias {
	m := make(map[string]gir.Alias)
	for _, v := range ns.Aliases {
		m[v.CType] = v
	}
	return m
}

func nsClasses(ns gir.Namespace) map[string]gir.Class {
	m := make(map[string]gir.Class)
	for _, v := range ns.Classes {
		m[v.CType] = v
	}
	return m
}

func nsInterfaces(ns gir.Namespace) map[string]gir.Interface {
	m := make(map[string]gir.Interface)
	for _, v := range ns.Interfaces {
		m[v.CType] = v
	}
	return m
}

func nsRecords(ns gir.Namespace) map[string]gir.Record {
	m := make(map[string]gir.Record)
	for _, v := range ns.Records {
		m[v.CType] = v
	}
	return m
}

func nsEnums(ns gir.Namespace) map[string]gir.Enum {
	m := make(map[string]gir.Enum)
	for _, v := range ns.Enums {
		m[v.CType] = v
	}
	return m
}

func nsFunctions(ns gir.Namespace) map[string]gir.Function {
	m := make(map[string]gir.Function)
	for _, v := range ns.Functions {
		m[v.CIdentifier] = v
	}
	return m
}

func nsUnions(ns gir.Namespace) map[string]gir.Union {
	m := make(map[string]gir.Union)
	for _, v := range ns.Unions {
		m[v.CType] = v
	}
	return m
}

func nsBitfiels(ns gir.Namespace) map[string]gir.Bitfield {
	m := make(map[string]gir.Bitfield)
	for _, v := range ns.Bitfields {
		m[v.CType] = v
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
