package generators

import (
	"fmt"
	"log"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/generators/value"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type RecordMethodGenerator struct {
	Doc    Generator
	GoName string

	parent *RecordGenerator

	FuncBody Generator

	// GoReceiver contains the go receiver param for the function call. It is also used as the first CGo Parameter.
	GoReceiver value.Converter

	// GoParameters contains the converters as Converters used to generate the go parameters. Compared to CGoParameters there might
	// be some parameters missing, e.g. out params get turned into return values in go. The girgen user is free to reorder this list.
	//
	// common reorderings are already applied, e.g. moving context to the front
	GoParameters value.ConverterList

	// GoReturns contains the same converters as Converters, but is used to generate the go return values.
	GoReturns value.ConverterList
}

// Generate implements Generator.
func (m *RecordMethodGenerator) Generate(w *file.Writer) {
	m.Doc.Generate(w)

	var ret string

	if len(m.GoReturns) == 1 {
		ret = " " + m.GoReturns[0].OutType()
	} else if len(m.GoReturns) > 1 {
		ret = "(" + m.GoReturns.OutTypeList() + ")"
	}

	fmt.Fprintf(w.Go(), "func (%s *%s) %s(%s)%s {\n", m.GoReceiver.InIdentifier(), m.parent.GoName, m.GoName, m.GoParameters.GoDeclList(), ret)
	m.FuncBody.Generate(w)

	if len(m.GoReturns) > 0 {
		fmt.Fprintf(w.Go(), "\n\treturn %s\n", m.GoReturns.OutIdentifierList())
	}

	fmt.Fprintf(w.Go(), "}\n\n")
}

func NewRecordMethodGenerator(ctx gencontext.GenerationContext, parent *RecordGenerator, m gir.Method) *RecordMethodGenerator {
	if !m.IsIntrospectable() {
		return nil
	}

	if m.Parameters == nil || m.Parameters.InstanceParameter == nil {
		panic("method parameters is nil, need at least instance parameter")
	}

	valueCount := len(m.Parameters.Parameters) + 1 // params + instance param

	if m.ReturnValue != nil {
		valueCount += 1
	}

	goName := strcases.SnakeToGo(true, m.Name)

	var docParams []ParamDoc
	var docReturns []ParamDoc

	var cGoReturn value.Converter = nil
	var cGoParameters = make(value.ConverterList, 0, valueCount)

	g := &RecordMethodGenerator{
		GoName: goName,

		parent: parent,

		GoParameters: make(value.ConverterList, 0, valueCount),
		GoReturns:    make(value.ConverterList, 0, valueCount),
	}

	instanceConv := value.NewInstanceParamConverter(ctx, value.ConvertGoToC, *m.Parameters.InstanceParameter, "_arg0", parent.ReceiverName)

	if instanceConv == nil {
		return nil
	}

	if instanceConv.ConversionDirection() == value.ConvertCToGo {
		panic("cannot handle out instance param")
	}

	cGoParameters = append(cGoParameters, instanceConv)
	g.GoReceiver = instanceConv

	if m.Parameters != nil {
		for i, param := range m.Parameters.Parameters {
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

	if m.ReturnValue != nil && !value.IsVoid(m.ReturnValue.AnyType) {
		cname, isArray := value.AnyTypeName(m.ReturnValue.AnyType)

		if isArray {
			log.Printf("skipping record method %s (%s) because of array return\n", g.GoName, m.Name)

			return nil
		}

		goRet := fmt.Sprintf("_%s", firstToLower(cname))

		conv := value.NewReturnConverter(ctx, value.ConvertCToGo, *m.ReturnValue, "_cret", goRet)

		if conv == nil {
			// conversion not possible, skip method
			return nil
		}

		cGoReturn = conv
		g.GoReturns = append(g.GoReturns, conv)
		docReturns = append(docReturns, ParamDocFromReturn(conv.OutIdentifier(), *m.ReturnValue))
	}

	if m.Throws {
		panic("throwing method unimplemented")
	}

	g.Doc = NewCallableGoDocGenerator(goName, m, 0, docParams, docReturns)

	g.FuncBody = NewCallableBodyGenerator("C."+m.CIdentifier, cGoParameters, cGoReturn)

	// TODO: perform common reorderings here, e.g. move context to front.

	return g
}
