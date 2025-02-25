package value

import "github.com/diamondburned/gotk4/gir/girgen/typesystem"

func newStringReturnConverter(meta *typesystem.TypeMetadata) Converter {
	return NoopConverter{}
}
