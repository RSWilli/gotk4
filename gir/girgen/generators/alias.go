package generators

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/gotmpl"
	"github.com/diamondburned/gotk4/gir/girgen/types"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

// Deprecated: old
var aliasTmpl = gotmpl.NewGoTemplate(`
	{{ $name := (PascalToGo .Name) }}
	{{ GoDoc . 0 }}
	type {{ $name }} = {{ .GoType }}
`)

// Deprecated: old
type aliasData struct {
	*gir.Alias
	GoType string
}

// CanGenerateAlias returns false if this alias cannot be generated.
//
// Deprecated: old
func CanGenerateAlias(gen FileGenerator, alias *gir.Alias) bool {
	if !alias.IsIntrospectable() || types.Filter(gen, alias.Name, alias.CType) {
		return false
	}

	return types.Resolve(gen, alias.Type) != nil
}

// GenerateAlias generates an alias declaration into the given file generator.
// If the generation fails or is ignored, then false is returned.
//
// Deprecated: old
func GenerateAlias(gen FileGeneratorWriter, alias *gir.Alias) bool {
	if !CanGenerateAlias(gen, alias) {
		return false
	}

	resolved := types.Resolve(gen, alias.Type)

	goType := resolved.PublicType(resolved.NeedsNamespace(gen.Namespace()))
	if goType == "" {
		// Use the C type directly if we can't find the Go equivalent.
		goType = "C." + alias.Type.CType
	}

	writer := FileWriterFromType(gen, alias)
	resolved.ImportPubl(gen, writer.Header())
	writer.Pen().WriteTmpl(aliasTmpl, aliasData{
		Alias:  alias,
		GoType: goType,
	})

	return true
}

type AliasGenerator struct {
	Doc SubGenerator

	*typesystem.Alias
}

func (g *AliasGenerator) Generate(w *file.Package) {
	g.Doc.Generate(w.Go())

	// TODO: imports of foreign type namespaces

	fmt.Fprintf(w.Go(), "type %s = %s\n", g.GoType(0), g.AliasedType.NamespacedGoType(0))
}

func NewAliasGenerator(alias *typesystem.Alias) *AliasGenerator {
	return &AliasGenerator{
		Doc:   NewTypeGoDocGenerator(alias),
		Alias: alias,
	}
}
