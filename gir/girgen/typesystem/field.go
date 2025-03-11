package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type Field struct {
	Doc Doc
	// CName is the original field name. This may not be valid in go, because C does not
	// reserve "type" as a name.
	CName string
	// CGoName is a sanitized name that does not interfere with go's keywords.
	// it should be used by the generator whenever in the go context.
	CGoName string
	Type    Type

	GoGetterName string
	GoSetterName string

	Readable bool
	Writable bool

	Bits int
}

func NewField(ns *Namespace, v gir.Field) *Field {
	if v.Private || !(v.IsReadable() || v.Writable) {
		return nil
	}

	var t Type
	var getterName string
	var setterName string

	// Callback fields are fields for virtual methods that can be overwritten
	//
	// this will be done by the bindings and not exposed to the user
	if v.Callback == nil {
		t = ns.findAnyType(v.AnyType)

		if t == nil {
			return nil
		}

		if v.IsReadable() {
			// TODO: check if this getter collides with any method
			getterName = strcases.SnakeToGo(true, v.Name) // no get prefix as per go convention
		}

		if v.Writable {
			// TODO: check if this setter collides with any method
			getterName = strcases.SnakeToGo(true, "set_"+v.Name)
		}
	}

	return &Field{
		Doc:      NewSimpleDoc(v.Doc),
		CName:    v.Name,
		CGoName:  DodgeReservedFieldName(v.Name),
		Type:     t,
		Readable: v.IsReadable(),
		Writable: v.Writable,

		GoGetterName: getterName,
		GoSetterName: setterName,
	}
}

// DodgeReservedFieldName replaces reserved go keywords with an underscore prefixed version
// the same way that cgo does (source needed)
func DodgeReservedFieldName(s string) string {
	if _, ok := GoKeywords[s]; ok {
		return "_" + s
	}

	return s
}
