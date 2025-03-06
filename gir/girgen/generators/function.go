package generators

import (
	"fmt"
	"log"
	"strings"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/generators/callable"
	"github.com/diamondburned/gotk4/gir/girgen/generators/value"
	"github.com/diamondburned/gotk4/gir/girgen/gotmpl"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
	"github.com/diamondburned/gotk4/gir/girgen/types"
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
	Doc    Generator
	GoName string

	FuncBody Generator

	// GoParameters contains the converters as Converters used to generate the go parameters. Compared to CGoParameters there might
	// be some parameters missing, e.g. out params get turned into return values in go. The girgen user is free to reorder this list.
	//
	// common reorderings are already applied, e.g. moving context to the front
	GoParameters value.ConverterList

	// GoReturns contains the same converters as Converters, but is used to generate the go return values.
	GoReturns value.ConverterList
}

// Generate implements Generator.
func (m *FunctionGenerator) Generate(w *file.Writer) {
	m.Doc.Generate(w)

	var ret string

	if len(m.GoReturns) == 1 {
		ret = " " + m.GoReturns[0].OutType()
	} else if len(m.GoReturns) > 1 {
		ret = "(" + m.GoReturns.OutTypeList() + ")"
	}

	fmt.Fprintf(w.Go(), "func %s(%s)%s {\n", m.GoName, m.GoParameters.GoDeclList(), ret)
	m.FuncBody.Generate(w)

	if len(m.GoReturns) > 0 {
		fmt.Fprintf(w.Go(), "\n\treturn %s\n", m.GoReturns.OutIdentifierList())
	}

	fmt.Fprintf(w.Go(), "}\n\n")
}

func NewFunctionGenerator(ctx gencontext.GenerationContext, f gir.Function) *FunctionGenerator {
	if !f.IsIntrospectable() || f.ShadowedBy != "" || f.MovedTo != "" {
		return nil
	}

	if f.Parameters != nil && f.Parameters.InstanceParameter != nil {
		panic("function with instance parameter not implemented")
	}

	valueCount := 0

	if f.Parameters != nil {
		valueCount += len(f.Parameters.Parameters)
	}

	if f.ReturnValue != nil {
		valueCount += 1
	}

	goName := strcases.SnakeToGo(true, f.Name)

	var docParams []ParamDoc
	var docReturns []ParamDoc

	var cGoReturn value.Converter = nil
	var cGoParameters = make(value.ConverterList, 0, valueCount)

	g := &FunctionGenerator{
		GoName: goName,

		GoParameters: make(value.ConverterList, 0, valueCount),
		GoReturns:    make(value.ConverterList, 0, valueCount),
	}

	if f.Parameters != nil {
		for i, param := range f.Parameters.Parameters {
			cArg := fmt.Sprintf("_arg%d", i+1)

			conv := value.NewParamConverter(ctx, value.ConvertGoToC, param.ParameterAttrs, cArg, DodgeReservedFieldName(param.Name))

			if conv == nil {
				// conversion not possible, skip method
				return nil
			}

			cGoParameters = append(cGoParameters, conv)

			if conv.ConversionDirection() == value.ConvertCToGo {
				// the converter turned the direction around, this is now a go return value
				g.GoReturns = append(g.GoReturns, conv)
				docReturns = append(docReturns, ParamDocFromParameter(conv.OutIdentifier(), param))
			} else {
				g.GoParameters = append(g.GoParameters, conv)
				docParams = append(docParams, ParamDocFromParameter(conv.InIdentifier(), param))
			}
		}
	}

	if f.ReturnValue != nil && !value.IsVoid(f.ReturnValue.AnyType) {
		cname, isArray := value.AnyTypeName(f.ReturnValue.AnyType)

		if isArray {
			log.Printf("skipping record method %s (%s) because of array return\n", g.GoName, f.Name)

			return nil
		}

		goRet := fmt.Sprintf("_%s", firstToLower(cname))

		conv := value.NewReturnConverter(ctx, value.ConvertCToGo, *f.ReturnValue, "_cret", goRet)

		if conv == nil {
			// conversion not possible, skip method
			return nil
		}

		cGoReturn = conv
		g.GoReturns = append(g.GoReturns, conv)
		docReturns = append(docReturns, ParamDocFromReturn(conv.OutIdentifier(), *f.ReturnValue))
	}

	if f.Throws {
		conv := value.NewThrowConverter(ctx, value.ConvertCToGo, "_cerr", "_goerr")

		if conv == nil {
			return nil
		}

		g.GoReturns = append(g.GoReturns, conv)
		cGoParameters = append(cGoParameters, conv)
		docReturns = append(docReturns, ParamDocFromThrow("error"))
	}

	g.Doc = NewCallableGoDocGenerator(goName, f, 0, docParams, docReturns)

	g.FuncBody = NewCallableBodyGenerator("C."+f.CIdentifier, cGoParameters, cGoReturn)

	// TODO: perform common reorderings here, e.g. move context to front.

	return g
}
