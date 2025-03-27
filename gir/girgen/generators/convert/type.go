package convert

import (
	"github.com/diamondburned/gotk4/gir/girgen/file"
)

// Converter descibes a single parameter conversion for a go->C function call.
type Converter interface {
	// Metdata convey additional information about the conversion
	// this is a useful debugging information that will get output into the
	// generated variable declaration
	Metadata() string

	Convert(file.CodeWriter)
}

type ConverterList []Converter
