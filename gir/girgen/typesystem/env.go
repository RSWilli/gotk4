package typesystem

import (
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/diamondburned/gotk4/gir"
)

type ParamCompareFunc func(a, b *Param) int

type env struct {
	cfg       Config
	nsCfg     NamespaceConfig
	namespace *Namespace

	minVersion gir.Version

	// user overridable settings via [Config]:

	ignore IgnoreFunc

	logger *slog.Logger
}

// sub returns a sub env that is an exact copy but with the given attrs used in the logger
func (e *env) sub(attrs ...any) *env {
	subenv := *e

	logger := subenv.logger.With(attrs...)
	subenv.logger = logger

	return &subenv
}

// sortGoParams moves the params according to go conventions:
// 1. context.Context is always first
// 2. error is always last
// 3. all other params don't get reordered
func (e *env) sortGoParams(ps []*Param) {
	slices.SortFunc(ps, func(a, b *Param) int {
		if a.GoType() == "context.Context" {
			return -1
		}
		if b.GoType() == "context.Context" {
			return 1
		}
		if a.GoType() == "error" {
			return 1
		}
		if b.GoType() == "error" {
			return -1
		}
		return 0
	})
}

// sortGoReturns moves the return values according to go conventions:
// 1. error is always last
// 2. all other return values don't get reordered
func (e *env) sortGoReturns(ps []*Param) {
	slices.SortFunc(ps, func(a, b *Param) int {
		if a.GoType() == "error" {
			return 1
		}
		if b.GoType() == "error" {
			return -1
		}
		return 0
	})
}

func (e *env) trampolinePrefix() string {
	return fmt.Sprintf("_gotk4_%s%d", e.namespace.GoName, e.namespace.Version.Major)
}

// skip returns true if the gir type/identifier should be skipped. Any optional parent can be passed
// to handle nested gir identifiers
func (e *env) skip(parent Type, anygir any) bool {
	name, attrs, elements := infoFromAnyGir(anygir)

	if e.ingoreDeprecated(name, attrs) {
		return true
	}

	for _, m := range e.namespace.Manual {
		if m.GIRName() == name {
			e.logger.Info("skipping manually implemented type", "name", name)
			return true
		}
	}

	var parentName string
	if parent != nil {
		parentName = parent.GIRName()
	}

	return e.ignore(parentName, name, attrs, elements)
}

type girWithInfoAttrs interface {
	GetInfoAttrs() gir.InfoAttrs
}

type girWithInfoElements interface {
	GetInfoElements() gir.InfoElements
}

func (e *env) ingoreDeprecated(name string, attrs gir.InfoAttrs) bool {
	if attrs.Deprecated && attrs.DeprecatedVersion.Less(e.minVersion) {
		e.logger.Info("skipping deprecated", "name", name, "deprecated-since", attrs.DeprecatedVersion, "min-version", e.minVersion)
		return true
	}

	return false
}

func (e *env) findAnyType(t gir.AnyType) (*Namespace, Type) {
	if t.Type != nil && t.Array != nil {
		panic("received invalid anytype")
	}

	if t.Type != nil {
		ns, typ := e.findType(t.Type)

		if typ == nil {
			return nil, nil
		}
		return ns, typ
	}

	if t.Array != nil {
		arr := e.getArrayType(t.Array)

		if arr == nil {
			return nil, nil
		}
		return nil, arr
	}

	// this happens e.g. on vararg params
	return nil, nil
}

// findType searches for a declared type in the namespace
func (e *env) findType(t *gir.Type) (*Namespace, Type) {
	var ns *Namespace
	var typ Type
	if t.CType != "" {
		// Ctype resolving is often more reliable than GIR name resolving, so we try it first

		ctype := cleanCType(t.CType)

		ns, typ = e.findTypeByCType(ctype)
	}

	if typ == nil {
		// try GIR name resolving as fallback
		ns, typ = e.findTypeByGIRName(t.Name)
	}

	if typ == nil {
		return nil, nil
	}

	typ = e.resolveInnerTypes(typ, t)

	if typ == nil {
		return nil, nil
	}

	return ns, typ
}

func (e *env) findTypeByCType(t string) (*Namespace, Type) {
	if ctypeIsIncompatible(t) {
		return nil, nil
	}

	primitive := findBuiltinPrimitiveByCType(t)

	if primitive != nil {
		return nil, primitive
	}

	typ := e.namespace.findLocalTypeByCType(t)

	if typ != nil {
		return nil, typ
	}

	// the referenced type is not in the current namespace, so we need to find it
	// in the included namespaces. We need to try all included namespaces, because
	// we have no idea which one is the right one. FIXME: we can know this by looking at the ctype prefix

	for _, reffedNS := range e.namespace.Included {
		typ := reffedNS.findLocalTypeByCType(t)

		if typ != nil {
			return reffedNS, typ
		}
	}

	e.logger.Debug("type not found", "type", t)
	return nil, nil
}

func (e *env) findTypeByGIRName(t string) (*Namespace, Type) {
	if replaced, ok := e.cfg.GIRReplacements[t]; ok {
		e.logger.Warn("replacing GIR type name", "type", t, "replaced by", replaced)
		t = replaced
	}

	if ctypeIsIncompatible(t) {
		return nil, nil
	}

	parts := strings.Split(t, ".")

	if len(parts) > 2 {
		panic("received invalid type name")
	}

	if len(parts) == 1 {
		primitive := findBuiltinPrimitiveByGIRName(t)

		if primitive != nil {
			return nil, primitive
		}

		typ := e.namespace.findLocalTypeByGIRName(t)

		if typ == nil {
			e.logger.Debug("type not found", "type", t)
			return nil, nil
		}

		return nil, typ
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
		e.logger.Warn("type referenced unknown namespace", "type", t, "referenced-ns", foreignNSName)
		return nil, nil
	}

	foreign := reffedNS.findLocalTypeByGIRName(foreignTypeName)

	if foreign != nil {
		return reffedNS, foreign
	}

	e.logger.Debug("type not found", "type", t)

	return nil, nil
}
