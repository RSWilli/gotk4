package file

import (
	"fmt"
	"io"
	"strings"

	"github.com/diamondburned/gotk4/gir/girgen/file/internal"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

// externCallbacks holds all extern C declarations of callbacks.
type externCallbacks map[*typesystem.Callback]struct{}

// reader returns an io.Reader that declares the extern C trampoline functions to be used in the C preamble of the
// generated file
func (cbs externCallbacks) reader() io.Reader {
	if len(cbs) == 0 {
		return io.MultiReader()
	}

	var block internal.CodeWriter

	for cb := range cbs {
		var params []string

		for _, p := range cb.CParameters() {
			params = append(params, p.CType())
		}

		fmt.Fprintf(&block, "// extern %s %s(%s);\n", cb.CReturn.CType(), cb.TrampolineName, strings.Join(params, ", "))
	}

	return &block
}
