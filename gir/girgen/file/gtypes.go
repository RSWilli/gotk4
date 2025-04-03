package file

import (
	"fmt"
	"io"

	"github.com/diamondburned/gotk4/gir/girgen/file/internal"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type gTypes []gType

type gType struct {
	typesystem.Marshalable
}

func (t gType) name() string {
	return fmt.Sprintf("GType%s", t.GoType(0))
}

func (ts gTypes) reader() io.Reader {
	var block internal.CodeWriter

	fmt.Fprintln(&block, "// GType values.")
	fmt.Fprintln(&block, "var (")
	block.Indent()

	var decls DeclarationWriter
	for _, t := range ts {
		fmt.Fprintf(&decls, "%s\t= glib.Type(C.%s())\n", t.name(), t.GLibGetType())
	}
	decls.WriteTo(&block)
	block.Unindent()

	fmt.Fprintln(&block, ")")

	fmt.Fprintln(&block)
	fmt.Fprintln(&block, "func init() {")
	block.Indent()
	fmt.Fprintln(&block, "glib.RegisterGValueMarshalers([]glib.TypeMarshaler{")
	block.Indent()

	for _, t := range ts {
		fmt.Fprintf(&block, "glib.TypeMarshaler{T: %s, F: marshal%s},\n", t.name(), t.GoType(0))
	}

	block.Unindent()

	fmt.Fprintln(&block, "})")
	block.Unindent()
	fmt.Fprintln(&block, "}")

	return &block
}
