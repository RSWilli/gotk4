package generators

import (
	"fmt"
	"strconv"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/gotmpl"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
	"github.com/diamondburned/gotk4/gir/girgen/types"
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
	Doc       Generator
	Name      string
	ShortName string
	Value     string
}

func bits(v string) string {
	b, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return v
	}

	return "0b" + strconv.FormatUint(b, 2)
}

type BitfieldGenerator struct {
	Doc  Generator
	Name string

	MethodReceiver  string
	StringedMembers []BitfieldMember // only unique members are considered, to avoid switch case clash
	MaxStringLen    int

	Marshaler Generator

	Members []BitfieldMember
}

func (g *BitfieldGenerator) Generate(w *file.Writer) {
	g.Doc.Generate(w)
	w.GoImport("fmt")
	w.GoImport("strings")

	fmt.Fprintf(w.Go(), "type %s C.guint\n\n", g.Name)

	fmt.Fprintln(w.Go(), "const (")
	for _, m := range g.Members {
		m.Doc.Generate(w)
		fmt.Fprintf(w.Go(), "\t%s %s = %s\n", m.Name, g.Name, m.Value)
	}
	fmt.Fprint(w.Go(), ")\n\n")

	g.Marshaler.Generate(w)

	fmt.Fprintln(w.Go(), "// String returns the names in string for AllocatorFlags.")
	fmt.Fprintf(w.Go(), "func (%s %s) String() string {\n", g.MethodReceiver, g.Name)
	fmt.Fprintf(w.Go(), "\tif %s == 0 {\n", g.MethodReceiver)
	fmt.Fprintf(w.Go(), "\t\treturn \"%s(0)\"\n", g.Name)
	fmt.Fprintf(w.Go(), "\t}\n\n")

	fmt.Fprintf(w.Go(), "\tvar builder strings.Builder\n")
	fmt.Fprintf(w.Go(), "\tbuilder.Grow(%d)\n\n", g.MaxStringLen)

	fmt.Fprintf(w.Go(), "\tfor %s != 0 {\n", g.MethodReceiver)
	fmt.Fprintf(w.Go(), "\t\tnext := %s & (%s - 1)\n", g.MethodReceiver, g.MethodReceiver)
	fmt.Fprintf(w.Go(), "\t\tbit := %s - next\n\n", g.MethodReceiver)

	fmt.Fprintf(w.Go(), "\t\tswitch bit {\n")
	for _, m := range g.StringedMembers {
		fmt.Fprintf(w.Go(), "\t\tcase %s:\n", m.Name)
		fmt.Fprintf(w.Go(), "\t\t\tbuilder.WriteString(\"%s|\")\n", m.ShortName)
	}
	fmt.Fprintf(w.Go(), "\t\tdefault:\n")
	fmt.Fprintf(w.Go(), "\t\t\tbuilder.WriteString(fmt.Sprintf(\"%s(0b%%b)|\", bit))\n", g.Name)
	fmt.Fprintf(w.Go(), "\t\t}\n\n")

	fmt.Fprintf(w.Go(), "\t\t%s = next\n", g.MethodReceiver)
	fmt.Fprintf(w.Go(), "\t}\n\n")

	fmt.Fprintf(w.Go(), "\treturn strings.TrimSuffix(builder.String(), \"|\")\n")
	fmt.Fprintf(w.Go(), "}\n\n")

	fmt.Fprintf(w.Go(), "// Has returns true if %s contains other\n", g.MethodReceiver)
	fmt.Fprintf(w.Go(), "func (%s %s) Has(other %s) bool {\n", g.MethodReceiver, g.Name, g.Name)
	fmt.Fprintf(w.Go(), "\treturn (%s & other) == other\n", g.MethodReceiver)
	fmt.Fprintf(w.Go(), "}\n\n")
}

func NewBitfieldGenerator(b gir.Bitfield) *BitfieldGenerator {
	// TODO: use gencontext lookup

	var members []BitfieldMember

	maxStrLen := 0

	for i, m := range b.Members {
		goName := formatEnumMember(m)
		mm := BitfieldMember{
			Doc:       NewGoDocGenerator(goName, m, 1),
			Name:      goName,
			ShortName: strcases.SnakeToGo(true, m.Name()),
			Value:     bits(m.Value),
		}
		members = append(members, mm)

		maxStrLen += len(mm.Name)
		if i > 0 {
			maxStrLen++ // account for '|'
		}
	}

	stringerMembers := dedupValues(members)

	goName := strcases.PascalToGo(b.Name)

	var marshalGen Generator = NoopGenerator{}

	if b.GLibGetType != "" {
		marshalGen = NewMarshalBifieldGenerator(goName)
	}

	return &BitfieldGenerator{
		Doc:             NewGoDocGenerator(goName, b, 0),
		Name:            goName,
		Members:         members,
		MethodReceiver:  strcases.FirstLetter(goName),
		Marshaler:       marshalGen,
		MaxStringLen:    maxStrLen,
		StringedMembers: stringerMembers,
	}
}

func dedupValues(members []BitfieldMember) []BitfieldMember {
	seen := make(map[string]struct{})
	uniques := make([]BitfieldMember, 0, len(members))

	for _, m := range members {
		if _, ok := seen[m.Value]; ok {
			continue
		}

		uniques = append(uniques, m)
		seen[m.Value] = struct{}{}
	}

	return uniques
}
