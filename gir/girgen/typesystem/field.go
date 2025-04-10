package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type Field struct {
	Doc

	Parent Type

	Identifier

	Type          CouldBeForeign[Type]
	CTypePointers int

	GoGetterName string
	GoSetterName string

	Readable bool
	Writable bool

	Bits int
}

func NewField(e *env, parent Type, v gir.Field) *Field {
	e = e.sub("field", v.Name)

	if e.skip(parent, v) {
		return nil
	}

	if v.Private || !(v.IsReadable() || v.Writable) {
		return nil
	}

	if v.Bits > 0 {
		return nil // TODO: what does bits mean?
	}

	var t CouldBeForeign[Type]
	var getterName string
	var setterName string
	var pointers int

	// Callback fields are fields for virtual methods that can be overwritten
	//
	// this will be done by the bindings and not exposed to the user
	if v.Callback == nil {
		ns, typ := e.findAnyType(v.AnyType)

		if typ == nil {
			return nil
		}

		t = CouldBeForeign[Type]{
			Namespace: ns,
			Type:      typ,
		}

		pointers = CountCTypePointers(CTypeFromAnytype(v.AnyType))

		if v.IsReadable() {
			// TODO: check if this getter collides with any method
			getterName = strcases.SnakeToGo(true, v.Name) // no get prefix as per go convention
		}

		if v.Writable {
			// TODO: check if this setter collides with any method
			setterName = strcases.SnakeToGo(true, "set_"+v.Name)
		}
	}

	return &Field{
		Doc:    NewSimpleDoc(v.Doc),
		Parent: parent,
		Identifier: &baseIdentifier{
			cIndentifier:   v.Name,
			cGoIndentifier: strcases.CGoField(v.Name),
			goIndentifier:  strcases.CGoField(v.Name),
		},
		Bits:          v.Bits,
		Type:          t,
		Readable:      v.IsReadable(),
		Writable:      v.Writable,
		CTypePointers: pointers,

		GoGetterName: getterName,
		GoSetterName: setterName,
	}
}
