package convert

import "github.com/diamondburned/gotk4/gir/girgen/typesystem"

func NewCToGoConverter(p *typesystem.Param) Converter {
	return &UnimplementedConverter{
		Param: p,
	}
}
