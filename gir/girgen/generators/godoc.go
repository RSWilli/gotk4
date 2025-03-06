package generators

import (
	"bufio"
	"bytes"
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/file"
)

// InfoFields contains common fields that a GIR schema type may contain.
type InfoFields struct {
	CName      string
	Attrs      *gir.InfoAttrs
	Elements   *gir.InfoElements
	Parameters *gir.Parameters
	Return     *gir.ReturnValue
}

// GetInfoFields gets the InfoFields from the given value.
func GetInfoFields(v interface{}) InfoFields {
	var inf InfoFields

	switch t := v.(type) {
	case gir.Alias:
		inf.CName = t.CType
		inf.Attrs = &t.InfoAttrs
		inf.Elements = &t.InfoElements
	case gir.Bitfield:
		inf.CName = t.CType
		inf.Attrs = &t.InfoAttrs
		inf.Elements = &t.InfoElements
	case gir.Class:
		inf.CName = t.CType
		inf.Attrs = &t.InfoAttrs
		inf.Elements = &t.InfoElements
	case gir.Interface:
		inf.CName = t.CType
		inf.Attrs = &t.InfoAttrs
		inf.Elements = &t.InfoElements
	case gir.Record:
		inf.CName = t.CType
		inf.Attrs = &t.InfoAttrs
		inf.Elements = &t.InfoElements
	case gir.Union:
		inf.CName = t.CType
		inf.Attrs = &t.InfoAttrs
		inf.Elements = &t.InfoElements
	case gir.Enum:
		inf.CName = t.CType
		inf.Attrs = &t.InfoAttrs
		inf.Elements = &t.InfoElements
	case gir.Member:
		inf.CName = t.CIdentifier
		inf.Attrs = &t.InfoAttrs
		inf.Elements = &t.InfoElements
	case gir.Constant:
		inf.CName = t.Name
		inf.Attrs = &t.InfoAttrs
		inf.Elements = &t.InfoElements
	case gir.Field:
		inf.CName = t.Name

	// Callables:
	case gir.Callback:
		inf.CName = t.CIdentifier
		inf.Attrs = &t.InfoAttrs
		inf.Elements = &t.InfoElements
	case gir.Function:
		inf.CName = t.CIdentifier
		inf.Attrs = &t.InfoAttrs
		inf.Elements = &t.InfoElements
	case gir.Method:
		inf.CName = t.CIdentifier
		inf.Attrs = &t.InfoAttrs
		inf.Elements = &t.InfoElements
	case gir.VirtualMethod:
		inf.CName = t.CIdentifier
		inf.Attrs = &t.InfoAttrs
		inf.Elements = &t.InfoElements
	default:
		panic(fmt.Sprintf("godoc generation received unsupported gir type %T", v))
	}

	return inf
}

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

func NewGoDocGenerator(goName string, girWithDoc any, indent int) *GoDocGenerator {
	info := GetInfoFields(girWithDoc)

	return &GoDocGenerator{
		Indentation: indent,
		DocString:   fmt.Sprintf("%s (%s) should be documented more", goName, info.CName),
	}
}

// ParamDoc is used to unify documentation of gir.Param and gir.ReturnValue, because we are converting out params to returns
type ParamDoc struct {
	Name     string
	Optional bool
	Doc      string
}

func ParamDocFromReturn(goName string, r gir.ReturnValue) ParamDoc {
	doc := ""

	if r.DocElements.Doc != nil {
		doc = r.DocElements.Doc.String
	}

	return ParamDoc{
		Name:     goName,
		Optional: false,
		Doc:      doc,
	}
}

func ParamDocFromThrow(goName string) ParamDoc {
	return ParamDoc{
		Name:     goName,
		Optional: true,
		Doc:      "an error",
	}
}

func ParamDocFromParameter(goName string, r gir.Parameter) ParamDoc {
	doc := ""

	if r.Doc != nil {
		doc = r.Doc.String
	}
	return ParamDoc{
		Name:     goName,
		Optional: r.Optional,
		Doc:      doc,
	}
}

func NewCallableGoDocGenerator(goName string, girWithDoc any, indent int, params []ParamDoc, returnValues []ParamDoc) *GoDocGenerator {
	g := NewGoDocGenerator(goName, girWithDoc, indent)

	var doc strings.Builder

	doc.WriteString(g.DocString)

	if len(params) > 0 {
		doc.WriteString("\n\nThe function takes the following parameters:\n\n")

		for _, param := range params {
			_ = param
			fmt.Fprintf(&doc, "\t- %s TODO\n", param.Name)
		}
	}

	if len(returnValues) > 0 {
		doc.WriteString("\n\nThe function returns the following values:\n\n")

		for _, rv := range returnValues {
			_ = rv
			fmt.Fprintf(&doc, "\t- %s TODO\n", rv.Name)
		}
	}

	g.DocString = doc.String()

	return g
}
