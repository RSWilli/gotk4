package generators

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/gotmpl"
	"github.com/diamondburned/gotk4/gir/girgen/logger"
	"github.com/diamondburned/gotk4/gir/girgen/types"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

// Deprecated: old
var constantTmpl = gotmpl.NewGoTemplate(strings.TrimSpace(`
	{{- GoDoc . 0 TrailingNewLine (OverrideSelfName .Name) -}}
	const {{ .Name }} = {{ .Value }}
`))

// Deprecated: old
type constantData struct {
	*gir.Constant
	Value string
}

// Deprecated: old
func GenerateConstant(gen FileGeneratorWriter, constant *gir.Constant) bool {
	goType := types.GIRBuiltinGo(constant.Type.Name)
	if goType == "" {
		gen.Logln(logger.Debug, "unknown constant type", constant.Type.Name)
		return false
	}

	if types.Filter(gen, constant.Name, constant.CType) {
		gen.Logln(logger.Debug, "filtered constant", constant.CType)
		return false
	}

	value := constant.Value
	if goType == "string" {
		value = strconv.Quote(value)
	}

	writer := FileWriterFromType(gen, constant)
	writer.Pen().WriteTmpl(constantTmpl, &constantData{
		Constant: constant,
		Value:    value,
	})

	return true
}

type ConstantGenerator struct {
	Doc SubGenerator

	*typesystem.Constant
}

func (g *ConstantGenerator) Generate(w *file.Writer) {
	g.Doc.Generate(w.Go())

	fmt.Fprintf(w.Go(), "const %s = %s\n", g.GoIndentifier(), g.GoValue)
}

func NewConstantGenerator(constant *typesystem.Constant) *ConstantGenerator {
	return &ConstantGenerator{
		Doc:      NewIdentifierGoDocGenerator(constant),
		Constant: constant,
	}
}
