package typesystem

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/diamondburned/gotk4/gir"
)

type ParamCompareFunc func(a, b *Param) int

type env struct {
	namespace *Namespace

	// user overridable settings via [Config]:

	skipTypeFunc func(t girType) bool

	compareParams  ParamCompareFunc
	compareReturns ParamCompareFunc

	ignoredNamespaces []string
}

func (c *env) sortGoParams(ps []*Param) {
	if c.compareParams == nil {
		return
	}

	slices.SortFunc(ps, c.compareParams)
}

func (c *env) sortGoReturns(ps []*Param) {
	if c.compareReturns == nil {
		return
	}

	slices.SortFunc(ps, c.compareReturns)
}

func (c env) trampolinePrefix() string {
	return fmt.Sprintf("_gotk4_%s%d", c.namespace.GoName, c.namespace.Version.Major)
}

func (c env) skipType(t girType) bool {
	return c.skipTypeFunc(t)
}

func (e *env) findAnyType(t gir.AnyType) Type {
	if t.Type != nil && t.Array != nil {
		panic("received invalid anytype")
	}

	if t.Type != nil {
		typ := e.findType(t.Type)

		if typ == nil {
			return nil
		}
		return typ
	}

	if t.Array != nil {
		arr := e.getArrayType(t.Array)

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
func (e *env) findType(t *gir.Type) Type {
	typ := e.findTypeByGIRName(t.Name)

	if typ == nil {
		return nil
	}

	typ = e.resolveInnerTypes(typ, t)

	if typ == nil {
		return nil
	}

	return WithPointers(t, typ)
}

func (e *env) findTypeByGIRName(t string) Type {
	if isIncompatible(t) {
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

		typ := e.namespace.findLocalTypeByGIRName(t)

		if typ == nil {
			log.Printf("type %s not found in namespace %s\n", t, e.namespace.v)
			return nil
		}

		return typ
	}

	foreignNSName := parts[0]
	foreignTypeName := parts[1]

	if e.isIgnoredNamespace(foreignNSName) {
		return nil
	}

	if foreignNSName == e.namespace.v.name {
		// some glib types are always referenced with glib prefix, e.g. HashTable
		// even in glib namespace.
		return e.findTypeByGIRName(foreignTypeName)
	}

	reffedNS, ok := e.namespace.Included[foreignNSName]

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

	log.Printf("type %s not found in namespace %s\n", t, e.namespace.v)

	return nil
}

func (c *env) isIgnoredNamespace(girnamespace string) bool {
	return slices.Contains(c.ignoredNamespaces, girnamespace)
}
