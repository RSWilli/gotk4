package generators

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type InterfaceGenerator struct {
	Doc SubGenerator

	*typesystem.Interface

	Marshaler Generator

	// infos used by sub generators:
	ReceiverName string

	// sub generators:
	Constructors GeneratorList
	Functions    GeneratorList
	Methods      MethodGeneratorList
}

func (g *InterfaceGenerator) Generate(w *file.Package) {
	w.GoImport("unsafe")
	w.GoImport("runtime")

	fmt.Fprintf(w.Go(), "// %s is the instance type used by all types implementing %s. It is used internally by the bindings. Users should use the interface [%s] instead.\n", g.GoType(0), g.CType(0), g.GoInterfaceName)
	fmt.Fprintf(w.Go(), "type %s struct {\n", g.GoType(0))
	fmt.Fprintf(w.Go(), "\t_ [0]func() // equal guard\n")
	fmt.Fprintf(w.Go(), "\t*%s\n", g.Parent.NamespacedGoType(0))
	// for _, inter := range g.Prerequesite {
	// 	fmt.Fprintf(w.Go(), "\t*%s\n", inter.GoType(0))
	// }
	fmt.Fprintf(w.Go(), "}\n\n")

	fmt.Fprintf(w.Go(), "var _ %s = (*%s)(nil)\n\n", g.GoInterfaceName, g.GoType(0))

	g.Doc.Generate(w.Go())
	fmt.Fprintf(w.Go(), "type %s interface {\n", g.GoInterfaceName)
	w.Go().Indent()
	fmt.Fprintln(w.Go(), g.Parent.WithForeignNamespace(g.Parent.Type.GoInterfaceName))
	// for inter := range g.PrerequesitesGoInterfaceNames() {
	// 	fmt.Fprintln(w.Go(), inter)
	// }

	w.Go().NewSection()

	g.Methods.GenerateInterfaceSignatures(w.Go())

	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n\n")

	baseClassIdentifier := "base"

	fmt.Fprintf(w.Go(), "func %s(%s *%s) *%s {\n", g.GoWrapBaseClassFunction, baseClassIdentifier, g.Parent.WithForeignNamespace(g.Parent.Type.GoType(0)), g.GoType(0))
	w.Go().Indent()
	fmt.Fprintf(w.Go(), "return &%s{\n", g.GoType(0))
	w.Go().Indent()

	fmt.Fprintf(w.Go(), "%s: %s,\n", g.Parent.Type.GoType(0), baseClassIdentifier)
	w.Go().Unindent()

	fmt.Fprintf(w.Go(), "}\n")
	w.Go().Unindent()

	fmt.Fprintf(w.Go(), "}\n\n")

	if g.Marshaler != nil {
		w.RegisterGType(g)
		g.Marshaler.Generate(w)
		fmt.Fprintln(w.Go())
	}

	// TODO: imports

	mkConstructor := func(constructorName, parentConstructorName string) {
		fmt.Fprintf(w.Go(), "func %s(c unsafe.Pointer) %s {\n", constructorName, g.GoInterfaceName)
		w.Go().Indent()
		fmt.Fprintf(w.Go(), "return %s(c).(%s)\n", parentConstructorName, g.GoInterfaceName)
		w.Go().Unindent()
		fmt.Fprintf(w.Go(), "}\n\n")
	}

	fmt.Fprintf(w.Go(), "// %s is used to convert raw %s pointers to go. This is used by the bindings internally.\n", g.GoUnsafeFromGlibBorrowFunction(), g.CType(0))
	mkConstructor(g.GoUnsafeFromGlibBorrowFunction(), g.Parent.WithForeignNamespace(g.Parent.Type.GoUnsafeFromGlibBorrowFunction()))
	fmt.Fprintf(w.Go(), "// %s is used to convert raw %s pointers to go while taking a reference and attaching a finalizer. This is used by the bindings internally.\n", g.GoUnsafeFromGlibNoneFunction(), g.CType(0))
	mkConstructor(g.GoUnsafeFromGlibNoneFunction(), g.Parent.WithForeignNamespace(g.Parent.Type.GoUnsafeFromGlibNoneFunction()))
	fmt.Fprintf(w.Go(), "// %s is used to convert raw %s pointers to go while attaching a finalizer. This is used by the bindings internally.\n", g.GoUnsafeFromGlibFullFunction(), g.CType(0))
	mkConstructor(g.GoUnsafeFromGlibFullFunction(), g.Parent.WithForeignNamespace(g.Parent.Type.GoUnsafeFromGlibFullFunction()))

	mkTransfer := func(transfername, baseTransferName string) {
		fmt.Fprintf(w.Go(), "func %s(c %s) unsafe.Pointer {\n", transfername, g.GoInterfaceName)
		fmt.Fprintf(w.Go(), "\treturn %s(c)\n", baseTransferName)
		fmt.Fprintf(w.Go(), "}\n\n")
	}

	fmt.Fprintf(w.Go(), "// %s is used to convert the instance to it's C value %s. This is used by the bindings internally.\n", g.GoUnsafeToGlibNoneFunction(), g.CType(0))
	mkTransfer(g.GoUnsafeToGlibNoneFunction(), g.Parent.WithForeignNamespace(g.Parent.Type.GoUnsafeToGlibNoneFunction()))

	fmt.Fprintf(w.Go(), "// %s is used to convert the instance to it's C value %s, while removeing the finalizer. This is used by the bindings internally.\n", g.GoUnsafeToGlibFullFunction(), g.CType(0))
	mkTransfer(g.GoUnsafeToGlibFullFunction(), g.Parent.WithForeignNamespace(g.Parent.Type.GoUnsafeToGlibFullFunction()))

	GenerateAll(
		w,
		g.Constructors,
		g.Functions,
		g.Methods,
	)
}

func NewInterfaceGenerator(c *typesystem.Interface) *InterfaceGenerator {
	var marshaler Generator

	if c.GLibGetType() != "" {
		marshaler = NewMarshalObjectGenerator(c, c.GoWrapBaseClassFunction)
	}

	g := &InterfaceGenerator{
		Doc:       NewTypeGoDocGenerator(c),
		Interface: c,
		Marshaler: marshaler,
	}

	for _, fn := range c.Functions {
		if fGen := NewCallableGenerator(fn); fGen != nil {
			g.Functions = append(g.Functions, fGen)
		}
	}

	for _, method := range c.Methods {
		if methGen := NewCallableGenerator(method); methGen != nil {
			g.Methods = append(g.Methods, methGen)
		}
	}

	return g
}

func wrapInterface(w file.CodeWriter, t typesystem.CouldBeForeign[*typesystem.Interface], baseClassIdentifier string) {
	fmt.Fprintf(w, "%s: %s{\n", t.Type.GoType(0), t.NamespacedGoType(0))

	w.Indent()

	// parent is always the base class
	fmt.Fprintf(w, "%s: *%s,\n", t.Type.Parent.Type.GoType(0), baseClassIdentifier)

	w.Unindent()

	fmt.Fprintf(w, "},\n")
}
