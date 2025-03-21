package file

import (
	"bytes"
	"fmt"
	"io"

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
	maxlen := 0

	for _, t := range ts {
		maxlen = max(len(t.name()), maxlen)
	}

	var buf bytes.Buffer

	fmt.Fprintln(&buf, "// GType values.")
	fmt.Fprintln(&buf, "var (")

	for _, t := range ts {
		fmt.Fprintf(&buf, "\t%-*s = coreglib.Type(C.%s())\n", maxlen, t.name(), t.GLibGetType())
	}

	fmt.Fprintln(&buf, ")")

	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "func init() {")
	fmt.Fprintln(&buf, "\tcoreglib.RegisterGValueMarshalers([]coreglib.TypeMarshaler{")

	for _, t := range ts {
		fmt.Fprintf(&buf, "\t\tcoreglib.TypeMarshaler{T: %s, F: %s},\n", t.name(), t.MarshalFuncName())
	}

	fmt.Fprintln(&buf, "\t})")
	fmt.Fprintln(&buf, "}")

	return &buf
}
