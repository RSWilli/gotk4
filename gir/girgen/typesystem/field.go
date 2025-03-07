package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
)

type Field struct {
	Doc   Doc
	CName string
	Type  Type

	Readable bool
	Writable bool

	Bits int
}

func NewField(ns *Namespace, v gir.Field) *Field {
	if v.Private || !(v.IsReadable() || v.Writable) {
		return nil
	}

	if v.Callback != nil {
		return nil
	}

	t := ns.findAnyType(v.AnyType)

	if t == nil {
		return nil
	}

	return &Field{
		Doc:      NewSimpleDoc(v.Doc),
		CName:    DodgeReservedFieldName(v.Name),
		Type:     t,
		Readable: v.IsReadable(),
		Writable: v.Writable,
	}
}

// DodgeReservedFieldName replaces reserved go keywords with an underscore prefixed version
// the same way that cgo does (source needed)
func DodgeReservedFieldName(s string) string {
	if s, ok := GoKeywords[s]; ok {
		return "_" + s
	}

	return s
}
