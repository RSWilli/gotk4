package typesystem

import (
	"strings"

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

	if _, ok := underlying.(*CastablePrimitive); !ok {
		return nil
	}

	return &Constant{
		Doc: NewDoc(&v.InfoAttrs, &v.InfoElements),
		Identifier: &baseIdentifier{
			goIndentifier:  strcases.SnakeToGo(true, strings.ToLower(v.Name)),
			cIndentifier:   cIdentifier,
			cGoIndentifier: "C." + cIdentifier,
		},
		GoValue: v.Value,
	}
}
