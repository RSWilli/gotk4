package generators

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type RecordGenerator struct {
	Doc               SubGenerator
	GenerateMarshaler bool

	*typesystem.Record

	// infos used by sub generators:
	ReceiverName string

	// sub generators:
	SubGenerators GeneratorList
}

func (g *RecordGenerator) Generate(w *file.Package) {
	if g.CgoUnrefFunction == "" {
		panic("cannot generate record without an unref method")
	}

	w.GoImport("unsafe")
	w.GoImport("runtime")

	g.Doc.Generate(w.Go())

	if g.IsTypeStructFor != nil {
		fmt.Fprintf(w.Go(), "// \n")
		fmt.Fprintf(w.Go(), "// %s is the type struct for [%s]\n", g.GoType(0), g.IsTypeStructFor.GoType(1))
	}

	// TODO: attach a cleanup field here of type runtime.Cleanup, and drop the SetFinalizer for AddCleanup
	fmt.Fprintf(w.Go(), "type %s struct {\n", g.GoType(0))
	fmt.Fprintf(w.Go(), "\t*%s\n", g.PrivateGoType)
	fmt.Fprintf(w.Go(), "}\n\n")

	fmt.Fprintf(w.Go(), "// %s is the struct that's finalized\n", g.PrivateGoType)
	fmt.Fprintf(w.Go(), "type %s struct {\n", g.PrivateGoType)
	fmt.Fprintf(w.Go(), "\tnative *%s\n", g.CGoType(0))
	fmt.Fprintf(w.Go(), "}\n\n")

	if g.GenerateMarshaler {
		// GoValueInitializer assertion:
		fmt.Fprintf(w.Go(), "var _ %s = (*%s)(nil)\n\n", g.Value().WithForeignNamespace("GoValueInitializer"), g.GoType(0))

		w.RegisterGType(g.Record)
		fmt.Fprintf(w.Go(), "func marshal%s(p unsafe.Pointer) (interface{}, error) {\n", g.GoType(0))
		fmt.Fprintf(w.Go(), "\tb := %s(p).Boxed()\n", g.Value().WithForeignNamespace(g.Value().Type.FromGlibBorrowFunction))
		fmt.Fprintf(w.Go(), "\treturn %s(b), nil\n", g.GoUnsafeFromGlibNoneFunction())
		fmt.Fprintf(w.Go(), "}\n\n")

		fmt.Fprintf(w.Go(), "func (r *%s) GoValueType() %s {\n", g.GoType(0), g.Type().NamespacedGoType(0))
		w.Go().Indent()
		fmt.Fprintf(w.Go(), "return %s\n", g.GoTypeName())
		w.Go().Unindent()
		fmt.Fprintf(w.Go(), "}\n\n")

		fmt.Fprintf(w.Go(), "func (r *%s) SetGoValue(v *%s) {\n", g.GoType(0), g.Value().NamespacedGoType(0))
		w.Go().Indent()
		fmt.Fprintf(w.Go(), "v.SetBoxed(unsafe.Pointer(r.native))\n")
		w.Go().Unindent()
		fmt.Fprintf(w.Go(), "}\n\n")
	}

	fmt.Fprintf(w.Go(), "// %s is used to convert raw %s pointers to go. This is used by the bindings internally.\n", g.GoUnsafeFromGlibBorrowFunction(), g.CGoType(0))
	fmt.Fprintf(w.Go(), "func %s(p unsafe.Pointer) *%s {\n", g.GoUnsafeFromGlibBorrowFunction(), g.GoType(0))
	fmt.Fprintf(w.Go(), "\treturn &%s{&%s{(*%s)(p)}}\n", g.GoType(0), g.PrivateGoType, g.CGoType(0))
	fmt.Fprintf(w.Go(), "}\n\n")

	mkFinalizer := func() {
		w.GoImportCore("profile")
		w.Go().Indent()

		fmt.Fprintf(w.Go(), "profile.Track(uintptr(unsafe.Pointer(wrapped.%s)), 1)\n", g.PrivateGoType)

		fmt.Fprintf(w.Go(), "runtime.SetFinalizer(\n")
		fmt.Fprintf(w.Go(), "\twrapped.%s,\n", g.PrivateGoType)
		fmt.Fprintf(w.Go(), "\tfunc (intern *%s) {\n", g.PrivateGoType)
		w.Go().Indent()
		g.unrefCall(w.Go(), "intern")
		fmt.Fprintf(w.Go(), "\tprofile.Untrack(uintptr(unsafe.Pointer(intern)))\n")
		w.Go().Unindent()
		fmt.Fprintf(w.Go(), "\t},\n")
		fmt.Fprintf(w.Go(), ")\n")
		w.Go().Unindent()
	}

	if g.GoUnsafeFromGlibNoneFunction() != "" {
		fmt.Fprintf(w.Go(), "// %s is used to convert raw %s pointers to go without transferring ownership. This is used by the bindings internally.\n", g.GoUnsafeFromGlibNoneFunction(), g.CGoType(0))
		fmt.Fprintf(w.Go(), "func %s(p unsafe.Pointer) *%s {\n", g.GoUnsafeFromGlibNoneFunction(), g.GoType(0))
		if g.CgoRefFunction != "" {
			// from none only refs if reffing is possible: TODO: this can produce bugs because we are borrowing otherwise
			fmt.Fprintf(w.Go(), "\t%s((*%s)(p))\n", g.CgoRefFunction, g.CGoType(0))
		} else {
			fmt.Fprintf(w.Go(), "\t// FIXME: this has no ref function, what should we do here?\n")
		}
		fmt.Fprintf(w.Go(), "\twrapped := %s(p)\n", g.GoUnsafeFromGlibBorrowFunction())
		mkFinalizer()
		fmt.Fprintf(w.Go(), "\treturn wrapped\n")
		fmt.Fprintf(w.Go(), "}\n\n")
	}

	if g.GoUnsafeFromGlibFullFunction() != "" {
		fmt.Fprintf(w.Go(), "// %s is used to convert raw %s pointers to go while taking ownership. This is used by the bindings internally.\n", g.GoUnsafeFromGlibFullFunction(), g.CGoType(0))
		fmt.Fprintf(w.Go(), "func %s(p unsafe.Pointer) *%s {\n", g.GoUnsafeFromGlibFullFunction(), g.GoType(0))
		fmt.Fprintf(w.Go(), "\twrapped := %s(p)\n", g.GoUnsafeFromGlibBorrowFunction())
		mkFinalizer()
		fmt.Fprintf(w.Go(), "\treturn wrapped\n")
		fmt.Fprintf(w.Go(), "}\n\n")
	}

	if g.CgoRefFunction != "" {
		fmt.Fprintf(w.Go(), "// %s increases the refcount on the underlying resource. This is used by the bindings internally.\n", g.GoUnsafeRefFunction)
		fmt.Fprintf(w.Go(), "// \n")
		fmt.Fprintf(w.Go(), "// When this is called without an associated call to [%s.%s], then [%s] will leak memory.\n", g.GoType(0), g.GoUnsafeUnrefFunction, g.GoType(0))
		fmt.Fprintf(w.Go(), "func %s(%s *%s) {\n", g.GoUnsafeRefFunction, g.ReceiverName, g.GoType(0))
		fmt.Fprintf(w.Go(), "\t%s(%s.native)\n", g.CgoRefFunction, g.ReceiverName)
		fmt.Fprintf(w.Go(), "}\n\n")
	}

	fmt.Fprintf(w.Go(), "// %s unrefs/frees the underlying resource. This is used by the bindings internally.\n", g.GoUnsafeUnrefFunction)
	fmt.Fprintf(w.Go(), "// \n")
	fmt.Fprintf(w.Go(), "// After this is called, no other method on [%s] is expected to work anymore.\n", g.GoType(0))
	fmt.Fprintf(w.Go(), "func %s(%s *%s) {\n", g.GoUnsafeUnrefFunction, g.ReceiverName, g.GoType(0))
	g.unrefCall(w.Go(), g.ReceiverName)
	fmt.Fprintf(w.Go(), "}\n\n")

	fmt.Fprintf(w.Go(), "// %s returns the underlying C pointer. This is used by the bindings internally.\n", g.GoUnsafeToGlibNoneFunction())
	fmt.Fprintf(w.Go(), "func %s(%s *%s) unsafe.Pointer {\n", g.GoUnsafeToGlibNoneFunction(), g.ReceiverName, g.GoType(0))
	fmt.Fprintf(w.Go(), "\treturn unsafe.Pointer(%s.native)\n", g.ReceiverName)
	fmt.Fprintf(w.Go(), "}\n\n")

	if g.GoUnsafeToGlibFullFunction() != "" {
		fmt.Fprintf(w.Go(), "// %s returns the underlying C pointer and gives up ownership.\n", g.GoUnsafeToGlibFullFunction())
		fmt.Fprintf(w.Go(), "// This is used by the bindings internally.\n")
		fmt.Fprintf(w.Go(), "func %s(%s *%s) unsafe.Pointer {\n", g.GoUnsafeToGlibFullFunction(), g.ReceiverName, g.GoType(0))
		fmt.Fprintf(w.Go(), "\truntime.SetFinalizer(%s.%s, nil)\n", g.ReceiverName, g.PrivateGoType)
		fmt.Fprintf(w.Go(), "\t_p := unsafe.Pointer(%s.native)\n", g.ReceiverName)
		fmt.Fprintf(w.Go(), "\t%s.native = nil // %s is invalid from here on\n", g.ReceiverName, g.GoType(0))
		fmt.Fprintf(w.Go(), "\treturn _p\n")
		fmt.Fprintf(w.Go(), "}\n\n")
	}

	// for typestructs we need a type cast to the parent type struct
	parentTypeStruct := g.ParentTypeStruct()
	if parentTypeStruct != nil {
		fmt.Fprintf(w.Go(), "// ParentClass returns the type struct of the parent class of this type struct.\n")
		fmt.Fprintf(w.Go(), "// This essentially casts the underlying c pointer.\n")
		fmt.Fprintf(w.Go(), "func (%s *%s) ParentClass() %s {\n", g.ReceiverName, g.GoType(0), parentTypeStruct.NamespacedGoType(1))
		w.Go().Indent()
		fmt.Fprintf(w.Go(), "parent := %s(%s(%s))\n", parentTypeStruct.WithForeignNamespace(parentTypeStruct.Type.FromGlibBorrowFunction), g.ToGlibNoneFunction, g.ReceiverName)
		fmt.Fprintf(w.Go(), "// attach a cleanup to keep the instance alive as long as the parent is referenced\n")
		fmt.Fprintf(w.Go(), "runtime.AddCleanup(parent, func(_ %s) {}, %s)\n", g.GoType(1), g.ReceiverName)
		fmt.Fprintf(w.Go(), "return parent\n")
		w.Go().Unindent()
		fmt.Fprintf(w.Go(), "}\n\n")
	}

	GenerateAll(
		w,
		g.SubGenerators,
	)
}

func (g *RecordGenerator) unrefCall(w file.CodeWriter, variable string) {
	if g.CgoUnrefNeedsUnsafeCast {
		// unsafe already imported above
		fmt.Fprintf(w, "\t%s(unsafe.Pointer(%s.native))\n", g.CgoUnrefFunction, variable)
	} else {
		fmt.Fprintf(w, "\t%s(%s.native)\n", g.CgoUnrefFunction, variable)
	}
}

func NewRecordGenerator(r *typesystem.Record) *RecordGenerator {
	g := &RecordGenerator{
		Doc:               NewGoDocGenerator(r),
		Record:            r,
		GenerateMarshaler: r.GLibGetType() != "",

		ReceiverName: strcases.ReceiverName(r.GoType(0)),
	}

	for _, constructor := range r.Constructors {
		if constGen := NewCallableGenerator(constructor); constGen != nil {
			g.SubGenerators = append(g.SubGenerators, constGen)
		}
	}

	for _, method := range r.Functions {
		if methGen := NewCallableGenerator(method); methGen != nil {
			g.SubGenerators = append(g.SubGenerators, methGen)
		}
	}

	for _, method := range r.Methods {
		if methGen := NewCallableGenerator(method); methGen != nil {
			g.SubGenerators = append(g.SubGenerators, methGen)
		}
	}

	return g
}
