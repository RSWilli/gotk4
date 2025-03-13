package typesystem

import (
	"log"
	"strings"

	"github.com/diamondburned/gotk4/gir"
)

type Namespace struct {
	v versionedNamespace

	Included map[string]*Namespace

	Name    string
	GoName  string
	Version gir.Version

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

	Functions []*CallableSignature
}

func (reg *Registry) newNamespace(sf skipFunc, ns *namespaceWithIncludes) *Namespace {
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

	ctx := &namespaceContext{
		skipTypeFunc: sf,
		Namespace:    namespace,
	}

	// types must be declared first. Some types contain nested references to other types.
	// these are deferred and resolved at the end of the namespace creation.

	for _, v := range ns.Unions {
		if t := DeclareUnion(ctx, v); t != nil {
			namespace.Unions = append(namespace.Unions, t)

			defer t.resolveNested(ctx, v)
		}
	}
	for _, v := range ns.Enums {
		if t := DeclareEnum(ctx, v); t != nil {
			namespace.Enums = append(namespace.Enums, t)
		}
	}
	for _, v := range ns.Bitfields {
		if t := DeclareBitfield(ctx, v); t != nil {
			namespace.Bitfields = append(namespace.Bitfields, t)
		}
	}
	for _, v := range ns.Callbacks {
		if t := DeclareCallback(ctx, v); t != nil {
			namespace.Callbacks = append(namespace.Callbacks, t)

			defer t.resolveParameters(ctx, v)
		}
	}
	for _, v := range ns.Interfaces {
		if t := DeclareInterface(ctx, v); t != nil {
			namespace.Interfaces = append(namespace.Interfaces, t)

			defer t.resolveNested(ctx, v)
		}
	}
	for _, v := range ns.Classes {
		if t := NewClass(ctx, v); t != nil {
			namespace.Classes = append(namespace.Classes, t)

			defer t.resolveNested(ctx, v)
		}
	}
	for _, v := range ns.Records {
		if t := DeclareRecord(ctx, v); t != nil {
			namespace.Records = append(namespace.Records, t)

			// needs to be after classes/interfaces, because this defer needs to run before the classes/interfaces:
			defer t.resolveNested(ctx, v)
		}
	}
	for _, v := range ns.Aliases {
		if t := DeclareAlias(ctx, v); t != nil {
			namespace.Aliases = append(namespace.Aliases, t)
		}
	}

	// declare these after declaring all types, because they reference the above:

	for _, v := range ns.Functions {
		if t := DeclareFunction(ctx, v); t != nil {
			namespace.Functions = append(namespace.Functions, t)
		}
	}
	for _, v := range ns.Constants {
		if t := DeclareConstant(ctx, v); t != nil {
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

	// this happens e.g. on vararg params
	return nil
}

// findType searches for a declared type in the namespace. It makes sure that the returned type
// contains the same amount of pointers as the given gir type
func (ns *Namespace) findType(t *gir.Type) Type {
	typ := ns.findTypeByGIRName(t.Name)

	if typ == nil {
		return nil
	}

	typ = resolveInnerTypes(ns, typ, t)

	if typ == nil {
		return nil
	}

	return WithPointers(t, typ)
}

func (ns *Namespace) findTypeByGIRName(t string) Type {
	if isIgnoredTypeName(t) {
		return nil
	}

	parts := strings.Split(t, ".")

	if len(parts) > 2 {
		panic("received invalid type name")
	}

	if len(parts) == 1 {
		primitive := findPrimitiveByName(t)

		if primitive != nil {
			return primitive
		}

		typ := ns.findLocalTypeByGIRName(t)

		if typ == nil {
			log.Printf("type %s not found in namespace %s\n", t, ns.v)
			return nil
		}

		return typ
	}

	foreignNSName := parts[0]
	foreignTypeName := parts[1]

	if isIgnoredNSName(foreignNSName) {
		return nil
	}

	if foreignNSName == ns.v.name {
		// some glib types are always referenced with glib prefix, e.g. HashTable
		// even in glib namespace.
		return ns.findTypeByGIRName(foreignTypeName)
	}

	reffedNS, ok := ns.Included[foreignNSName]

	if !ok {
		log.Printf("type %s referenced unknown namespace %s\n", t, foreignNSName)
		return nil
	}

	foreign := reffedNS.findLocalTypeByGIRName(foreignTypeName)

	if foreign != nil {
		return &ForeignType{
			SourceNamespace: reffedNS,
			Type:            foreign,
		}
	}

	log.Printf("type %s not found in namespace %s\n", t, ns.v)

	return nil
}

// findLocalTypeByGIRName returns the [Type] for the named type
func (n *Namespace) findLocalTypeByGIRName(girname string) Type {
	return n.findLocalTypeWith(func(t Type) bool {
		return t.GIRName() == girname
	})
}

// findLocalTypeWith returns the [Type] where the predicate returns true
func (n *Namespace) findLocalTypeWith(pred func(t Type) bool) Type {
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
