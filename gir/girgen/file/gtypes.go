package file

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/diamondburned/gotk4/gir/girgen/file/internal"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type gTypes []gType

type gType struct {
	typesystem.Type
}

func (t gType) name() string {
	return fmt.Sprintf("GType%s", t.GoType())
}

func (ts gTypes) reader() io.Reader {
	var block internal.CodeWriter

	fmt.Fprintln(&block, "// GType values.")
	fmt.Fprintln(&block, "var (")
	block.Indent()

	decls := tabwriter.NewWriter(&block, 0, 0, 1, ' ', 0)
	for _, t := range ts {
		fmt.Fprintf(decls, "%s\t= glib.Type(C.%s())\n", t.name(), t.GLibGetType())
	}
	decls.Flush()
	block.Unindent()

	fmt.Fprintln(&block, ")")

	fmt.Fprintln(&block)
	fmt.Fprintln(&block, "func init() {")
	block.Indent()
	fmt.Fprintln(&block, "glib.RegisterGValueMarshalers([]glib.TypeMarshaler{")
	block.Indent()

	for _, t := range ts {
		fmt.Fprintf(&block, "glib.TypeMarshaler{T: %s, F: %s},\n", t.name(), t.MarshalFuncName())
	}

	block.Unindent()

	fmt.Fprintln(&block, "})")
	block.Unindent()
	fmt.Fprintln(&block, "}")

	return &block
}
