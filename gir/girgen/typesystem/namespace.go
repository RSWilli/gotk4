package typesystem

import (
	"log"
	"strings"

	"github.com/diamondburned/gotk4/gir"
)

type Namespace struct {
	v versionedNamespace

	Included map[string]*Namespace

	GoName       string
	MajorVersion int

	Packages  []string
	CIncludes []string

	Constants  []*Constant
	Aliases    []*Alias
	Bitfields  []*Bitfield
	Enums      []*Enum
	Callbacks  []*Callback
	Classes    []*Class
	Interfaces []*Interface
	Records    []*Record
	Unions     []*Union

	Functions []*Callable
}

func newNamespace(reg *Registry, ns *namespaceWithIncludes) *Namespace {
	namespace := &Namespace{
		v:            ns.versionedName,
		GoName:       goPackageName(ns.Name),
		MajorVersion: ns.versionedName.majorVersion,
		Included:     make(map[string]*Namespace, len(ns.includes)),
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

	// types must be declared first. Some types contain nested references to other types.
	// these are deferred and resolved at the end of the namespace creation.

	for _, v := range ns.Unions {
		if t := NewUnion(namespace, v); t != nil {
			namespace.Unions = append(namespace.Unions, t)

			defer t.resolveNested(namespace, v)
		}
	}
	for _, v := range ns.Enums {
		if t := NewEnum(namespace, v); t != nil {
			namespace.Enums = append(namespace.Enums, t)
		}
	}
	for _, v := range ns.Bitfields {
		if t := NewBitfield(namespace, v); t != nil {
			namespace.Bitfields = append(namespace.Bitfields, t)
		}
	}
	for _, v := range ns.Callbacks {
		if t := NewCallback(namespace, v); t != nil {
			namespace.Callbacks = append(namespace.Callbacks, t)

			defer t.resolveParameters(namespace, v)
		}
	}
	for _, v := range ns.Records {
		if t := NewRecord(namespace, v); t != nil {
			namespace.Records = append(namespace.Records, t)

			defer t.resolveNested(namespace, v)
		}
	}
	for _, v := range ns.Interfaces {
		if t := NewInterface(namespace, v); t != nil {
			namespace.Interfaces = append(namespace.Interfaces, t)

			defer t.resolveNested(namespace, v)
		}
	}
	for _, v := range ns.Classes {
		if t := NewClass(namespace, v); t != nil {
			namespace.Classes = append(namespace.Classes, t)

			defer t.resolveNested(namespace, v)
		}
	}
	for _, v := range ns.Aliases {
		if t := NewAlias(namespace, v); t != nil {
			namespace.Aliases = append(namespace.Aliases, t)
		}
	}

	// create these after initializing all types:

	for _, v := range ns.Functions {
		if t := NewFunction(namespace, v); t != nil {
			namespace.Functions = append(namespace.Functions, t)
		}
	}
	for _, v := range ns.Constants {
		if t := NewConstant(namespace, v); t != nil {
			namespace.Constants = append(namespace.Constants, t)
		}
	}

	return namespace
}

func (ns *Namespace) findAnyType(t gir.AnyType) Type {
	if t.Type != nil && t.Array != nil {
		panic("received invalid anytype")
	}

	if t.Type != nil {
		typ := ns.findType(t.Type)

		if typ == nil {
			return nil
		}
		return typ
	}

	if t.Array != nil {
		arr := getArrayType(ns, t.Array)

		if arr == nil {
			return nil
		}
		return arr
	}

	panic("received invalid anytype")
}

func (ns *Namespace) findType(t *gir.Type) Type {
	if isIgnoredType(t) {
		return nil
	}

	parts := strings.Split(t.Name, ".")

	if len(parts) > 2 {
		panic("received invalid type name")
	}

	if len(parts) == 1 {
		primitive := findPrimitiveType(t)

		if primitive != nil {
			return primitive
		}

		typ := ns.findLocalType(t)

		if typ == nil {
			log.Printf("type %s not found in namespace %s\n", t.Name, ns.v)
			return nil
		}

		return typ
	}

	// remove the namespace identifier in front:
	foreignType := *t
	foreignType.Name = parts[1]

	if parts[0] == ns.v.name {
		// some glib types are referenced with glib prefix, e.g. HashTable
		return ns.findType(&foreignType)
	}

	reffedNS, ok := ns.Included[parts[0]]

	if !ok {
		log.Printf("type %s referenced unknown namespace %s\n", t.Name, parts[0])
		return nil
	}

	foreign := reffedNS.findLocalType(&foreignType)

	if foreign != nil {
		return &ForeignType{
			SourceNamespace: ns,
			Type:            foreign,
		}
	}

	log.Printf("type %s not found in namespace %s\n", t.Name, ns.v)

	return nil
}

// findLocalType returns the [Type] for the named type
func (n *Namespace) findLocalType(gir *gir.Type) Type {
	return n.findTypeWith(func(t Type) bool {
		return t.GIRName() == gir.Name
	})
}

// findCType returns the [Type] for the ctype. ctype must be pointerless
func (n *Namespace) findCType(ctype string) Type {
	ctype = strings.TrimPrefix("const ", ctype)

	// TODO: find primitive by ctype

	return n.findTypeWith(func(t Type) bool {
		return t.CType() == ctype
	})
}

// findTypeWith returns the [Type] where the predicate returns true
func (n *Namespace) findTypeWith(pred func(t Type) bool) Type {
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
