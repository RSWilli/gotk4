package generators

import (
	"fmt"
	"log"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type MethodGenerator interface {
	Generator
	// GoMethodDeclaration is used to get the method signature from the sub generator to put them in the class interface
	GoMethodDeclaration() string
}

type MethodGeneratorList []MethodGenerator

var _ Generator = MethodGeneratorList{}

func (list MethodGeneratorList) Generate(w *file.Writer) {
	for _, g := range list {
		if g == nil {
			continue
		}

		g.Generate(w)
	}
}

type ClassGenerator struct {
	Doc Generator

	*typesystem.Class

	Marshaler Generator

	// infos used by sub generators:
	ReceiverName string

	// sub generators:
	Constructors GeneratorList
	Getters      GeneratorList
	Setters      GeneratorList
	Methods      MethodGeneratorList
}

func (g *ClassGenerator) Generate(w *file.Writer) {
	w.GoImportCoreGlib()
	w.GoImport("unsafe")
	w.GoImport("runtime")

	g.Doc.Generate(w)

	fmt.Fprintf(w.Go(), "type %s struct {\n", g.GoType())
	fmt.Fprintf(w.Go(), "\t_ [0]func() // equal guard\n")
	fmt.Fprintf(w.Go(), "\t%s\n", g.Parent.GoType())
	fmt.Fprintf(w.Go(), "}\n\n")

	if g.Marshaler != nil {
		g.Marshaler.Generate(w)
	}

	// TODO: imports
	parents := g.AllParents()

	fmt.Fprintf(w.Go(), "func %s(obj *coreglib.Object) *%s {\n", g.GoWrapFunctionName, g.GoType())
	fmt.Fprintf(w.Go(), "\treturn &%s{\n", g.GoType())
	for _, p := range parents[:len(parents)-1] {
		fmt.Fprintf(w.Go(), "\t%s: %s{\n", typesystem.UnderlyingType(p).GoType(), p.GoType())
	}

	fmt.Fprintf(w.Go(), "\t%s: obj,\n", typesystem.UnderlyingType(parents[len(parents)-1]).GoType())

	for range parents[:len(parents)-1] {
		fmt.Fprintf(w.Go(), "\t},\n")
	}

	fmt.Fprintf(w.Go(), "\t}\n")
	fmt.Fprintf(w.Go(), "}\n\n")

	GenerateAll(
		w,
		g.Constructors,
		g.Getters,
		g.Setters,
	)
}

func NewClassGenerator(c *typesystem.Class) *ClassGenerator {
	if !c.Valid {
		return nil
	}

	if !extendsGlibObject(c) {
		log.Printf("skipping class %s because it does not extend GObject\n", c.GoType())
		return nil // FIXME: this is not a problem per se, but tricky.
	}

	if c.Parent == nil {
		log.Printf("skipping generation of %s because it is GObject\n", c.GoType())
		return nil // FIXME: intercept this type in the typesystem
	}

	var marshaler Generator

	if c.GLibGetType() != "" {
		marshaler = NewMarshalObjectGenerator(c, c.GoWrapFunctionName)
	}

	g := &ClassGenerator{
		Doc:       NewTypeGoDocGenerator(c, 0),
		Class:     c,
		Marshaler: marshaler,

		ReceiverName: strcases.ReceiverName(c.GoType()),
	}

	// for _, constructor := range r.Constructors {
	// 	if constGen := NewRecordConstructorGenerator(ctx, g, constructor); constGen != nil {
	// 		g.Constructors = append(g.Constructors, constGen)
	// 	}
	// }

	// for _, method := range r.Methods {
	// 	if method.Name == "weak_ref" || method.Name == "weak_unref" {
	// 		continue
	// 	}

	// 	if method.Name == "ref" {
	// 		g.CgoRefFunction = "C." + method.CIdentifier
	// 		continue
	// 	}

	// 	if method.Name == "unref" {
	// 		g.CgoUnrefFunction = "C." + method.CIdentifier
	// 		continue
	// 	}

	// 	if methGen := NewRecordMethodGenerator(ctx, g, method); methGen != nil {
	// 		g.Methods = append(g.Methods, methGen)
	// 	}
	// }

	return g
}

func extendsGlibObject(c *typesystem.Class) bool {
	if c.CType() == "GObject" {
		return true
	}

	if c.Parent == nil {
		return false
	}

	switch parent := c.Parent.(type) {
	case *typesystem.Class:
		return extendsGlibObject(parent)
	case *typesystem.ForeignType:
		return foreignExtendsGlibObject(parent)
	}

	return false
}

func foreignExtendsGlibObject(c *typesystem.ForeignType) bool {
	if c.CType() == "GObject" {
		return true
	}

	switch parent := c.Type.(type) {
	case *typesystem.Class:
		return extendsGlibObject(parent)
	case *typesystem.ForeignType:
		return foreignExtendsGlibObject(parent)
	}

	return false
}
