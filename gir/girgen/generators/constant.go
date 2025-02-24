package generators

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/gotmpl"
	"github.com/diamondburned/gotk4/gir/girgen/logger"
	"github.com/diamondburned/gotk4/gir/girgen/types"
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
	c    gir.Constant
	Doc  Generator
	Name string
	// Value is directly printed, must be quoted if it's a string
	Value string
}

func (g *ConstantGenerator) Generate(w *file.Writer) {
	g.Doc.Generate(w)

	fmt.Fprintf(w.Go(), "const %s = %s\n", g.Name, g.Value)
}

func NewConstantGenerator(ctx gencontext.GenerationContext, constant gir.Constant) *ConstantGenerator {
	if !constant.IsIntrospectable() {
		return nil
	}

	name := constant.Name

	resolvedType := ctx.LookupType(constant.Type.CType)

	if resolvedType == nil {
		log.Printf("skipping constant %s because %s did not map to a known go type", constant.Name, constant.Type.Name)
		return nil
	}

	goValue := constant.Value

	if resolvedType.GoType == "string" {
		goValue = strconv.Quote(goValue)
	}

	return &ConstantGenerator{
		c:     constant,
		Doc:   NewGoDocGenerator(constant, 0),
		Name:  name,
		Value: goValue,
	}
}
