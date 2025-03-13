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

func DeclareConstant(ctx context, v gir.Constant) *Constant {
	if ctx.skipType(v) {
		return nil
	}

	return &Constant{
		Doc: NewDoc(&v.InfoAttrs, &v.InfoElements),
		Identifier: &baseIdentifier{
			goIndentifier:  strcases.SnakeToGo(true, v.Name),
			cIndentifier:   v.Name,
			cGoIndentifier: "C." + v.Name,
		},
		GoValue: "C." + v.Name, // we use the C constant directly as this prevents any of the quoting issues
	}
}
