package generators

import (
	"log"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

func GetParents(ctx gencontext.GenerationContext, class gir.Class) []*typesystem.TypeMetadata {
	var parents []*typesystem.TypeMetadata

	current := &class

	for current.Parent != "" {
		parentMeta := ctx.LookupType(current.Parent, "")

		if parentMeta == nil {
			log.Printf("could not find parent class %s\n", current.Parent)
			return nil
		}

		parentClass, ok := parentMeta.GirType.(*gir.Class)

		if !ok {
			log.Printf("parent class of %s is of type %T\n", current.Name, parentMeta.GirType)
			return nil
		}

		parents = append(parents, parentMeta)

		ctx = gencontext.Namespaced(ctx, parentMeta.Namespace)

		current = parentClass
	}

	return parents
}
