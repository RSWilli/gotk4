package value

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type BooleanReturnConverter struct {
	GoReturnIdent  string
	GoType         string
	CGoReturnIdent string
	CgoType        string
}

func (b *BooleanReturnConverter) CGoParameterDecl() string       { return "" }
func (b *BooleanReturnConverter) CGoParameterIdentifier() string { return "" }
func (b *BooleanReturnConverter) GoParameterDecl() string        { return "" }
func (b *BooleanReturnConverter) GoParameterIdentifier() string  { return "" }

func (b *BooleanReturnConverter) CGoReturnDecl() string {
	return fmt.Sprintf("%s %s", b.CGoReturnIdent, b.CgoType)
}

func (b *BooleanReturnConverter) CGoReturnIdentifier() string {
	return b.CGoReturnIdent
}

func (b *BooleanReturnConverter) GoReturnDecl() string {
	return fmt.Sprintf("%s %s", b.GoReturnIdent, b.GoType)
}

func (b *BooleanReturnConverter) GoReturnIdentifier() string {
	return b.GoReturnIdent
}

func (b *BooleanReturnConverter) Generate(importer, FunctionCallSections) {
	panic("unimplemented")
}

func newBooleanReturnConverter(meta *typesystem.TypeMetadata) Converter {
	return &BooleanReturnConverter{
		GoReturnIdent:  "ok",
		CGoReturnIdent: "cret",
		GoType:         meta.GoType,
		CgoType:        meta.CGoType,
	}
}

func newBooleanParamConverter(meta *typesystem.TypeMetadata, param gir.Parameter) Converter {
	return NoopConverter{}
}
