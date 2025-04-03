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
	if !v.IsIntrospectable() {
		return nil
	}

	if e.skip(nil, v) {
		return nil
	}

	// for some reason the name of the constant is listed under the ctype
	cIdentifier := v.CType

	ns, underlying := e.findType(&v.Type)

	if ns != nil {
		e.logger.Warn("skipping foreign constant")
		return nil
	}

	if underlying == nil {
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
