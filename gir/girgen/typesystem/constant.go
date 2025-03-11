package typesystem

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type Constant struct {
	Doc     Doc
	GoName  string
	GoValue string
}

func DeclareConstant(ns *Namespace, v gir.Constant) *Constant {
	if !v.IsIntrospectable() {
		return nil
	}

	return &Constant{
		Doc:     NewDoc(&v.InfoAttrs, &v.InfoElements),
		GoName:  strcases.SnakeToGo(true, v.Name),
		GoValue: "C." + v.CType, // we use the C constant directly as this prevents any of the quoting issues
	}
}
