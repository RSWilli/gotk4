package generators

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/gotmpl"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
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
	Doc  Generator
	Name string
	// AliasFor must be a valid go type
	AliasFor *typesystem.TypeMetadata
}

func (g *AliasGenerator) Generate(w *file.Writer) {
	g.Doc.Generate(w)

	for _, imp := range g.AliasFor.RequiredImports {
		w.GoImport(imp)
	}

	fmt.Fprintf(w.Go(), "type %s = %s\n", g.Name, g.AliasFor.GoType())
}

func NewAliasGenerator(ctx gencontext.GenerationContext, alias gir.Alias) *AliasGenerator {
	if !alias.IsIntrospectable() || !alias.Type.IsIntrospectable() {
		return nil
	}

	resolvedType := ctx.LookupType(alias.Name, "")

	if resolvedType == nil {
		return nil
	}

	goName := strcases.PascalToGo(alias.Name)

	return &AliasGenerator{
		Doc:      NewGoDocGenerator(goName, alias, 0),
		Name:     goName,
		AliasFor: resolvedType,
	}
}
