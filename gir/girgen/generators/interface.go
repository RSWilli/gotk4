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

func (g *InterfaceGenerator) Generate(w *file.Writer) {
	w.GoImport("unsafe")
	w.GoImport("runtime")

	g.Doc.Generate(w.Go())

	fmt.Fprintf(w.Go(), "type %s struct {\n", g.GoType())
	fmt.Fprintf(w.Go(), "\t_ [0]func() // equal guard\n")
	fmt.Fprintf(w.Go(), "\t*%s\n", g.Parent.GoType())
	// for _, inter := range g.Prerequesite {
	// 	fmt.Fprintf(w.Go(), "\t*%s\n", inter.GoType())
	// }
	fmt.Fprintf(w.Go(), "}\n\n")

	fmt.Fprintf(w.Go(), "var _ %s = (*%s)(nil)\n\n", g.GoInterfaceName, g.GoType())

	fmt.Fprintf(w.Go(), "type %s interface {\n", g.GoInterfaceName)
	w.Go().Indent()
	fmt.Fprintln(w.Go(), g.ParentGoInterfaceName())
	// for inter := range g.PrerequesitesGoInterfaceNames() {
	// 	fmt.Fprintln(w.Go(), inter)
	// }
	// fmt.Fprintln(w.Go())

	g.Methods.GenerateInterfaceSignatures(w.Go())

	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n\n")

	baseClassIdentifier := "base"

	fmt.Fprintf(w.Go(), "func %s(%s *%s) *%s {\n", g.GoWrapBaseClassFunction, baseClassIdentifier, g.Parent.GoType(), g.GoType())
	w.Go().Indent()
	fmt.Fprintf(w.Go(), "return &%s{\n", g.GoType())
	w.Go().Indent()

	fmt.Fprintf(w.Go(), "%s: %s,\n", typesystem.UnderlyingType(g.Parent).GoType(), baseClassIdentifier)
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
		fmt.Fprintf(w.Go(), "func %s(c unsafe.Pointer) *%s {\n", constructorName, g.GoType())
		w.Go().Indent()
		fmt.Fprintf(w.Go(), "base := %s(c)\n", parentConstructorName)
		fmt.Fprintf(w.Go(), "return %s(base)\n", g.GoWrapBaseClassFunction)
		w.Go().Unindent()
		fmt.Fprintf(w.Go(), "}\n\n")
	}

	fmt.Fprintf(w.Go(), "// %s is used to convert raw %s pointers to go. This is used by the bindings internally.\n", g.GoUnsafeBorrowFunction, g.CType())
	mkConstructor(g.GoUnsafeBorrowFunction, g.ParentGoUnsafeBorrowFunction())
	fmt.Fprintf(w.Go(), "// %s is used to convert raw %s pointers to go while taking a reference and attaching a finalizer. This is used by the bindings internally.\n", g.GoUnsafeTransferNoneFunction, g.CType())
	mkConstructor(g.GoUnsafeTransferNoneFunction, g.ParentGoUnsafeTransferNoneFunction())
	fmt.Fprintf(w.Go(), "// %s is used to convert raw %s pointers to go while attaching a finalizer. This is used by the bindings internally.\n", g.GoUnsafeTransferFullFunction, g.CType())
	mkConstructor(g.GoUnsafeTransferFullFunction, g.ParentGoUnsafeTransferFullFunction())

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

func wrapInterface(w file.CodeWriter, t typesystem.Type, baseClassIdentifier string) {
	fmt.Fprintf(w, "%s: %s{\n", typesystem.UnderlyingType(t).GoType(), t.GoType())

	var parent typesystem.Type

	switch t := t.(type) {
	case *typesystem.ForeignType:
		parent = t.Type.(*typesystem.Interface).Parent
	case *typesystem.Interface:
		parent = t.Parent
	default:
		panic("invalid type")
	}

	w.Indent()

	// parent is always the base class
	fmt.Fprintf(w, "%s: *%s,\n", typesystem.UnderlyingType(parent).GoType(), baseClassIdentifier)

	w.Unindent()

	fmt.Fprintf(w, "},\n")
}
