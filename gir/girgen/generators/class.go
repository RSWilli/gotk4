package generators

import (
	"fmt"
	"log"

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

	g.Doc.Generate(w.Go())

	fmt.Fprintf(w.Go(), "type %s struct {\n", g.GoType())
	fmt.Fprintf(w.Go(), "\t_ [0]func() // equal guard\n")
	fmt.Fprintf(w.Go(), "\t%s\n", g.Parent.GoType())
	if len(g.Implements) > 0 {
		fmt.Fprintf(w.Go(), "\t// implemented interfaces:\n")
	}
	for _, inter := range g.Implements {
		fmt.Fprintf(w.Go(), "\t%s\n", inter.GoType())
	}
	fmt.Fprintf(w.Go(), "}\n\n")

	fmt.Fprintf(w.Go(), "var _ %s = (*%s)(nil)\n\n", g.GoInterfaceName, g.GoType())

	fmt.Fprintf(w.Go(), "// %s is the interface that is implemented by all types  extending %s \n", g.GoInterfaceName, g.GoType())
	fmt.Fprintf(w.Go(), "type %s interface {\n", g.GoInterfaceName)
	w.Go().Indent()
	fmt.Fprintln(w.Go(), g.ParentGoInterfaceName())
	for inter := range g.ImplementedGoInterfaceNames() {
		fmt.Fprintln(w.Go(), inter)
	}
	fmt.Fprintf(w.Go(), "%s() *%s\n", g.GoPrivateUpcastMethod, g.GoType())
	fmt.Fprintln(w.Go())

	g.Methods.GenerateInterfaceSignatures(w.Go())

	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n\n")

	g.generateWrapFunction(w.Go())

	if g.Marshaler != nil {
		w.RegisterGType(g)
		g.Marshaler.Generate(w)
		fmt.Fprintln(w.Go())
	}

	// TODO: imports

	mkConstructor := func(constructorName, baseConstructorName string) {
		fmt.Fprintf(w.Go(), "func %s(c unsafe.Pointer) %s {\n", constructorName, g.GoInterfaceName)
		fmt.Fprintf(w.Go(), "\treturn %s(c).(%s)\n", baseConstructorName, g.GoInterfaceName)
		fmt.Fprintf(w.Go(), "}\n\n")
	}

	fmt.Fprintf(w.Go(), "// %s is used to convert raw %s pointers to go. This is used by the bindings internally.\n", g.GoUnsafeFromGlibBorrowFunction, g.CType())
	mkConstructor(g.GoUnsafeFromGlibBorrowFunction, g.BaseClassGoUnsafeFromGlibBorrowFunction())
	fmt.Fprintf(w.Go(), "// %s is used to convert raw %s pointers to go while taking a reference and attaching a finalizer. This is used by the bindings internally.\n", g.GoUnsafeFromGlibNoneFunction, g.CType())
	mkConstructor(g.GoUnsafeFromGlibNoneFunction, g.BaseClassGoUnsafeFromGlibNoneFunction())
	fmt.Fprintf(w.Go(), "// %s is used to convert raw %s pointers to go while attaching a finalizer. This is used by the bindings internally.\n", g.GoUnsafeFromGlibFullFunction, g.CType())
	mkConstructor(g.GoUnsafeFromGlibFullFunction, g.BaseClassGoUnsafeFromGlibFullFunction())

	fmt.Fprintf(w.Go(), "func (%s *%s) %s() *%s {\n", strcases.ReceiverName(g.GoType()), g.GoType(), g.GoPrivateUpcastMethod, g.GoType())
	fmt.Fprintf(w.Go(), "\treturn %s\n", strcases.ReceiverName(g.GoType()))
	fmt.Fprintf(w.Go(), "}\n\n")

	mkTransfer := func(transfername, baseTransferName string) {
		fmt.Fprintf(w.Go(), "func %s(c %s) unsafe.Pointer {\n", transfername, g.GoInterfaceName)
		fmt.Fprintf(w.Go(), "\treturn %s(c)\n", baseTransferName)
		fmt.Fprintf(w.Go(), "}\n\n")
	}

	fmt.Fprintf(w.Go(), "// %s is used to convert the instance to it's C value %s. This is used by the bindings internally.\n", g.GoUnsafeToGlibNoneFunction, g.CType())
	mkTransfer(g.GoUnsafeToGlibNoneFunction, g.BaseClassGoUnsafeToGlibNoneFunction())

	fmt.Fprintf(w.Go(), "// %s is used to convert the instance to it's C value %s, while removeing the finalizer. This is used by the bindings internally.\n", g.GoUnsafeToGlibNoneFunction, g.CType())
	mkTransfer(g.GoUnsafeToGlibNoneFunction, g.BaseClassGoUnsafeToGlibNoneFunction())

	GenerateAll(
		w,
		g.Constructors,
		g.Functions,
		g.Methods,
	)
}

func (g *ClassGenerator) generateWrapFunction(w file.CodeWriter) {
	baseClassIdentifier := "base"

	fmt.Fprintf(w, "func %s(%s *%s) *%s {\n", g.GoWrapBaseClassFunction, baseClassIdentifier, g.BaseClass().GoType(), g.GoType())
	w.Indent()
	fmt.Fprintf(w, "return &%s{\n", g.GoType())
	w.Indent()

	wrapClass(w, g.Class.Parent, baseClassIdentifier)

	w.Indent()
	for _, inter := range g.Implements {
		wrapInterface(w, inter, baseClassIdentifier)
	}
	w.Unindent()

	fmt.Fprintf(w, "}\n")
	w.Unindent()

	fmt.Fprintf(w, "}\n\n")
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
func wrapClass(w file.CodeWriter, t typesystem.Type, baseClassIdentifier string) {
	parent := typesystem.GetClassParent(t)

	if parent == nil {
		// current type is the base class
		fmt.Fprintf(w, "%s: *%s,\n", typesystem.UnderlyingType(t).GoType(), baseClassIdentifier)

		w.Unindent()
		return
	}

	var currentNs *typesystem.Namespace
	var implementedInterfaces []typesystem.Type

	switch t := t.(type) {
	case *typesystem.ForeignType:
		currentNs = t.SourceNamespace
		implementedInterfaces = t.Type.(*typesystem.Class).Implements
	case *typesystem.Class:
		implementedInterfaces = t.Implements
	default:
		log.Panicf("unexpected class parent %T", t)
	}

	fmt.Fprintf(w, "%s: %s{\n", typesystem.UnderlyingType(t).GoType(), t.GoType())

	w.Indent()
	defer w.Unindent()

	switch t := parent.(type) {
	case *typesystem.ForeignType:
		wrapClass(w, t, baseClassIdentifier)
	case *typesystem.Class:
		// the class is foreign if resolved from another foreign type
		if currentNs == nil {
			wrapClass(w, t, baseClassIdentifier)
		} else {
			foreign := &typesystem.ForeignType{
				SourceNamespace: currentNs,
				Type:            t,
			}

			wrapClass(w, foreign, baseClassIdentifier)
		}
	default:
		log.Panicf("unexpected class parent %T", t)
	}

	w.Indent()
	for _, inter := range implementedInterfaces {
		wrapInterface(w, inter, baseClassIdentifier)
	}
	w.Unindent()

	fmt.Fprintf(w, "},\n")
}
