package generators

import (
	"fmt"
	"log"
	"strings"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/generators/callable"
	"github.com/diamondburned/gotk4/gir/girgen/generators/value"
	"github.com/diamondburned/gotk4/gir/girgen/gotmpl"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
	"github.com/diamondburned/gotk4/gir/girgen/types"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

var functionTmpl = gotmpl.NewGoTemplate(`
	{{ GoDoc . 0 }}
	func {{ .Name }}{{ .Tail }} {{ .Block }}
`)

// GenerateFunction generates the function call for the given GIR function.
func GenerateFunction(gen FileGeneratorWriter, fn *gir.Function) bool {
	return GeneratePrefixedFunction(gen, fn, "")
}

// GeneratePrefixedFunction generates the given GIR function with the prefix
// prepended into the name.
func GeneratePrefixedFunction(gen FileGeneratorWriter, fn *gir.Function, prefix string) bool {
	if fn.CIdentifier == "" || types.Filter(gen, fn.Name, fn.CIdentifier) {
		return false
	}

	typ := gir.TypeFindResult{
		NamespaceFindResult: gen.Namespace(),
		Type:                fn,
	}

	callableGen := callable.NewGenerator(gen)
	if !callableGen.Use(&typ, &fn.CallableAttrs) {
		return false
	}

	if prefix != "" {
		prefix = strcases.Go(prefix)

		// Check if this function is actually a constructor.
		if strings.HasPrefix(callableGen.Name, "New") {
			callableGen.Name = strings.TrimPrefix(callableGen.Name, "New")
			callableGen.Name = "New" + prefix + callableGen.Name
		} else {
			callableGen.Name = prefix + callableGen.Name
		}
	}

	callableGen.CoalesceTail()

	writer := FileWriterFromType(gen, fn)
	writer.Pen().WriteTmpl(functionTmpl, &callableGen)
	file.ApplyHeader(writer, &callableGen)
	return true
}

type FunctionGenerator struct {
	Doc Generator

	Converters value.ConverterList

	*typesystem.CallableSignature
}

// Generate implements Generator.
func (m *FunctionGenerator) Generate(w *file.Writer) {
	m.Doc.Generate(w)

	fmt.Fprintf(w.Go(), "%s {\n", m.GoSignature())

	// TODO converters

	if len(m.GoReturns) > 0 {
		fmt.Fprintf(w.Go(), "\n\treturn %s\n", m.GoReturns.GoIdentifiers())
	}

	fmt.Fprintf(w.Go(), "}\n\n")
}

func NewCallableGenerator(f *typesystem.CallableSignature) *FunctionGenerator {

	if f.Parameters.InstanceParam != nil && typesystem.IsForeignType(f.Parameters.InstanceParam.Type) {
		log.Panicf("received callable %s with foreign receiver %s", f.CIndentifier(), f.Parameters.InstanceParam.GoName)
	}

	return &FunctionGenerator{
		Doc:               NewCallableGoDocGenerator(f),
		CallableSignature: f,
		Converters:        make(value.ConverterList, 0), // TODO converters
	}
}
