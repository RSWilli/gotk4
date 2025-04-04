package convert

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

// UnimplementedConverter is used for any conversion that is not implemented.
//
// the conversion will panic, which means that this will generate fine, but crash at runtime
type UnimplementedConverter struct {
	Param *typesystem.Param
}

// Metadata implements Converter.
func (n *UnimplementedConverter) Metadata() string {
	return fmt.Sprintf(
		"%s, transfer: %s, scope: %s, implicit: %t, skip: %t, optional: %t, nullable: %t, caller-allocates: %t, has closure: %t, has destroy: %t",
		n.Param.Direction,
		n.Param.TransferOwnership,
		n.Param.Scope,
		n.Param.Implicit,
		n.Param.Skip,
		n.Param.Optional,
		n.Param.Nullable,
		n.Param.CallerAllocates,
		n.Param.Closure != nil,
		n.Param.Destroy != nil,
	)
}

// Convert implements Converter.
func (n *UnimplementedConverter) Convert(w file.File) {
	fmt.Fprintf(w.Go(), "panic(\"unimplemented conversion of %s (%s)\")\n", n.Param.GoType(), n.Param.CType())
}

var _ Converter = (*UnimplementedConverter)(nil)
