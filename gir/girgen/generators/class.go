package generators

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type ClassGenerator struct {
	Doc SubGenerator

	*typesystem.Class

	Marshaler Generator

	// sub generators:
	Constructors GeneratorList
	Functions    GeneratorList
	Methods      MethodGeneratorList
}

func (g *ClassGenerator) Generate(w *file.Package) {
	w.GoImport("unsafe")

	fmt.Fprintf(w.Go(), "// %s is the instance type used by all types extending %s. It is used internally by the bindings. Users should use the interface [%s] instead.\n", g.GoType(0), g.CType(0), g.GoInterfaceName)
	fmt.Fprintf(w.Go(), "type %s struct {\n", g.GoType(0))
	w.Go().Indent()
	fmt.Fprintf(w.Go(), "_ [0]func() // equal guard\n")

	w.GoImportNamespace(g.Parent.Namespace)
	fmt.Fprintf(w.Go(), "%s\n", g.Parent.NamespacedGoType(0))

	if len(g.Implements) > 0 {
		fmt.Fprintf(w.Go(), "// implemented interfaces:\n")
	}
	for _, inter := range g.Implements {
		w.GoImportNamespace(inter.Namespace)
		// don't use the interface type for the instance struct
		fmt.Fprintln(w.Go(), inter.NamespacedGoType(0))
	}
	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n\n")

	fmt.Fprintf(w.Go(), "var _ %s = (*%s)(nil)\n\n", g.GoInterfaceName, g.GoType(0))

	g.Doc.Generate(w.Go())
	fmt.Fprintf(w.Go(), "type %s interface {\n", g.GoInterfaceName)
	w.Go().Indent()
	if g.ManuallyExtended {
		fmt.Fprintf(w.Go(), "%sExtManual // handwritten functions\n", g.GoInterfaceName)
	}
	fmt.Fprintln(w.Go(), g.ParentGoInterfaceName())
	for inter := range g.ImplementedGoInterfaceNames() {
		fmt.Fprintln(w.Go(), inter)
	}
	fmt.Fprintf(w.Go(), "%s() *%s\n", g.GoPrivateUpcastMethod, g.GoType(0))
	w.Go().NewSection()

	g.Methods.GenerateInterfaceSignatures(w)

	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n\n")

	g.generateWrapFunction(w)

	if g.Marshaler != nil {
		w.RegisterGType(g)
		g.Marshaler.Generate(w)
		fmt.Fprintln(w.Go())
	}

	mkConstructor := func(constructorName, baseConstructorName string) {
		fmt.Fprintf(w.Go(), "func %s(c unsafe.Pointer) %s {\n", constructorName, g.GoInterfaceName)
		fmt.Fprintf(w.Go(), "\treturn %s(c).(%s)\n", baseConstructorName, g.GoInterfaceName)
		fmt.Fprintf(w.Go(), "}\n\n")
	}

	fmt.Fprintf(w.Go(), "// %s is used to convert raw %s pointers to go while taking a reference and attaching a finalizer. This is used by the bindings internally.\n", g.GoUnsafeFromGlibNoneFunction(), g.CType(0))
	mkConstructor(g.GoUnsafeFromGlibNoneFunction(), g.BaseClassGoUnsafeFromGlibNoneFunction())
	fmt.Fprintf(w.Go(), "// %s is used to convert raw %s pointers to go while attaching a finalizer. This is used by the bindings internally.\n", g.GoUnsafeFromGlibFullFunction(), g.CType(0))
	mkConstructor(g.GoUnsafeFromGlibFullFunction(), g.BaseClassGoUnsafeFromGlibFullFunction())

	fmt.Fprintf(w.Go(), "func (%s *%s) %s() *%s {\n", strcases.ReceiverName(g.GoType(0)), g.GoType(0), g.GoPrivateUpcastMethod, g.GoType(0))
	fmt.Fprintf(w.Go(), "\treturn %s\n", strcases.ReceiverName(g.GoType(0)))
	fmt.Fprintf(w.Go(), "}\n\n")

	mkTransfer := func(transfername, baseTransferName string) {
		fmt.Fprintf(w.Go(), "func %s(c %s) unsafe.Pointer {\n", transfername, g.GoInterfaceName)
		fmt.Fprintf(w.Go(), "\treturn %s(c)\n", baseTransferName)
		fmt.Fprintf(w.Go(), "}\n\n")
	}

	fmt.Fprintf(w.Go(), "// %s is used to convert the instance to it's C value %s. This is used by the bindings internally.\n", g.GoUnsafeToGlibNoneFunction(), g.CType(0))
	mkTransfer(g.GoUnsafeToGlibNoneFunction(), g.BaseClassGoUnsafeToGlibNoneFunction())

	fmt.Fprintf(w.Go(), "// %s is used to convert the instance to it's C value %s, while removeing the finalizer. This is used by the bindings internally.\n", g.GoUnsafeToGlibFullFunction(), g.CType(0))
	mkTransfer(g.GoUnsafeToGlibFullFunction(), g.BaseClassGoUnsafeToGlibFullFunction())

	GenerateAll(
		w,
		g.Constructors,
		g.Functions,
		g.Methods,
	)
}

func (g *ClassGenerator) generateWrapFunction(w file.File) {
	baseClassIdentifier := "base"
	baseClass := g.BaseClass()

	w.GoImportNamespace(baseClass.Namespace)

	fmt.Fprintf(w.Go(), "func %s(%s *%s) *%s {\n", g.GoWrapBaseClassFunction, baseClassIdentifier, baseClass.WithForeignNamespace(baseClass.Type.GoType(0)), g.GoType(0))
	w.Go().Indent()
	fmt.Fprintf(w.Go(), "return &%s{\n", g.GoType(0))
	w.Go().Indent()

	wrapClass(w, typesystem.CouldBeForeign[*typesystem.Class]{Namespace: nil, Type: g.Class}, baseClassIdentifier)

	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n")
	w.Go().Unindent()

	fmt.Fprintf(w.Go(), "}\n\n")
}

func NewClassGenerator(c *typesystem.Class) *ClassGenerator {
	var marshaler Generator

	if c.GLibGetType() != "" {
		marshaler = NewMarshalObjectGenerator(c, c.GoWrapBaseClassFunction)
	}

	g := &ClassGenerator{
		Doc:       NewTypeGoDocGenerator(c),
		Class:     c,
		Marshaler: marshaler,
	}

	for _, constructor := range c.Constructors {
		g.Constructors = append(g.Constructors, NewCallableGenerator(constructor))
	}

	for _, fn := range c.Functions {
		g.Functions = append(g.Functions, NewCallableGenerator(fn))

	}

	for _, method := range c.Methods {
		g.Methods = append(g.Methods, NewCallableGenerator(method))
	}

	return g
}

// wrapClass generates the tree like structure needed to construct the whole struct
// the given type is always the current class, and the function will output the Struct fields
// to fill the parent and implemented interfaces
func wrapClass(w file.File, t typesystem.CouldBeForeign[*typesystem.Class], baseClassIdentifier string) {
	parent := t.Type.Parent

	if parent.Type == nil {
		panic("parent.Type is nil, this means wrapClass was called on the baseclass")
	}

	if parent.Type.Parent.Type == nil {
		// current parent is the base class
		fmt.Fprintf(w.Go(), "%s: *%s,\n", parent.Type.GoType(0), baseClassIdentifier)
	} else {
		if parent.Namespace == nil && t.Namespace != nil {
			// parent of t is local to t, but not to the current file
			parent.Namespace = t.Namespace
		}

		w.GoImportNamespace(parent.Namespace)

		fmt.Fprintf(w.Go(), "%s: %s{\n", parent.Type.GoType(0), parent.NamespacedGoType(0))
		w.Go().Indent()
		wrapClass(w, parent, baseClassIdentifier)
		w.Go().Unindent()
		fmt.Fprintf(w.Go(), "},\n")
	}

	for _, inter := range t.Type.Implements {
		wrapInterface(w.Go(), inter, baseClassIdentifier)
	}

}
