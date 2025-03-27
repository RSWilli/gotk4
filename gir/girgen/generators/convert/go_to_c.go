package convert

import "github.com/diamondburned/gotk4/gir/girgen/typesystem"

func NewGoToCConverter(p *typesystem.Param) Converter {
	if p.Implicit {
		return &ImplicitConverter{
			Param: p,
		}
	}
	if p.Skip {
		return &SkippedConverter{
			Param: p,
		}
	}

	if p.CallerAllocates {
		return &UnimplementedConverter{
			Param: p,
		}
	}

	if p.Direction == "out" {
		return &UnimplementedConverter{
			Param: p,
		}
	}

	switch p.Type {
	case typesystem.Guint:

	}

	return &UnimplementedConverter{
		Param: p,
	}
}
