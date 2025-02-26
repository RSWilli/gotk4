package generators

import (
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/generators/value"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type RecordConstructorGenerator struct {
	GoName string
	CName  string

	Params value.ConverterList
	Return value.Converter
}

// Generate implements Generator.
func (r *RecordConstructorGenerator) Generate(w *file.Writer) {
	if len(r.Params) > 0 {
		w.GoImport("runtime")
	}

	r.Params.AddImports(w)
	r.Return.AddImports(w)

	fmt.Fprintf(w.Go(), "// %s constructs a struct %s.\n", r.GoName, r.Return.OutType()) // TODO: use godoc generator instead
	fmt.Fprintf(w.Go(), "func %s(%s) %s {\n", r.GoName, r.Params.GoDeclList(), r.Return.OutType())

	for _, param := range r.Params {
		fmt.Fprintf(w.Go(), "\tvar %s %s // out\n", param.OutIdentifier(), param.OutType())
	}

	fmt.Fprintf(w.Go(), "\tvar %s %s // in\n", r.Return.InIdentifier(), r.Return.InType())

	fmt.Fprintln(w.Go())

	for _, param := range r.Params {
		fmt.Fprintln(w.Go(), param.Conversion())
	}

	// func call:
	fmt.Fprintf(w.Go(), "\t%s = %s(%s)\n", r.Return.InIdentifier(), r.CName, r.Params.OutIdentifierList())

	for _, param := range r.Params {
		fmt.Fprintf(w.Go(), "\truntime.KeepAlive(%s)\n", param.InIdentifier())
	}

	fmt.Fprintln(w.Go())

	fmt.Fprintf(w.Go(), "\tvar %s %s // out\n\n", r.Return.OutIdentifier(), r.Return.OutType())

	fmt.Fprintln(w.Go(), r.Return.Conversion())

	fmt.Fprintf(w.Go(), "\treturn %s\n", r.Return.OutIdentifier())
	fmt.Fprintf(w.Go(), "}\n\n")
}

func NewRecordConstructorGenerator(ctx gencontext.GenerationContext, parent *RecordGenerator, c gir.Constructor) *RecordConstructorGenerator {
	if !c.IsIntrospectable() {
		return nil
	}

	if c.ReturnValue == nil {
		// a constructor without a return value???
		return nil
	}

	conv := value.NewReturnConverter(ctx, value.ConvertCToGo, *c.ReturnValue)

	g := &RecordConstructorGenerator{
		GoName: goConstructorName(c.Name, parent.GoName),
		CName:  "C." + c.CIdentifier,

		Return: conv,
	}

	if c.Parameters != nil {
		for i, param := range c.Parameters.Parameters {
			conv := value.NewParamConverter(ctx, value.ConvertGoToC, i, param)

			if conv == nil {
				return nil
			}

			if conv.ConversionDirection() == value.ConvertCToGo {
				panic("unhandled out constructor param")
			}

			g.Params = append(g.Params, conv)
		}
	}

	return g
}

// goConstructorName turns the girname into a hopefully unique function name
//
// e.g. BufferList.new_sized -> NewBufferListSized
func goConstructorName(girname string, recordName string) string {
	pascal := strcases.SnakeToGo(true, girname)

	noNew := strings.TrimPrefix(pascal, "New")

	return "New" + recordName + noNew
}
