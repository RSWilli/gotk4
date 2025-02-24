package gencontext

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

// GenerationContext contains the information that is useful for every generator in the tree. It gets passed
// through the generator constructors.
type GenerationContext interface {
	NamespaceMetadata(ns *gir.Namespace) *typesystem.NamespaceMetadata
	LookupType(typ string) *typesystem.TypeMetadata
}

type baseGenerationContext struct {
	Typesystem *typesystem.Registry
}

func Base(typesystem *typesystem.Registry) GenerationContext {
	return &baseGenerationContext{
		Typesystem: typesystem,
	}
}

func (ctx *baseGenerationContext) NamespaceMetadata(ns *gir.Namespace) *typesystem.NamespaceMetadata {
	return ctx.Typesystem.GetNamespaceMetadata(ns)
}

func (ctx *baseGenerationContext) LookupType(typ string) *typesystem.TypeMetadata {
	return ctx.Typesystem.LookupType(typ)
}

func Namespaced(namespace *typesystem.NamespaceMetadata, base GenerationContext) GenerationContext {
	return &namespacedGenerationContext{
		base:          base,
		namespaceMeta: namespace,
	}
}

type namespacedGenerationContext struct {
	base          GenerationContext
	namespaceMeta *typesystem.NamespaceMetadata
}

func (ctx *namespacedGenerationContext) NamespaceMetadata(ns *gir.Namespace) *typesystem.NamespaceMetadata {
	return ctx.base.NamespaceMetadata(ns)
}

func (ctx *namespacedGenerationContext) LookupType(typ string) *typesystem.TypeMetadata {
	return ctx.namespaceMeta.Namespace.LookupType(typ)
}
