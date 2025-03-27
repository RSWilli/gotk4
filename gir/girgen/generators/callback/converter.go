package callback

import (
	"strings"

	"github.com/diamondburned/gotk4/gir/girgen/file"
)

// Converter descibes a single parameter conversion for a C->go function call.
type Converter interface {
	GoIdentifier() string
	GoType() string

	CGoIdentifier() string
	CGoType() string

	// Metdata convey additional information about the conversion
	// this is a useful debugging information that will get output into the
	// generated variable declaration
	Metadata() string

	PreCallConvert(file.CodeWriter)

	// CallExpression is the expression passed to the C function call as a param.
	//
	// this may be different from the CGoIdentifier, because for out values we are passing a reference
	CallExpression() string

	PostCallConvert(file.CodeWriter)
}

type ConverterList []Converter

func (l ConverterList) Call() string {
	var expressions []string

	for _, c := range l {
		expressions = append(expressions, c.CallExpression())
	}

	return strings.Join(expressions, ", ")
}
