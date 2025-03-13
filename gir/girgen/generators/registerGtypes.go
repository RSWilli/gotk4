package generators

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type RegisterGTypesGenerator struct {
	FoundTypes []typesystem.Type
}

func NewRegisterGTypeGenerator(
	ns *typesystem.Namespace,
) *RegisterGTypesGenerator {
	gen := &RegisterGTypesGenerator{}

	for _, v := range ns.Enums {
		if v.GLibGetType() == "" {
			continue
		}
		gen.FoundTypes = append(gen.FoundTypes, v)
	}
	for _, v := range ns.Bitfields {
		if v.GLibGetType() == "" {
			continue
		}
		gen.FoundTypes = append(gen.FoundTypes, v)
	}
	for _, v := range ns.Interfaces {
		if v.GLibGetType() == "" {
			continue
		}
		gen.FoundTypes = append(gen.FoundTypes, v)
	}
	for _, v := range ns.Records {
		if v.GLibGetType() == "" {
			continue
		}
		gen.FoundTypes = append(gen.FoundTypes, v)
	}
	for _, v := range ns.Unions {
		if v.GLibGetType() == "" {
			continue
		}
		gen.FoundTypes = append(gen.FoundTypes, v)
	}

	return gen
}

func (g *RegisterGTypesGenerator) Generate(w *file.Writer) {
	w.GoImportCoreGlib()

	// TODO: doing this in a central place means that we need a way to prevent registering a type if
	// the generator skips it. It would be better to do this decentralized. Maybe namespace can do this better, since we embed the Types
	// in the generators

	// Need this for g_value_get_boxed
	w.AddPackage("glib-2.0")
	w.CInclude("glib-object.h")

	fmt.Fprintln(w.Go(), "// GType values.")
	fmt.Fprintln(w.Go(), "var (")

	for _, t := range g.FoundTypes {
		// TODO: pad name correctly to auto align without formatting
		fmt.Fprintf(w.Go(), "\tGType%s=coreglib.Type(C.%s())\n", t.GoType(), t.GLibGetType())
	}

	fmt.Fprintln(w.Go(), ")")

	fmt.Fprintln(w.Go())
	fmt.Fprintln(w.Go(), "func init() {")
	fmt.Fprintln(w.Go(), "\tcoreglib.RegisterGValueMarshalers([]coreglib.TypeMarshaler{")

	for _, t := range g.FoundTypes {
		fmt.Fprintf(w.Go(), "\t\tcoreglib.TypeMarshaler{T: GType%s, F: %s},\n", t.GoType(), t.MarshalFuncName())
	}

	fmt.Fprintln(w.Go(), "\t})")
	fmt.Fprintln(w.Go(), "}")
}
