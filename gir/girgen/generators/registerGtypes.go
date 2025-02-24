package generators

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type GType struct {
	TypeFunc string
	GoName   string
}

func (gt GType) VarName() string {
	return fmt.Sprintf("GType%s", gt.GoName)
}

func (gt GType) MarshalFuncName() string {
	return fmt.Sprintf("marshal%s", gt.GoName) // The marshal function is not generated here, see [MarshalGenerator]
}

type RegisterGTypesGenerator struct {
	FoundTypes []GType
}

func NewRegisterGTypeGenerator(
	ctx gencontext.GenerationContext,
	ns *gir.Namespace,
) *RegisterGTypesGenerator {
	// TODO: use gencontext lookup for GoNames

	gen := &RegisterGTypesGenerator{}

	for _, v := range ns.Enums {
		if v.GLibGetType == "" {
			continue
		}
		gen.FoundTypes = append(gen.FoundTypes, GType{
			TypeFunc: v.GLibGetType,
			GoName:   strcases.PascalToGo(v.Name),
		})
	}
	for _, v := range ns.Bitfields {
		if v.GLibGetType == "" {
			continue
		}
		gen.FoundTypes = append(gen.FoundTypes, GType{
			TypeFunc: v.GLibGetType,
			GoName:   strcases.PascalToGo(v.Name),
		})
	}
	for _, v := range ns.Interfaces {
		if v.GLibGetType == "" {
			continue
		}
		gen.FoundTypes = append(gen.FoundTypes, GType{
			TypeFunc: v.GLibGetType,
			GoName:   strcases.PascalToGo(v.Name),
		})
	}
	for _, v := range ns.Records {
		if v.GLibGetType == "" {
			continue
		}
		gen.FoundTypes = append(gen.FoundTypes, GType{
			TypeFunc: v.GLibGetType,
			GoName:   strcases.PascalToGo(v.Name),
		})
	}
	for _, v := range ns.Unions {
		if v.GLibGetType == "" {
			continue
		}
		gen.FoundTypes = append(gen.FoundTypes, GType{
			TypeFunc: v.GLibGetType,
			GoName:   strcases.PascalToGo(v.Name),
		})
	}

	return gen
}

func (g *RegisterGTypesGenerator) Generate(w *file.Writer) {
	w.GoImportCoreGlib()

	// Need this for g_value_get_boxed
	w.AddPackage("glib-2.0")
	w.CInclude("glib-object.h")

	fmt.Fprintln(w.Go(), "// GType values.")
	fmt.Fprintln(w.Go(), "var(")

	for _, t := range g.FoundTypes {
		// TODO: pad name correctly to auto align without formatting
		fmt.Fprintf(w.Go(), "\t%s=coreglib.Type(C.%s())\n", t.VarName(), t.TypeFunc)
	}

	fmt.Fprintln(w.Go(), ")")

	fmt.Fprintln(w.Go())
	fmt.Fprintln(w.Go(), "func init() {")
	fmt.Fprintln(w.Go(), "\tcoreglib.RegisterGValueMarshalers([]coreglib.TypeMarshaler{")

	for _, t := range g.FoundTypes {
		fmt.Fprintf(w.Go(), "\t\tcoreglib.TypeMarshaler{T: %s, F: %s},\n", t.VarName(), t.MarshalFuncName())
	}

	fmt.Fprintln(w.Go(), "\t})")
	fmt.Fprintln(w.Go(), "}")
}
