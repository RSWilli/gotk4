package generators

import (
	"bytes"
	"fmt"
	"io"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/generators/callback"
	"github.com/diamondburned/gotk4/gir/girgen/generators/value"
	"github.com/diamondburned/gotk4/gir/girgen/gotmpl"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
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
	Doc Generator

	*typesystem.Callback

	Converters value.ConverterList
}

// Generate implements Generator.
func (c *CallbackGenerator) Generate(w *file.Writer) {
	c.generateGo(w)
	c.generateExport(w)
}

func (c *CallbackGenerator) generateGo(w *file.Writer) {
	c.Doc.Generate(w)

	// TODO: the caller must declare the extern C function trampoline, because it can be referenced from another package, see _gotk4_glib2_CompareDataFunc

	ret := c.GoReturns.GoDeclarations()

	if ret != "" {
		ret = " (" + ret + ")"
	}

	fmt.Fprintf(w.Go(), "type %s func(%s)%s\n", c.GoType(), c.GoParameters.GoDeclarations(), ret)

	fmt.Fprintln(w.Go())
}

func (c *CallbackGenerator) generateExport(w *file.Writer) {
	w.Exported.GoImportCore("gbox")

	fmt.Fprintf(w.Exported.Go(), "//export %s\n", c.TrampolineName)

	cret := ""
	if c.CReturn != nil {
		cret = fmt.Sprintf(" (%s %s)", c.CReturn.CName, c.CReturn.Type.CGoType())
	}

	fmt.Fprintf(w.Exported.Go(), "func %s(%s)%s {\n", c.TrampolineName, c.CParameters.CGoDeclarations(), cret)

	var secInputPre bytes.Buffer
	var secInputConv bytes.Buffer
	var secFnCall bytes.Buffer
	var secOutputPre bytes.Buffer
	var secOutputConv bytes.Buffer
	var secReturn bytes.Buffer

	goReturns := c.GoReturns.GoIdentifiers()

	if goReturns == "" {
		fmt.Fprintf(&secFnCall, "\tfn(%s)\n", c.GoParameters.GoIdentifiers())
	} else {
		fmt.Fprintf(&secFnCall, "\t%s := fn(%s)\n", goReturns, c.GoParameters.GoIdentifiers())
	}
	if c.CReturn != nil {
		fmt.Fprintf(&secReturn, "\treturn %s\n", c.CReturn.CName)
	}

	// generate all value conversions:
	// for _, v := range c.Converters {
	// v.AddImports(&w.Exported)
	// if v.ConversionDirection() == value.ConvertCToGo {
	// 	fmt.Fprintf(&secInputPre, "\tvar %s %s // out\n", v.OutIdentifier(), v.OutType())
	// 	fmt.Fprint(&secInputConv, v.Conversion())
	// } else if goReturns != "" {
	// 	fmt.Fprintf(&secOutputPre, "\tvar _ %s\n", v.InType()) // this is not needed, but here for debugging purposes
	// 	fmt.Fprint(&secOutputConv, v.Conversion())
	// }
	// }

	// separate the sections with newlines:
	fmt.Fprintln(&secInputPre)
	fmt.Fprintln(&secInputConv)
	fmt.Fprintln(&secFnCall)
	fmt.Fprintln(&secOutputPre)
	fmt.Fprintln(&secOutputConv)

	// write the grouped sections:
	io.Copy(w.Exported.Go(), io.MultiReader(&secInputPre, &secInputConv, &secFnCall, &secOutputPre, &secOutputConv, &secReturn))

	fmt.Fprintln(w.Exported.Go(), "}")
	fmt.Fprintln(w.Exported.Go())
}

func NewCallbackGenerator(cb *typesystem.Callback) *CallbackGenerator {
	if cb.InstanceParam != nil {
		panic("callback with instance param unimplemented")
	}

	g := &CallbackGenerator{
		Doc:      NewTypeGoDocGenerator(cb),
		Callback: cb,

		Converters: make(value.ConverterList, 0), // TODO: convert
	}

	return g
}
