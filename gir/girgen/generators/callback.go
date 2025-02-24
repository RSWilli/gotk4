package generators

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/generators/callback"
	"github.com/diamondburned/gotk4/gir/girgen/generators/value"
	"github.com/diamondburned/gotk4/gir/girgen/gotmpl"
)

var callbackTmpl = gotmpl.NewGoTemplate(`
	{{ GoDoc . 0 }}
	type {{ .GoName }} func{{ .GoTail }}
`)

var callbackExportTmpl = gotmpl.NewGoTemplate(`
	//export {{ .CGoName }}
	func {{ .CGoName }}{{ .CGoTail }} {{ .Block }}
`)

// GenerateCallback generates a callback type declaration and handler into the
// given file generator.
func GenerateCallback(gen FileGeneratorWriter, cb *gir.Callback) bool {
	generator := callback.NewGenerator(gen)
	generator.Parent = cb
	if !generator.Use(&cb.CallableAttrs) {
		return false
	}

	{
		writer := FileWriterFromType(gen, cb)
		writer.Pen().WriteTmpl(callbackTmpl, &generator)
		for _, result := range generator.Results {
			result.Resolved.ImportPubl(gen, writer.Header())
		}
	}

	{
		writer := FileWriterExportedFromType(gen, cb)
		writer.Pen().WriteTmpl(callbackExportTmpl, &generator)
		file.ApplyHeader(writer, &generator)
	}

	return true
}

type CallbackGenerator struct {
	Doc     Generator
	GoName  string
	CgoName string

	// Values contains all the parameter generators. These are executed in order and are used
	// for converting the cgo parameters to go, and the go returns to cgo. This is
	Values value.ConverterList

	// CgoParameters contains the same generators as Values. They are used to generate the cgo function signature.
	//
	// It is not recommended to reorder them, because this will most likely break the types
	CGoParameters value.ConverterList

	// CGoReturn contains the return param of the cgo function. C can only return one value.
	CGoReturn value.Converter

	// GoParameters contains the same generators as values, but here the girgen user can reorder them via the hook stage.
	//
	// common reorderings are already applied, e.g. moving context to the front
	GoParameters value.ConverterList

	// GoReturns contains the same generators as Values, but is used to generate the go return values.
	GoReturns value.ConverterList
}

// Generate implements Generator.
func (c *CallbackGenerator) Generate(w *file.Writer) {
	c.generateGo(w)
	c.generateExport(w)
}

func (c *CallbackGenerator) generateGo(w *file.Writer) {
	c.Doc.Generate(w)
	fmt.Fprintf(w.Go(), "type %s func(%s)%s\n", c.GoName, c.GoParameters.GoParameterDeclList(), c.GoReturns.GoReturnDeclList())

	fmt.Fprintln(w.Go())
}

func (c *CallbackGenerator) generateExport(w *file.Writer) {
	fmt.Fprintf(w.Exported.Go(), "//export %s\n", c.CgoName)
	fmt.Fprintf(w.Exported.Go(), "func %s(%s)%s {\n", c.CgoName, c.CGoParameters.CGoParameterDeclList(), c.CGoReturn.CGoReturnDecl())

	sections := value.NewFunctionCallSections()

	fmt.Fprintf(sections.FnCall(), "%s = fn(%s)\n", c.GoReturns.GoReturnIdentifierList(), c.GoParameters.GoParameterIdentifierList())

	// generate all value conversions:
	for _, value := range c.Values {
		value.Generate(&w.Exported, sections)
	}

	// write the grouped sections:
	sections.WriteTo(w.Exported.Go())

	fmt.Fprintln(w.Exported.Go(), "}")
	fmt.Fprintln(w.Exported.Go())
}

func NewCallbackGenerator(ctx gencontext.GenerationContext, cb gir.Callback) *CallbackGenerator {
	meta := ctx.LookupType(cb.Name)

	if meta == nil {
		return nil
	}

	var valueCount int

	if cb.Parameters != nil {
		valueCount += len(cb.Parameters.Parameters)
	}

	if cb.ReturnValue != nil {
		valueCount += 1
	}

	if cb.Parameters != nil && cb.Parameters.InstanceParameter != nil {
		// valueCount += 1

		panic("unimplemented instance param for callbacks")
	}

	g := &CallbackGenerator{
		Doc:     NewGoDocGenerator(cb, 0),
		GoName:  meta.GoType,
		CgoName: meta.CGoType,

		CGoReturn:     value.NoopConverter{},
		Values:        make(value.ConverterList, 0, valueCount),
		CGoParameters: make(value.ConverterList, 0, valueCount),
		GoParameters:  make(value.ConverterList, 0, valueCount),
		GoReturns:     make(value.ConverterList, 0, valueCount),
	}

	if cb.Parameters != nil {
		for i, param := range cb.Parameters.Parameters {
			conv := value.NewParamConverter(ctx, i, param)

			if conv == nil {
				// conversion not possible, skip callback
				return nil
			}

			// any converter could decide to e.g. move an "out" param into the go returns, so we add them
			// to all of the ConverterLists

			g.Values = append(g.Values, conv)
			g.CGoParameters = append(g.CGoParameters, conv)
			g.GoParameters = append(g.GoParameters, conv)
			g.GoReturns = append(g.GoReturns, conv)
		}
	}

	if cb.ReturnValue != nil {
		conv := value.NewReturnConverter(ctx, *cb.ReturnValue)

		if conv == nil {
			// conversion not possible, skip callback
			return nil
		}

		g.CGoReturn = conv

		g.Values = append(g.Values, conv)
		g.CGoParameters = append(g.CGoParameters, conv)
		g.GoParameters = append(g.GoParameters, conv)
		g.GoReturns = append(g.GoReturns, conv)
	}

	// TODO: perform common reorderings here, e.g. move context to front.

	return g
}
