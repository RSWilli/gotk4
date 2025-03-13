package generators

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type GoDocGenerator struct {
	Indentation int

	DocString string
}

func (docg *GoDocGenerator) Generate(w *file.Writer) {
	// scan the lines of the comment and prefix each line with tabs and "// "
	r := strings.NewReader(docg.DocString)
	scanner := bufio.NewScanner(r) // scan lines

	for scanner.Scan() {
		w.Go().Write(bytes.Repeat([]byte{'\t'}, docg.Indentation))
		w.Go().WriteString("// ")
		w.Go().Write(scanner.Bytes())
		w.Go().WriteByte('\n')
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

func NewIdentifierGoDocGenerator(identifier DocumentedIdentifier, indent int) *GoDocGenerator {
	return &GoDocGenerator{
		Indentation: indent,
		DocString:   fmt.Sprintf("%s (%s) should be documented more", identifier.GoIndentifier(), identifier.CIndentifier()),
	}
}

func NewTypeGoDocGenerator(typ DocumentedType, indent int) *GoDocGenerator {
	// info := GetInfoFields(girWithDoc)

	return &GoDocGenerator{
		Indentation: indent,
		DocString:   fmt.Sprintf("%s (%s) should be documented more", typ.GoType(), typ.CType()),
	}
}

func NewCallableGoDocGenerator(callable *typesystem.CallableSignature) *GoDocGenerator {
	g := NewIdentifierGoDocGenerator(callable, 0)

	var doc strings.Builder

	doc.WriteString(g.DocString)

	if len(callable.GoParameters) > 0 {
		doc.WriteString("\n\nThe function takes the following parameters:\n\n")

		for _, param := range callable.GoParameters {
			_ = param
			fmt.Fprintf(&doc, "\t- %s TODO\n", param.GoName)
		}
	}

	if len(callable.GoReturns) > 0 {
		doc.WriteString("\n\nThe function returns the following values:\n\n")

		for _, rv := range callable.GoReturns {
			_ = rv
			fmt.Fprintf(&doc, "\t- %s TODO\n", rv.GoName)
		}
	}

	g.DocString = doc.String()

	return g
}
