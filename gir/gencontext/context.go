package gencontext

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

// GenerationContext contains the information that is useful for every generator in the tree. It gets passed
// through the generator constructors.
type GenerationContext interface {
	Namespace(ns *gir.Namespace) *typesystem.Namespace
	LookupType(typename, cType string) *typesystem.TypeMetadata
}

type baseGenerationContext struct {
	Typesystem *typesystem.Registry
}

func Base(typesystem *typesystem.Registry) GenerationContext {
	return &baseGenerationContext{
		Typesystem: typesystem,
	}
}

func (ctx *baseGenerationContext) Namespace(ns *gir.Namespace) *typesystem.Namespace {
	panic("")
}

func (ctx *baseGenerationContext) LookupType(typename, cType string) *typesystem.TypeMetadata {
	panic("")
}

func Namespaced(base GenerationContext, namespace *typesystem.Namespace) GenerationContext {
	return &namespacedGenerationContext{
		base:      base,
		namespace: namespace,
	}
}

type namespacedGenerationContext struct {
	base      GenerationContext
	namespace *typesystem.Namespace
}

func (ctx *namespacedGenerationContext) Namespace(ns *gir.Namespace) *typesystem.Namespace {
	return ctx.base.Namespace(ns)
}

func (ctx *namespacedGenerationContext) LookupType(typename, cType string) *typesystem.TypeMetadata {
	panic("")
}
