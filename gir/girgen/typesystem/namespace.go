package typesystem

import (
	"log"

	"github.com/diamondburned/gotk4/gir"
)

type Namespace struct {
	v versionedName

	Included map[string]*Namespace

	Name    string
	GoName  string
	Version gir.Version

	Packages  []string
	CIncludes []string

	// User overwritten types for resolving in other namespaces:
	Manual []Type

	// immediately available types:
	Bitfields []*Bitfield
	Enums     []*Enum

	// lazily resolved types, which need a declare and a resolve step
	Aliases    []*Alias
	Callbacks  []*Callback
	Classes    []*Class
	Interfaces []*Interface
	Records    []*Record
	Unions     []*Union

	// identifiers, these are eagerly resolved
	Constants []*Constant
	Functions []*CallableSignature
}

func (reg *Registry) newNamespace(cfg Config, ns *namespaceWithIncludes) *Namespace {

	namespace := &Namespace{
		v:        ns.versionedName,
		Name:     ns.Name,
		GoName:   goPackageName(ns.Name),
		Version:  ns.versionedName.version,
		Included: make(map[string]*Namespace, len(ns.includes)),
	}

	for ident, incl := range ns.includes {
		reffedNS := reg.findNS(incl.versionedName)

		if reffedNS == nil {
			log.Printf("could not find referenced namespace %s for %s\n", incl.versionedName, ns.versionedName)
			continue
		}

		namespace.Included[ident] = reffedNS
	}

	for _, cincl := range ns.repository.CIncludes {
		namespace.CIncludes = append(namespace.CIncludes, cincl.Name)
	}
	for _, pkg := range ns.repository.Packages {
		namespace.Packages = append(namespace.Packages, pkg.Name)
	}

	e := cfg.getNamespaceEnv(namespace)

	if e == nil {
		log.Printf("ignoring ignored namespace %s", ns.versionedName)
		return nil
	}

	for _, t := range e.nsCfg.ManualTypes {
		namespace.Manual = append(namespace.Manual, t)
	}

	// these types are directly valid and will only omit child declarations afterwards:
	for _, v := range ns.Unions {
		if t := DeclareUnion(e, v); t != nil {
			namespace.Unions = append(namespace.Unions, t)
		}
	}
	for _, v := range ns.Enums {
		if t := DeclareEnum(e, v); t != nil {
			namespace.Enums = append(namespace.Enums, t)
		}
	}
	for _, v := range ns.Bitfields {
		if t := DeclareBitfield(e, v); t != nil {
			namespace.Bitfields = append(namespace.Bitfields, t)
		}
	}
	for _, v := range ns.Records {
		if t := DeclareRecord(e, v); t != nil {
			namespace.Records = append(namespace.Records, t)
		}
	}

	// these types may have references to other types that are not yet known to be valid:
	var unresolvedClasses []*Class
	var unresolvedInterfaces []*Interface
	var unresolvedCallbacks []*Callback
	var unresolvedAliases []*Alias

	for _, v := range ns.Callbacks {
		if t := DeclareCallback(e, v); t != nil {
			unresolvedCallbacks = append(unresolvedCallbacks, t)
		}
	}
	for _, v := range ns.Interfaces {
		if t := DeclareInterface(e, v); t != nil {
			unresolvedInterfaces = append(unresolvedInterfaces, t)
		}
	}
	for _, v := range ns.Classes {
		if t := DeclareClass(e, v); t != nil {
			unresolvedClasses = append(unresolvedClasses, t)
		}
	}
	for _, v := range ns.Aliases {
		if t := DeclareAlias(e, v); t != nil {
			unresolvedAliases = append(unresolvedAliases, t)
		}
	}

	namespace.resolveAll(
		e,
		unresolvedClasses,
		unresolvedInterfaces,
		unresolvedCallbacks,
		unresolvedAliases,
	)

	// declare these after resolving all types, because they reference the above:
	for _, v := range namespace.Unions {
		v.declareNested(e)
	}
	for _, v := range namespace.Records {
		v.declareNested(e)
	}
	for _, v := range namespace.Classes {
		v.declareNested(e)
	}
	for _, v := range namespace.Interfaces {
		v.declareNested(e)
	}
	for _, v := range ns.Functions {
		if t := DeclareFunction(e, nil, v); t != nil {
			namespace.Functions = append(namespace.Functions, t)
		}
	}
	for _, v := range ns.Constants {
		if t := DeclareConstant(e, v); t != nil {
			namespace.Constants = append(namespace.Constants, t)
		}
	}

	return namespace
}

// resolveAll tries to resolve the given unresolved types until either:
//
// * all types are resolved
//
// * a single iteration does not shrink the list of unresolved types
func (n *Namespace) resolveAll(e *env, unresolvedClasses []*Class, unresolvedInterfaces []*Interface, unresolvedCallbacks []*Callback, unresolvedAliases []*Alias) {
	for {
		stillUnresolvedClasses := unresolvedClasses[0:0]
		stillUnresolvedInterfaces := unresolvedInterfaces[0:0]
		stillUnresolvedCallbacks := unresolvedCallbacks[0:0]
		stillUnresolvedAliases := unresolvedAliases[0:0]

		for _, v := range unresolvedClasses {
			if v.resolve(e) {
				n.Classes = append(n.Classes, v)
			} else {
				stillUnresolvedClasses = append(stillUnresolvedClasses, v)
			}
		}

		for _, v := range unresolvedInterfaces {
			if v.resolve(e) {
				n.Interfaces = append(n.Interfaces, v)
			} else {
				stillUnresolvedInterfaces = append(stillUnresolvedInterfaces, v)
			}
		}

		for _, v := range unresolvedCallbacks {
			if v.resolveParameters(e) {
				n.Callbacks = append(n.Callbacks, v)
			} else {
				stillUnresolvedCallbacks = append(stillUnresolvedCallbacks, v)
			}
		}

		for _, v := range unresolvedAliases {
			if v.resolve(e) {
				n.Aliases = append(n.Aliases, v)
			} else {
				stillUnresolvedAliases = append(stillUnresolvedAliases, v)
			}
		}

		if len(stillUnresolvedClasses) == 0 &&
			len(stillUnresolvedInterfaces) == 0 &&
			len(stillUnresolvedCallbacks) == 0 &&
			len(stillUnresolvedAliases) == 0 {
			// we are done
			log.Printf("successfully resolved all classes, interfaces, callbacks and aliases in %s", n.v)
			return
		}

		if len(stillUnresolvedClasses) == len(unresolvedClasses) &&
			len(stillUnresolvedInterfaces) == len(unresolvedInterfaces) &&
			len(stillUnresolvedCallbacks) == len(unresolvedCallbacks) &&
			len(stillUnresolvedAliases) == len(unresolvedAliases) {
			// we did not make progress, so we drop the unresolvable types
			log.Printf("could not resolve %d classes, %d interfaces, %d callbacks and %d aliases in %s", len(unresolvedClasses), len(unresolvedInterfaces), len(unresolvedCallbacks), len(unresolvedAliases), n.v)

			return
		}

		unresolvedClasses = stillUnresolvedClasses
		unresolvedInterfaces = stillUnresolvedInterfaces
		unresolvedCallbacks = stillUnresolvedCallbacks
		unresolvedAliases = stillUnresolvedAliases
	}
}

// findLocalTypeByGIRName returns the [Type] for the named type
func (n *Namespace) findLocalTypeByGIRName(girname string) Type {
	return n.findLocalTypeWith(func(t Type) bool {
		return t.GIRName() == girname
	})
}

// findLocalTypeWith returns the [Type] where the predicate returns true
func (n *Namespace) findLocalTypeWith(pred func(t Type) bool) Type {
	for _, m := range n.Manual {
		if pred(m) {
			return m
		}
	}

	for _, a := range n.Aliases {
		if pred(a) {
			return a
		}
	}

	for _, c := range n.Classes {
		if pred(c) {
			return c
		}
	}

	for _, c := range n.Interfaces {
		if pred(c) {
			return c
		}
	}

	for _, c := range n.Records {
		if pred(c) {
			return c
		}
	}
	for _, c := range n.Callbacks {
		if pred(c) {
			return c
		}
	}

	for _, c := range n.Enums {
		if pred(c) {
			return c
		}
	}

	for _, c := range n.Unions {
		if pred(c) {
			return c
		}
	}

	for _, c := range n.Bitfields {
		if pred(c) {
			return c
		}
	}

	return nil
}
