package generators

import (
	"fmt"
	"strconv"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/gotmpl"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
	"github.com/diamondburned/gotk4/gir/girgen/types"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

// Deprecated: old
var bitfieldTmpl = gotmpl.NewGoTemplate(`
	{{ GoDoc . 0 (OverrideSelfName .GoName) }}
	type {{ .GoName }} C.guint

	const (
		{{ range .Members -}}
		{{- $name := ($.FormatMember .) -}}
		{{- GoDoc . 1 TrailingNewLine (OverrideSelfName $name) -}}
		{{- $name }} {{ $.GoName }} = {{ $.Bits .Value }}
		{{ end -}}
	)

	{{ if .Marshaler }}
	func marshal{{.GoName}}(p uintptr) (interface{}, error) {
		return {{ .GoName }}(coreglib.ValueFromNative(unsafe.Pointer(p)).Flags()), nil
	}
	{{ end }}

	{{ $recv := FirstLetter .GoName }}
	// String returns the names in string for {{.GoName}}.
	func ({{$recv}} {{.GoName}}) String() string {
		if {{$recv}} == 0 {
			return "{{.GoName}}(0)"
		}

		var builder strings.Builder
		builder.Grow({{.StrLen}})

		for {{$recv}} != 0 {
			next := {{$recv}} & ({{$recv}} - 1)
			bit := {{$recv}} - next

			switch bit {
			{{- range .UniqueMembers }} {{ $name := $.FormatMember . }}
			case {{$name}}: builder.WriteString("{{SnakeToGo true .Name}}|")
			{{- end }}
			default: builder.WriteString(fmt.Sprintf("{{.GoName}}(0b%b)|", bit))
			}

			{{$recv}} = next
		}

		return strings.TrimSuffix(builder.String(), "|")
	}

	// Has returns true if {{$recv}} contains other.
	func ({{$recv}} {{.GoName}}) Has(other {{.GoName}}) bool {
		return ({{$recv}} & other) == other
	}
`)

// Deprecated: old
type bitfieldData struct {
	*gir.Bitfield
	GoName    string
	StrLen    int // length of all enum strings concatenated
	Marshaler bool
	Members   []gir.Member

	gen FileGenerator
}

// Deprecated: old
func (*bitfieldData) Bits(v string) string {
	b, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return v
	}

	return "0b" + strconv.FormatUint(b, 2)
}

// Deprecated: old
func (b *bitfieldData) FormatMember(member gir.Member) string {
	return formatEnumMember(member)
}

// Deprecated: old
func (b *bitfieldData) UniqueMembers() []gir.Member {
	return uniqueValuesEnumMembers(b.Members)
}

// CanGenerateBitfield returns false if the bitfield cannot be generated.
//
// Deprecated: old
func CanGenerateBitfield(gen FileGenerator, bitfield *gir.Bitfield) bool {
	if !bitfield.IsIntrospectable() || types.Filter(gen, bitfield.Name, bitfield.CType) {
		return false
	}
	return true
}

// GenerateBitfield generates a bitfield type declaration as well as the
// constants and the type marshaler into the given file generator. If the
// generation fails or is ignored, then false is returned.
//
// Deprecated: old
func GenerateBitfield(gen FileGeneratorWriter, bitfield *gir.Bitfield) bool {
	if !CanGenerateBitfield(gen, bitfield) {
		return false
	}

	goName := strcases.PascalToGo(bitfield.Name)
	writer := FileWriterFromType(gen, bitfield)

	data := bitfieldData{
		Bitfield: bitfield,
		GoName:   goName,
		Members:  make([]gir.Member, 0, len(bitfield.Members)),
		gen:      gen,
	}

	if gtype, ok := GenerateGType(gen, bitfield.Name, bitfield.GLibGetType); ok {
		data.Marshaler = true
		writer.Header().AddMarshaler(gtype.GetType, goName)
	}

	if bitfield.GLibGetType != "" && !types.FilterCType(gen, bitfield.GLibGetType) {
	}

	// Need this for String().
	writer.Header().Import("strings")
	writer.Header().Import("fmt")

	for i, member := range bitfield.Members {
		if v, _ := strconv.ParseInt(member.Value, 10, 64); v < 0 {
			// Ignore negative values that are bodged by GIR.
			continue
		}

		data.Members = append(data.Members, member)
		data.StrLen += len(data.FormatMember(member))
		if i > 0 {
			data.StrLen++ // account for '|'
		}
	}

	// Cap the StrLen.
	if data.StrLen > 256 {
		data.StrLen = 256
	}

	writer.Pen().WriteTmpl(bitfieldTmpl, &data)
	return true
}

type BitfieldMember struct {
	Doc SubGenerator

	*typesystem.Member
}

func bits(v string) string {
	b, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return v
	}

	return "0b" + strconv.FormatUint(b, 2)
}

type BitfieldGenerator struct {
	Doc SubGenerator

	*typesystem.Bitfield

	MethodReceiver string

	Marshaler Generator

	Members []BitfieldMember
}

func (g *BitfieldGenerator) Generate(w *file.Package) {
	g.Doc.Generate(w.Go())

	fmt.Fprintf(w.Go(), "type %s C.gint\n\n", g.GoType(0))

	fmt.Fprintln(w.Go(), "const (")
	w.Go().Indent()
	for _, m := range g.Members {
		m.Doc.Generate(w.Go())
		fmt.Fprintf(w.Go(), "%s %s = %s\n", m.GoIndentifier(), g.GoType(0), m.Value)
	}
	w.Go().Unindent()
	fmt.Fprint(w.Go(), ")\n\n")

	if g.Marshaler != nil {
		w.RegisterGType(g)
		g.Marshaler.Generate(w)
	}

	fmt.Fprintf(w.Go(), "// Has returns true if %s contains other\n", g.MethodReceiver)
	fmt.Fprintf(w.Go(), "func (%s %s) Has(other %s) bool {\n", g.MethodReceiver, g.GoType(0), g.GoType(0))
	fmt.Fprintf(w.Go(), "\treturn (%s & other) == other\n", g.MethodReceiver)
	fmt.Fprintf(w.Go(), "}\n\n")
}

func NewBitfieldGenerator(bf *typesystem.Bitfield) *BitfieldGenerator {
	var members []BitfieldMember

	for _, m := range bf.Members {
		mm := BitfieldMember{
			Doc:    NewIdentifierGoDocGenerator(m),
			Member: m,
		}
		members = append(members, mm)
	}

	var marshalGen Generator

	if bf.GLibGetType() != "" {
		marshalGen = NewMarshalBifieldGenerator(bf)
	}

	return &BitfieldGenerator{
		Doc:      NewTypeGoDocGenerator(bf),
		Bitfield: bf,

		Members:        members,
		MethodReceiver: strcases.ReceiverName(bf.GoType(0)),
		Marshaler:      marshalGen,
	}
}
