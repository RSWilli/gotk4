package generators

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

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

	// CReturnType contains the actual c return type needed for the extern declare
	CReturnType string
	// CParamTypes contains the actual c param types needed for the extern declare
	CParamTypes []string

	// ClosureArg
	ClosureArg string

	// Converters contains all the parameter converters. These are executed in order and are used
	// for converting the cgo parameters to go, and the go returns to cgo.
	Converters value.ConverterList

	// CgoParameters contains the converters as Converters used to generate the cgo function signature.
	//
	// It is not recommended to reorder them, because this will most likely break the types
	CGoParameters value.ConverterList

	// CGoReturn contains the return param of the cgo function. C can only return one value.
	CGoReturn value.Converter

	// GoParameters contains the converters as Converters used to generate the go parameters. Compared to CGoParameters there might
	// be some parameters missing, e.g. out params get turned into return values in go. The girgen user is free to reorder this list.
	//
	// common reorderings are already applied, e.g. moving context to the front
	GoParameters value.ConverterList

	// GoReturns contains the same converters as Converters, but is used to generate the go return values.
	GoReturns value.ConverterList
}

// Generate implements Generator.
func (c *CallbackGenerator) Generate(w *file.Writer) {
	c.generateGo(w)
	c.generateExport(w)
}

func (c *CallbackGenerator) generateGo(w *file.Writer) {
	c.Doc.Generate(w)

	// FIXME: maybe this should be declared at the caller site, because it can be referenced from another package, see _gotk4_glib2_CompareDataFunc
	fmt.Fprintf(w.C(), "extern %s %s(%s);\n", c.CReturnType, c.CgoName, strings.Join(c.CParamTypes, ", "))

	ret := c.GoReturns.GoDeclList()

	if ret != "" {
		ret = " (" + ret + ")"
	}

	fmt.Fprintf(w.Go(), "type %s func(%s)%s\n", c.GoName, c.GoParameters.GoDeclList(), ret)

	fmt.Fprintln(w.Go())
}

func (c *CallbackGenerator) generateExport(w *file.Writer) {
	w.Exported.GoImportCore("gbox")

	fmt.Fprintf(w.Exported.Go(), "//export %s\n", c.CgoName)

	cret := ""
	if c.CGoReturn != nil {
		// CGoReturn converts go->c, so the values are in "Out"
		cret = fmt.Sprintf(" (%s %s)", c.CReturnType, c.CGoReturn.OutType())
	}

	fmt.Fprintf(w.Exported.Go(), "func %s(%s)%s {\n", c.CgoName, c.CGoParameters.CDeclList(), cret)

	fmt.Fprintf(w.Exported.Go(), "\tvar fn %s\n", c.GoName)
	fmt.Fprintf(w.Exported.Go(), "\t{\n")
	fmt.Fprintf(w.Exported.Go(), "\t\tv := gbox.Get(uintptr(%s))\n", c.ClosureArg)
	fmt.Fprintf(w.Exported.Go(), "\t\tif v == nil {\n")
	fmt.Fprintf(w.Exported.Go(), "\t\t\tpanic(`callback not found`)\n")
	fmt.Fprintf(w.Exported.Go(), "\t\t}\n")
	fmt.Fprintf(w.Exported.Go(), "\t\tfn = v.(%s)\n", c.GoName)
	fmt.Fprintf(w.Exported.Go(), "\t}\n\n")

	var secInputPre bytes.Buffer
	var secInputConv bytes.Buffer
	var secFnCall bytes.Buffer
	var secOutputPre bytes.Buffer
	var secOutputConv bytes.Buffer
	var secReturn bytes.Buffer

	goReturns := c.GoReturns.InIdentifierList()

	if goReturns == "" {
		fmt.Fprintf(&secFnCall, "\tfn(%s)\n", c.GoParameters.OutIdentifierList())
	} else {
		fmt.Fprintf(&secFnCall, "\t%s := fn(%s)\n", goReturns, c.GoParameters.OutIdentifierList())
	}
	if c.CGoReturn != nil {
		fmt.Fprintf(&secReturn, "\treturn %s\n", c.CGoReturn.OutIdentifier())
	}

	// generate all value conversions:
	for _, v := range c.Converters {
		v.AddImports(&w.Exported)
		if v.ConversionDirection() == value.ConvertCToGo {
			fmt.Fprintf(&secInputPre, "\tvar %s %s // out\n", v.OutIdentifier(), v.OutType())
			fmt.Fprint(&secInputConv, v.Conversion())
		} else if goReturns != "" {
			fmt.Fprintf(&secOutputPre, "\tvar _ %s\n", v.InType()) // this is not needed, but here for debugging purposes
			fmt.Fprint(&secOutputConv, v.Conversion())
		}
	}

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

func NewCallbackGenerator(ctx gencontext.GenerationContext, cb gir.Callback) *CallbackGenerator {
	if !cb.IsIntrospectable() {
		return nil
	}

	if strings.HasSuffix(cb.Name, "DestroyNotify") {
		return nil
	}

	callbackMeta := ctx.LookupType(cb.Name, "")

	if callbackMeta == nil {
		return nil
	}

	if cb.Parameters == nil {
		return nil // we cannot call a go closure without more info
	}

	closureArg, ok := findClosureArg(cb.Parameters.Parameters)

	if !ok {
		return nil // we cannot call a go closure without user data
	}

	valueCount := len(cb.Parameters.Parameters)

	if cb.ReturnValue != nil {
		valueCount += 1
	}

	if cb.Parameters != nil && cb.Parameters.InstanceParameter != nil {
		panic("unimplemented instance param for callbacks")
	}

	var docParams []ParamDoc
	var docReturns []ParamDoc

	g := &CallbackGenerator{
		GoName:  callbackMeta.GoBaseType,
		CgoName: callbackMeta.CGoBaseType,

		CReturnType: "void",

		CGoReturn:     nil,
		Converters:    make(value.ConverterList, 0, valueCount),
		CGoParameters: make(value.ConverterList, 0, valueCount),
		GoParameters:  make(value.ConverterList, 0, valueCount),
		GoReturns:     make(value.ConverterList, 0, valueCount),
	}

	if cb.Parameters != nil {
		for i, param := range cb.Parameters.Parameters {

			ctype, isArray := value.AnyTypeC(param.AnyType)

			if isArray {
				// TODO: handle array types
				log.Printf("skipping callback %s (%s) because of array param", g.GoName, g.CgoName)
				return nil
			}

			cname := fmt.Sprintf("arg%d", i+1)
			goname := fmt.Sprintf("_%s", param.Name)

			conv := value.NewParamConverter(ctx, value.ConvertCToGo, param.ParameterAttrs, cname, goname)

			if conv == nil {
				// conversion not possible, skip callback
				return nil
			}

			g.CParamTypes = append(g.CParamTypes, ctype)

			g.CGoParameters = append(g.CGoParameters, conv)

			if i == closureArg {
				g.ClosureArg = conv.InIdentifier()
				continue // don't convert or call go func with this argument
			}

			g.Converters = append(g.Converters, conv)

			if conv.ConversionDirection() == value.ConvertGoToC {
				// the converter turned the direction around, this is now a go return value
				g.GoReturns = append(g.GoReturns, conv)
				docReturns = append(docReturns, ParamDocFromParameter(conv.InIdentifier(), param))
			} else {
				g.GoParameters = append(g.GoParameters, conv)
				docParams = append(docParams, ParamDocFromParameter(conv.OutIdentifier(), param))
			}
		}
	}

	if cb.ReturnValue != nil && !value.IsVoid(cb.ReturnValue.AnyType) {
		ctype, isArray := value.AnyTypeC(cb.ReturnValue.AnyType)

		if isArray {
			log.Printf("skipping %s (%s) because of array return\n", g.GoName, cb.Name)
			return nil // TODO: handle array return type
		}

		goname := fmt.Sprintf("_%s", firstToLower(ctype))

		if ctype == "gboolean" {
			goname = "ok"
		}

		conv := value.NewReturnConverter(ctx, value.ConvertGoToC, *cb.ReturnValue, "cret", goname)

		if conv == nil {
			// conversion not possible, skip callback
			return nil
		}

		g.CReturnType = cb.ReturnValue.AnyType.Type.CType

		g.Converters = append(g.Converters, conv)

		g.CGoReturn = conv
		g.GoReturns = append(g.GoReturns, conv)

		docReturns = append(docReturns, ParamDocFromReturn(conv.InIdentifier(), *cb.ReturnValue))
	}

	if cb.Throws {
		panic("throwing callback unimplemented")
	}

	g.Doc = NewCallableGoDocGenerator(g.GoName, cb, 0, docParams, docReturns)

	// TODO: perform common reorderings here, e.g. move context to front.

	return g
}

// findClosureArg returns the index of the closure argument, and whether it was found
func findClosureArg(params []gir.Parameter) (int, bool) {
	for _, param := range params {
		if param.Closure != nil {
			return *param.Closure, true
		}
	}

	return 0, false
}
