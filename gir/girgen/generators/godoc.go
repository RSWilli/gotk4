package generators

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type GoDocGenerator struct {
	DocString string
	GIRDoc    typesystem.Doc
}

func (docg *GoDocGenerator) Generate(w file.CodeWriter) {
	// scan the lines of the comment and prefix each line with "// "
	r := strings.NewReader(docg.DocString)
	scanner := bufio.NewScanner(r) // scan lines

	for scanner.Scan() {
		w.Write([]byte("// "))
		w.Write(scanner.Bytes())
		w.Write([]byte("\n"))
	}

	var zerodoc typesystem.Doc

	if docg.GIRDoc == zerodoc {
		return
	}

	w.Write([]byte("//\n"))

	// scan the lines of the comment and prefix each line with "//\t" to signify a quote/code example
	r = strings.NewReader(docg.GIRDoc.Doc)
	scanner = bufio.NewScanner(r) // scan lines

	for scanner.Scan() {
		w.Write([]byte("// "))
		w.Write(scanner.Bytes())
		w.Write([]byte("\n"))
	}

	if docg.GIRDoc.Deprecated {
		w.Write([]byte("//\n"))
		w.Write([]byte("// Deprecated: "))
		if docg.GIRDoc.DeprecatedVersion != "" {
			fmt.Fprintf(w, "(since %s) ", docg.GIRDoc.DeprecatedVersion)
		}

		r := strings.NewReader(docg.GIRDoc.DocDeprecated)
		scanner := bufio.NewScanner(r) // scan lines

		scanner.Scan() // initial deprecated line is already prefied

		fmt.Fprintf(w, "%s\n", scanner.Bytes())

		for scanner.Scan() {
			w.Write([]byte("// "))
			w.Write(scanner.Bytes())
			w.Write([]byte("\n"))
		}
	}
}

type DocumentedType interface {
	typesystem.Type
	typesystem.Documented
}

type DocumentedIdentifier interface {
	typesystem.Identifier
	typesystem.Documented
}

func NewSignalGoDocGenerator(sig *typesystem.Signal) *GoDocGenerator {
	var docstring string

	if sig.Action {
		docstring = fmt.Sprintf("%s emits the \"%s\" signal", sig.GoName, sig.Name)
	} else {
		docstring = fmt.Sprintf("%s connects the provided callback to the \"%s\" signal", sig.GoName, sig.Name)
	}

	return &GoDocGenerator{
		DocString: docstring,
		GIRDoc:    sig.Doc,
	}
}

func NewIdentifierGoDocGenerator(identifier DocumentedIdentifier) *GoDocGenerator {
	return &GoDocGenerator{
		DocString: fmt.Sprintf("%s wraps %s", identifier.GoIndentifier(), identifier.CIndentifier()),
		GIRDoc:    identifier.Documentation(),
	}
}

func NewTypeGoDocGenerator(typ DocumentedType) *GoDocGenerator {
	gotype := typ.GoType(0)

	switch t := typ.(type) {
	case *typesystem.Class:
		gotype = t.GoInterfaceName
	case *typesystem.Interface:
		gotype = t.GoInterfaceName
	}

	return &GoDocGenerator{
		DocString: fmt.Sprintf("%s wraps %s", gotype, typ.CType(0)),
		GIRDoc:    typ.Documentation(),
	}
}

func NewCallableGoDocGenerator(callable *typesystem.CallableSignature) *GoDocGenerator {
	g := NewIdentifierGoDocGenerator(callable)

	var doc strings.Builder

	doc.WriteString(g.DocString)

	if len(callable.GoParameters) > 0 {
		doc.WriteString("\n\nThe function takes the following parameters:\n\n")

		for _, param := range callable.GoParameters {
			if param.Skip || param.Implicit {
				continue
			}
			fmt.Fprintf(&doc, "\t- %s \n", paramDocListItem(param))
		}
	}

	if len(callable.GoReturns) > 0 {
		doc.WriteString("\nThe function returns the following values:\n\n")

		for _, rv := range callable.GoReturns {
			fmt.Fprintf(&doc, "\t- %s \n", paramDocListItem(rv))
		}
	}

	g.DocString = doc.String()

	return g
}

func paramDocListItem(p *typesystem.Param) string {
	docStr := fmt.Sprintf("%s %s", p.GoName, p.GoType())

	if p.Nullable {
		docStr += " (nullable)"
	}

	// if p.Optional { // out params may be optional, but we always use them as go returns
	// 	docStr = " (optional)"
	// }

	if p.Doc.Doc != "" {
		docStr += ": "
		docStr += p.Doc.Doc
	}

	return docStr
}
