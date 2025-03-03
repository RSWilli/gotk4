package generators

import (
	"fmt"
	"log"
	"strings"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/generators/value"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type RecordConstructorGenerator struct {
	GoName string

	Params value.ConverterList
	Return value.Converter

	FuncBody Generator
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

	r.FuncBody.Generate(w)

	fmt.Fprintf(w.Go(), "\n\treturn %s\n", r.Return.OutIdentifier())
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

	goRet := fmt.Sprintf("_%s", firstToLower(parent.GoName))

	conv := value.NewReturnConverter(ctx, value.ConvertCToGo, *c.ReturnValue, "_cret", goRet)

	if conv == nil {
		return nil // cannot convert
	}

	g := &RecordConstructorGenerator{
		GoName: goConstructorName(c.Name, parent.GoName),

		Return: conv,
	}

	if c.Parameters != nil {
		for i, param := range c.Parameters.Parameters {

			cArg := fmt.Sprintf("_arg%d", i+1)

			conv := value.NewParamConverter(ctx, value.ConvertGoToC, param.ParameterAttrs, cArg, DodgeReservedFieldName(param.Name))

			if conv == nil {
				return nil
			}

			if conv.ConversionDirection() == value.ConvertCToGo {
				log.Printf("skipping constructor %s (%s) because of out param\n", g.GoName, c.CIdentifier)
				return nil
			}

			g.Params = append(g.Params, conv)
		}
	}

	g.FuncBody = NewCallableBodyGenerator("C."+c.CIdentifier, g.Params, g.Return)

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
