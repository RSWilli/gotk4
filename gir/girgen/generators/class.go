package generators

import (
	"fmt"
	"log"
	"strings"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
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
	Doc          Generator
	GoName       string
	CGoType      string
	GoNameParent string

	Marshaler              Generator
	GoWrapCoreObjectFnName string

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

	fmt.Fprintf(w.Go(), "type %s struct {\n", g.GoName)
	fmt.Fprintf(w.Go(), "\t_ [0]func() // equal guard\n")
	fmt.Fprintf(w.Go(), "\t%s\n", g.GoNameParent)
	fmt.Fprintf(w.Go(), "}\n\n")

	fmt.Fprintf(w.Go(), "// %s is the struct that's finalized\n", g.GoNameParent)
	fmt.Fprintf(w.Go(), "type %s struct {\n", g.GoNameParent)
	fmt.Fprintf(w.Go(), "\tnative *%s\n", g.CGoType)
	fmt.Fprintf(w.Go(), "}\n\n")

	if g.Marshaler != nil {
		g.Marshaler.Generate(w)
	}

	fmt.Fprintf(w.Go(), "func %s(obj *coreglib.Object) *%s {\n", g.GoWrapCoreObjectFnName, g.GoName)
	fmt.Fprintf(w.Go(), "\tpanic(\"TODO\")\n")
	fmt.Fprintf(w.Go(), "}\n\n")

	GenerateAll(
		w,
		g.Constructors,
		g.Getters,
		g.Setters,
	)
}

func NewClassGenerator(ctx gencontext.GenerationContext, c gir.Class) *ClassGenerator {
	if !c.IsIntrospectable() || strings.HasSuffix(c.Name, "Private") {
		return nil
	}

	meta := ctx.LookupType(c.Name, c.CType)

	if meta == nil {
		return nil
	}

	parents := GetParents(ctx, c)

	if len(parents) == 0 && c.Parent != "" {
		log.Printf("skipping class %s because parent %s is unknown\n", c.Name, c.Parent)
		return nil
	}

	if len(parents) == 0 || parents[len(parents)-1].GoBaseType != "gobject.Object" {
		log.Printf("FIXME: skipping class %s because it doesn't extend GObject\n", c.Name)
		return nil
	}

	var marshaler Generator
	wrapFn := "wrap" + meta.GoBaseType

	if c.GLibGetType != "" {
		marshaler = NewMarshalObjectGenerator(meta.GoBaseType, "wrap"+meta.GoBaseType)
	}

	// goPrivate := firstToLower(meta.GoBaseType)

	g := &ClassGenerator{
		Doc:                    NewGoDocGenerator(meta.GoBaseType, c, 0),
		GoName:                 meta.GoBaseType,
		CGoType:                meta.CGoBaseType,
		GoNameParent:           parents[0].GoBaseType,
		Marshaler:              marshaler,
		GoWrapCoreObjectFnName: wrapFn,

		ReceiverName: strcases.FirstLetter(meta.GoBaseType),
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
