package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type Constant struct {
	Doc
	Identifier

	GoValue string
}

func DeclareConstant(e *env, v gir.Constant) *Constant {
	if e.skipType(v) { // TODO: a constant is not a type
		return nil
	}

	// for some reason the name of the constant is listed under the ctype
	cIdentifier := v.CType

	underlying := e.findType(&v.Type)

	if underlying == nil {
		return nil
	}

	if Pointers(underlying) > 0 {
		return nil
	}

	if underlying == Utf8 {
		return nil
	}

	return &Constant{
		Doc: NewDoc(&v.InfoAttrs, &v.InfoElements),
		Identifier: &baseIdentifier{
			goIndentifier:  strcases.SnakeToGo(true, v.Name),
			cIndentifier:   cIdentifier,
			cGoIndentifier: "C." + cIdentifier,
		},
		GoValue: "C." + cIdentifier, // we use the C constant directly as this prevents any of the quoting issues
	}
}
