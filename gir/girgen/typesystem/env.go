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

	minVersion gir.Version

	// user overridable settings via [Config]:

	ignore IgnoreFunc

	compareParams  ParamCompareFunc
	compareReturns ParamCompareFunc
}

func (e *env) sortGoParams(ps []*Param) {
	if e.compareParams == nil {
		return
	}

	slices.SortFunc(ps, e.compareParams)
}

func (e *env) sortGoReturns(ps []*Param) {
	if e.compareReturns == nil {
		return
	}

	slices.SortFunc(ps, e.compareReturns)
}

func (e *env) trampolinePrefix() string {
	return fmt.Sprintf("_gotk4_%s%d", e.namespace.GoName, e.namespace.Version.Major)
}

// skip returns true if the gir type/identifier should be skipped. Any optional parent can be passed
// to handle nested gir identifiers
func (e *env) skip(parent Type, anygir any) bool {
	name, kind := infoFromAnyGir(anygir)

	if e.ingoreDeprecated(name, kind, anygir) {
		return true
	}

	for _, m := range e.namespace.Manual {
		if m.GIRName() == name {
			log.Printf("skipping %s %s because it is manually implemented", kind, name)
			return true
		}
	}

	var parentName string
	if parent != nil {
		parentName = parent.GIRName()
	}

	id := GIRIdentifier{
		Parent: parentName,
		Name:   name,
		Kind:   kind,
	}

	return e.ignore(id)
}

type girWithInfoAttrs interface {
	GetInfoAttrs() gir.InfoAttrs
}

func (e *env) ingoreDeprecated(name string, kind GIRKind, anygir any) bool {
	gt, ok := anygir.(girWithInfoAttrs)

	if !ok {
		return false
	}

	attrs := gt.GetInfoAttrs()

	if attrs.Deprecated && attrs.DeprecatedVersion.Less(e.minVersion) {
		log.Printf("skipping %s %s in %s that is deprecated since %s, min allowed version: %s", kind, name, e.namespace.v, attrs.DeprecatedVersion, e.minVersion)
		return true
	}

	return false
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
		return mkForeign(reffedNS, foreign)
	}

	log.Printf("type %s not found in namespace %s\n", t, e.namespace.v)

	return nil
}
