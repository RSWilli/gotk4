package value

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
)

func newUserDataParamConverter(ctx gencontext.GenerationContext, goCallBackTypeName string, conversionValueIndex ConversionValueIndex, param gir.Parameter) Converter {
	return NoopConverter{}
}

// fmt.Fprintf(w.Exported.Go(), "var fn %s\n", c.GoName)
// fmt.Fprintf(w.Exported.Go(), "{\n")
// fmt.Fprintf(w.Exported.Go(), "\tv := gbox.Get(uintptr(%s))\n", userDataArg)
// fmt.Fprintf(w.Exported.Go(), "\tif v == nil {")
// fmt.Fprintf(w.Exported.Go(), "\t\tpanic(`callback not found`)\n")
// fmt.Fprintf(w.Exported.Go(), "\t}\n")
// fmt.Fprintf(w.Exported.Go(), "\tfn = v.(%s)\n", c.GoName)
// fmt.Fprintf(w.Exported.Go(), "}\n\n")
