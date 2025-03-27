package generators

import (
	"fmt"
	"log"
	"strings"
	"text/tabwriter"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/generators/callable"
	"github.com/diamondburned/gotk4/gir/girgen/generators/convert"
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

// CallableGenerator generates the code for a Go->C call. We convert the parameters, do the cgo call
// and convert the returns back to go. Some of the go returns are actually "out" params in C though.
type CallableGenerator struct {
	Doc SubGenerator

	// CCallExpressions contains the call expressions in the order expected by the c function
	CCallExpressions callable.CallExpressionList

	// ParamConverters contains all the go->c converters needed for the c call
	ParamConverters convert.ConverterList

	// ReturnConverters contains all the c->go converters needed to return all values
	ReturnConverters convert.ConverterList

	Signature *typesystem.CallableSignature
}

// GenerateInterfaceSignature implements MethodGenerator.
func (m *CallableGenerator) GenerateInterfaceSignature(w file.CodeWriter) {
	m.Doc.Generate(w)

	fmt.Fprintln(w, m.GoInterfaceDeclaration())
}

// Generate implements Generator.
func (m *CallableGenerator) Generate(w *file.Writer) {
	m.Doc.Generate(w.Go())

	fmt.Fprintf(w.Go(), "%s {\n", m.GoSignature())
	w.Go().Indent()

	// TODO imports

	decls := tabwriter.NewWriter(w.Go(), 0, 0, 1, ' ', 0) // this vertically aligns the decls without formatting
	for i, param := range m.Signature.GoParameters {
		conv := m.ParamConverters[i]
		fmt.Fprintf(decls, "var\t%s\t%s\t// %s\n", param.CName, param.Type.CGoType(), conv.Metadata())
	}
	for i, ret := range m.Signature.GoReturns {
		conv := m.ReturnConverters[i]
		fmt.Fprintf(decls, "var\t%s\t%s\t// %s\n", ret.CName, ret.Type.CGoType(), conv.Metadata())
	}
	decls.Flush()

	fmt.Fprintln(w.Go())

	for _, c := range m.ParamConverters {
		c.Convert(w.Go())
	}

	fmt.Fprintln(w.Go())

	fmt.Fprintln(w.Go(), m.CGoCall())

	fmt.Fprintln(w.Go())

	for _, c := range m.ReturnConverters {
		c.Convert(w.Go())
	}

	if len(m.Signature.GoReturns) > 0 {
		fmt.Fprintf(w.Go(), "\nreturn %s\n", m.Signature.GoReturns.GoIdentifiers())
	}

	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n\n")
}

func NewCallableGenerator(f *typesystem.CallableSignature) *CallableGenerator {

	if f.Parameters.InstanceParam != nil && typesystem.IsForeignType(f.Parameters.InstanceParam.Type) {
		log.Panicf("received callable %s with foreign receiver %s", f.CIndentifier(), f.Parameters.InstanceParam.GoName)
	}

	gen := &CallableGenerator{
		Doc:              NewCallableGoDocGenerator(f),
		Signature:        f,
		ParamConverters:  nil,
		ReturnConverters: nil,
	}

	for _, param := range f.CParameters() {
		gen.CCallExpressions = append(gen.CCallExpressions, callable.CallExpression{Param: param})
	}

	for _, param := range f.GoParameters {
		gen.ParamConverters = append(gen.ParamConverters, convert.NewGoToCConverter(param))
	}

	for _, param := range f.GoReturns {
		gen.ReturnConverters = append(gen.ReturnConverters, convert.NewCToGoConverter(param))
	}

	return gen
}

// GoSignature returns a string of the go function signature.
func (m *CallableGenerator) GoSignature() string {
	var recv string
	if m.Signature.InstanceParam != nil {
		recv = fmt.Sprintf(" (%s %s)", m.Signature.InstanceParam.GoName, m.Signature.InstanceParam.Type.GoType())
	}

	var ret string
	if len(m.Signature.GoReturns) == 1 {
		ret = " " + m.Signature.GoReturns[0].Type.GoType()
	} else if len(m.Signature.GoReturns) > 1 {
		ret = " (" + m.Signature.GoReturns.GoTypes() + ")"
	}

	return fmt.Sprintf("func%s %s(%s)%s", recv, m.Signature.GoIndentifier(), m.Signature.GoParameters.GoDeclarations(), ret)
}

// GoInterfaceDeclaration returns a string of the go function signature needed for an interface declaration.
func (m *CallableGenerator) GoInterfaceDeclaration() string {
	var ret string
	if len(m.Signature.GoReturns) == 1 {
		ret = " " + m.Signature.GoReturns[0].Type.GoType()
	} else if len(m.Signature.GoReturns) > 1 {
		ret = " (" + m.Signature.GoReturns.GoTypes() + ")"
	}

	return fmt.Sprintf("%s(%s)%s", m.Signature.GoIndentifier(), m.Signature.GoParameters.GoTypes(), ret)
}

// GoSignature returns a string of the go function signature.
func (m *CallableGenerator) CGoCall() string {
	creturn := m.Signature.CGoReturn()
	var ret string
	if creturn != nil {
		ret = fmt.Sprintf("%s = ", creturn.CName)
	}

	return fmt.Sprintf("%s%s(%s)", ret, m.Signature.CGoIndentifier(), m.CCallExpressions.Call())
}
