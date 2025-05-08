package generators

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type SignalEmitGenerator struct {
	Doc SubGenerator

	*typesystem.Signal
}

// Generate implements MethodGenerator.
func (s *SignalEmitGenerator) Generate(w *file.Package) {
	s.Doc.Generate(w.Go())

	for _, p := range s.Parameters() {
		w.GoImportType(p.Type)
	}

	var ret string
	if s.Signal.GoReturn() != nil {
		ret = " " + s.Signal.GoReturn().GoType()
	}

	fmt.Fprintf(w.Go(), "func (o *%s) %s(%s)%s {\n", s.InstanceParam.Type.Type.GoType(0), s.GoName, s.Parameters().GoDeclarations(), ret)
	w.Go().Indent()
	returnCast := ""
	if ret != "" {
		fmt.Fprint(w.Go(), "return ")

		returnCast = fmt.Sprintf(".(%s)", s.Signal.GoReturn().GoType())
	}
	params := s.Parameters()
	if len(params) == 0 {
		fmt.Fprintf(w.Go(), "o.Emit(\"%s\")%s\n", s.Name, returnCast)
	} else {
		fmt.Fprintf(w.Go(), "o.Emit(\"%s\", %s)%s\n", s.Name, s.Parameters().GoIdentifiers(), returnCast)
	}
	w.Go().Unindent()
	fmt.Fprintln(w.Go(), "}")
	fmt.Fprintln(w.Go(), "")
}

// GenerateInterfaceSignature implements MethodGenerator.
func (s *SignalEmitGenerator) GenerateInterfaceSignature(w file.File) {
	s.Doc.Generate(w.Go())

	var ret string
	if s.Signal.GoReturn() != nil {
		ret = " " + s.Signal.GoReturn().GoType()
	}

	fmt.Fprintf(w.Go(), "%s(%s)%s\n", s.GoName, s.Parameters().GoTypes(), ret)
}

func NewSignalEmitGenerator(sig *typesystem.Signal) *SignalEmitGenerator {
	return &SignalEmitGenerator{
		Doc:    NewGoDocGenerator(sig),
		Signal: sig,
	}
}
