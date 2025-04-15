package generators

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/gotmpl"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
	"github.com/diamondburned/gotk4/gir/girgen/types"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

var enumTmpl = gotmpl.NewGoTemplate(`
	{{ GoDoc . 0 (OverrideSelfName .GoName) }}
	type {{ .GoName }} C.gint

	{{ if .IsIota }}
	const (
		{{ range $ix, $member := .Members }}
		{{- $name := $.FormatMember $member -}}
		{{- GoDoc . 1 TrailingNewLine (OverrideSelfName $name) -}}
		{{- if (eq $ix 0) }}
		{{- $name }} {{ $.GoName }} = iota
		{{- else }}
		{{- $name }}
		{{- end }}
		{{ end }}
	)
	{{ else }}
	const (
		{{ range .Members -}}
		{{- $name := $.FormatMember . -}}
		{{- GoDoc . 1 TrailingNewLine (OverrideSelfName $name) -}}
		{{- $name }} {{ $.GoName }} = {{ .Value }}
		{{ end -}}
	)
	{{ end }}

	{{ if .Marshaler }}
	func marshal{{ .GoName }}(p uintptr) (interface{}, error) {
		return {{ .GoName }}(coreglib.ValueFromNative(unsafe.Pointer(p)).Enum()), nil
	}
	{{ end }}

	{{ $recv := FirstLetter .GoName }}
	// String returns the name in string for {{ .GoName }}.
	func ({{ $recv }} {{ .GoName }}) String() string {
		switch {{ $recv }} {
		{{- range .UniqueMembers }} {{ $name := $.FormatMember . }}
		case {{ $name }}: return "{{ SnakeToGo true .Name }}"
		{{- end }}
		default: return fmt.Sprintf("{{ .GoName }}(%d)", {{ $recv }})
		}
	}
`)

// Deprecated: old
type enumData struct {
	*gir.Enum
	GoName    string
	IsIota    bool
	Marshaler bool
}

var numberMap = map[rune]string{
	'0': "Zero",
	'1': "One",
	'2': "Two",
	'3': "Three",
	'4': "Four",
	'5': "Five",
	'6': "Six",
	'7': "Seven",
	'8': "Eight",
	'9': "Nine",
}

// Deprecated: old
func (eg *enumData) FormatMember(member gir.Member) string {
	return formatEnumMember(member)
}

// Deprecated: old
func (eg *enumData) UniqueMembers() []gir.Member {
	return uniqueValuesEnumMembers(eg.Members)
}

// formatEnumMember returns the enum member's Go name.
func formatEnumMember(member gir.Member) string {
	// Pop the namespace off. Probably works most of the time.
	if parts := strings.SplitN(member.CIdentifier, "_", 2); len(parts) == 2 {
		member.CIdentifier = parts[1]
	}

	memberName := strcases.SnakeToGo(true, strings.ToLower(member.CIdentifier))

	// TODO: prepend GoName instead.
	r, sz := utf8.DecodeRuneInString(memberName)
	if sz > 0 && unicode.IsNumber(r) {
		memberName = numberMap[r] + memberName[sz:]
	}

	return memberName
}

// uniqueValuesEnumMembers returns the enum members with unique values only.
func uniqueValuesEnumMembers(members []gir.Member) []gir.Member {
	uniques := make([]gir.Member, 0, len(members))
	known := make(map[string]struct{}, len(members))

	for _, member := range members {
		_, isKnown := known[member.Value]
		if isKnown {
			continue
		}

		uniques = append(uniques, member)
		known[member.Value] = struct{}{}
	}

	return uniques
}

// CanGenerateEnum returns false if the given enum cannot be generated.
//
// Deprecated: old
func CanGenerateEnum(gen FileGenerator, enum *gir.Enum) bool {
	if !enum.IsIntrospectable() || types.Filter(gen, enum.Name, enum.CType) {
		return false
	}
	return true
}

// GenerateEnum generates an enum type declaration as well as the constants and
// the type marshaler.
//
// Deprecated: old
func GenerateEnum(gen FileGeneratorWriter, enum *gir.Enum) bool {
	if !CanGenerateEnum(gen, enum) {
		return false
	}

	goName := strcases.PascalToGo(enum.Name)
	writer := FileWriterFromType(gen, enum)

	data := enumData{
		Enum:   enum,
		GoName: goName,
		IsIota: true,
	}

	if gtype, ok := GenerateGType(gen, enum.Name, enum.GLibGetType); ok {
		data.Marshaler = true
		writer.Header().AddMarshaler(gtype.GetType, goName)
	}

	for i := 0; i < len(enum.Members); i++ {
		if enum.Members[i].Value != strconv.Itoa(i) {
			data.IsIota = false
			break
		}
	}

	writer.Header().Import("fmt")
	writer.Pen().WriteTmpl(enumTmpl, &data)
	return true
}

type EnumMember struct {
	Doc SubGenerator

	*typesystem.Member
}

type EnumGenerator struct {
	Doc SubGenerator

	*typesystem.Enum

	Members []EnumMember

	Marshaler Generator
}

func (g *EnumGenerator) Generate(w *file.Package) {
	w.GoImport("fmt")

	g.Doc.Generate(w.Go())

	fmt.Fprintf(w.Go(), "type %s C.int\n\n", g.GoType(0))

	fmt.Fprintf(w.Go(), "const (\n")

	w.Go().Indent()
	for _, member := range g.Members {
		member.Doc.Generate(w.Go())

		fmt.Fprintf(w.Go(), "%s %s = %s\n", member.GoIndentifier(), g.GoType(0), member.Value)
	}
	w.Go().Unindent()

	fmt.Fprint(w.Go(), ")\n\n")

	if g.Marshaler != nil {
		w.RegisterGType(g)
		g.Marshaler.Generate(w)
	}

	fmt.Fprint(w.Go(), "\n")

	if g.CanMarshal() {
		// GoValueInitializer assertion:
		fmt.Fprintf(w.Go(), "var _ %s = %s(0)\n\n", g.Value().WithForeignNamespace("GoValueInitializer"), g.GoType(0))

		fmt.Fprintf(w.Go(), "func (e %s) InitGoValue(v *%s) {\n", g.GoType(0), g.Value().NamespacedGoType(0))
		w.Go().Indent()
		fmt.Fprintf(w.Go(), "v.Init(%s)\n", g.GoTypeName())
		fmt.Fprintf(w.Go(), "v.SetEnum(int(e))\n")
		w.Go().Unindent()
		fmt.Fprintf(w.Go(), "}\n\n")
	}

	// Stringer:
	fmt.Fprintf(w.Go(), "func (e %s) String() string {\n", g.GoType(0))
	w.Go().Indent()
	fmt.Fprintf(w.Go(), "switch e {\n")
	w.Go().Indent()
	for _, member := range g.Enum.Members.Uniques() {
		fmt.Fprintf(w.Go(), "case %s: return \"%s\"\n", member.GoIndentifier(), strcases.SnakeToGo(true, member.GoIndentifier()))
	}
	fmt.Fprintf(w.Go(), "default: return fmt.Sprintf(\"%s(%%d)\", e)\n", g.GoType(0))
	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n")
	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n\n")
}

func NewEnumGenerator(enum *typesystem.Enum) *EnumGenerator {
	members := make([]EnumMember, 0, len(enum.Members))

	for _, member := range enum.Members {
		members = append(members, EnumMember{
			Doc: NewIdentifierGoDocGenerator(member),

			Member: member,
		})
	}

	var marshalGen Generator

	if enum.GLibGetType() != "" {
		marshalGen = NewMarshalEnumGenerator(enum)
	}

	return &EnumGenerator{
		Doc:       NewTypeGoDocGenerator(enum),
		Enum:      enum,
		Members:   members,
		Marshaler: marshalGen,
	}
}
