package genmain

import (
	"flag"
	"log"

	"maps"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/generators"
	"github.com/diamondburned/gotk4/gir/girgen/logger"
	"github.com/diamondburned/gotk4/gir/girgen/types"
	"github.com/diamondburned/gotk4/gir/girgen/types/typeconv"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

var (
	Output  string
	Verbose bool
	ListPkg bool
	CgoLink bool
)

func init() {
	flag.StringVar(&Output, "o", "", "output directory to mkdir in")
	flag.BoolVar(&Verbose, "v", Verbose, "log verbosely (debug mode)")
	flag.BoolVar(&ListPkg, "l", ListPkg, "only list packages and exit")
	flag.BoolVar(&CgoLink, "cgo-link", CgoLink, "generate everything to link using cgo instead of girepository")
}

// ParseFlag calls flag.Parse() and initializes external global options.
func ParseFlag() {
	flag.Parse()

	if !ListPkg && Output == "" {
		log.Fatalln("Missing -o output directory.")
	}

	if Verbose {
		girgen.DefaultOpts.LogLevel = logger.Debug
	}
}

type Package struct {
	// Name is the pkg-config name.
	Name string
	// Namespaces is the possible namespaces within it. Refer to
	// ./cmd/gir_namespaces.
	Namespaces []string
}

// HasNamespace returns true if the package allows all namespaces or has the
// given namespace in the list.
func (pkg *Package) HasNamespace(n *gir.Namespace) bool {
	if pkg.Namespaces == nil {
		return true
	}

	namespace := gir.VersionedNamespace(n)
	for _, name := range pkg.Namespaces {
		if name == namespace {
			return true
		}
	}

	return false
}

// Data contains generation data that genmain uses to generate.
type Data struct {
	// Module is the Go Module name that the generator is running for. An
	// example is "github.com/diamondburned/gotk4/pkg".
	Module string
	// Packages lists pkg-config packages and optionally the namespaces to be
	// generated. If the list of namespaces is nil, then everything is
	// generated.
	Packages []Package
	// KnownPackages is similar to Packages, but no packages in this list will
	// be used to generate code. This list automatically includes Packages.
	KnownPackages []Package
	// ImportOverrides is the list of imports to defer to another library,
	// usually because it's tedious or impossible to generate.
	//
	// Not included: coreglib (gotk3/gotk3/glib).
	ImportOverrides map[string]string
	// ExternOverrides adds into ImportOverrides packages that were generated
	// from the given GIR repositories, with the map key being the Go module
	// root for those packages. It internally invokes LoadExternOverrides.
	ExternOverrides map[string]gir.Repositories
	// PkgExceptions contains a list of file names that won't be deleted off of
	// pkg/.
	PkgExceptions []string
	// GenerateExceptions contains the keys of the underneath ImportOverrides
	// map.
	GenerateExceptions []string
	// PkgGenerated contains a list of file names that are packages generated
	// using the given Packages list. It is manually updated.
	PkgGenerated []string
	// Preprocessors defines a list of preprocessors that the main generator
	// will use. It's mostly used for renaming colliding types/identifiers.
	Preprocessors []types.Preprocessor
	// Postprocessors is a map of versioned namespace names to a list of
	// functions that are called to modify any file before it is written out.
	Postprocessors map[string][]girgen.Postprocessor
	// ExtraGoContents contains the contents of files that are appended into
	// generated outputs. It is used to add custom implementations of missing
	// functions. It is a simpler version of Postprocessors.
	ExtraGoContents map[string]string
	// Filters defines a list of GIR types to be filtered. The map key is the
	// namespace, and the values are list of names.
	Filters []types.FilterMatcher
	// ProcessConverters is a list of things that can override a type converter.
	ProcessConverters []typeconv.ConversionProcessor
	// DynamicLinkNamespaces lists namespaces that should be generated directly
	// using Cgo. It includes important core packages as well as packages that
	// are small but performance-sensitive.
	DynamicLinkNamespaces []string
	// SingleFile, if true, will make all NamespaceGenerators generate a single
	// output file per package instead of correlating it to the source file.
	SingleFile bool
}

// Overlay joins the given list of data into a single Data. The last Data in the
// list will be used for generation.
func Overlay(data ...Data) Data {
	overlay := data[:len(data)-1]
	last := data[len(data)-1]

	for _, datum := range overlay {
		if last.ExternOverrides == nil {
			last.ExternOverrides = make(map[string]gir.Repositories)
		}

		last.KnownPackages = append(last.KnownPackages, datum.Packages...)
		last.ExternOverrides[datum.Module] = MustLoadPackages(datum.Packages)
		last.Preprocessors = append(last.Preprocessors, datum.Preprocessors...)
		last.Filters = append(last.Filters, datum.Filters...)
		last.ProcessConverters = append(last.ProcessConverters, datum.ProcessConverters...)
		last.DynamicLinkNamespaces = append(last.DynamicLinkNamespaces, datum.DynamicLinkNamespaces...)
	}

	return last
}

// Run runs the application.
func Run(data Data) {
	ParseFlag()

	log.Println("loading packages...")
	// load known packages first and then the packages we want to generate,
	// to keep the order for the typesystem
	repos := MustLoadPackages(data.KnownPackages)
	MustAddPackages(&repos, data.Packages)
	PrintAddedPkgs(repos)

	if ListPkg {
		return
	}

	Generate(repos, data)
}

// Generate generates the packages based on the given data.
func Generate(repos gir.Repositories, data Data) {
	err := CleanDirectory(Output, data.PkgExceptions)

	if err != nil {
		log.Fatalln("failed to clean output directory:", err)
	}

	overrides := data.ImportOverrides
	if overrides == nil {
		overrides = map[string]string{}
	}

	for mod, extern := range data.ExternOverrides {
		maps.Copy(overrides, LoadExternOverrides(mod, extern))
	}

	types.ApplyPreprocessors(repos, data.Preprocessors)

	tsCfg := typesystem.Config{
		Namespaces: map[string]typesystem.NamespaceConfig{
			"cairo-1": {
				Ignored: true, // FIXME: manually implemented
			},
			"Atspi-2": {
				Ignored: true, // Missing AtspiDevice
			},
			"GLib-2": {
				MinVersion: "2.80",
				ManualTypes: []typesystem.Type{
					&typesystem.ForeignType{
						SourceNamespace: &typesystem.Namespace{GoName: "gbox"},
						Type: &typesystem.Callback{
							BaseType: typesystem.BaseType{
								GirName:       "DestroyNotify",
								GoTyp:         "DestroyNotify",
								CGoTyp:        "C.GDestroyNotify",
								CTyp:          "GDestroyNotify",
								GlibGetTypeFn: "",
							},
							Parameters:     &typesystem.Parameters{},
							TrampolineName: "callbackDelete",
						},
					},
				},
				// Ignored: []typesystem.IgnoreFunc{
				// 	typesystem.IgnoreMatching(typesystem.GIRCallbackPattern("DestroyNotify")),
				// },
			},
			"GObject-2": {
				ManualTypes: []typesystem.Type{
					&typesystem.ForeignType{
						SourceNamespace: &typesystem.Namespace{GoName: "coreglib"},
						Type: &typesystem.Class{
							BaseType: typesystem.BaseType{
								GirName: "Object",
								GoTyp:   "Object",
								CTyp:    "GObject",
								CGoTyp:  "C.GObject",
							},
							GoInterfaceName:              "Objector",
							Doc:                          typesystem.Doc{},
							GoUnsafeBorrowFunction:       "TODO",
							GoUnsafeTransferFullFunction: "AssumeOwnership",
							GoUnsafeTransferNoneFunction: "Take",
							GoUnsafeToGlibNoneMethod:     "TODO",
							GoUnsafeToGlibFullMethod:     "TODO",
						},
					},
					&typesystem.ForeignType{
						SourceNamespace: &typesystem.Namespace{GoName: "coreglib"},
						Type: &typesystem.Record{
							BaseType: typesystem.BaseType{
								GirName: "ObjectClass",
								GoTyp:   "ObjectClass",
								CTyp:    "GObjectClass",
								CGoTyp:  "C.GObjectClass",
							},
						},
					},
					&typesystem.ForeignType{
						SourceNamespace: &typesystem.Namespace{GoName: "coreglib"},
						Type: &typesystem.Record{
							BaseType: typesystem.BaseType{
								GirName: "Value",
								GoTyp:   "Value",
								CTyp:    "GValue",
								CGoTyp:  "C.GValue",
							},
						},
					},
					// &typesystem.ForeignType{
					// 	SourceNamespace: &typesystem.Namespace{GoName: "coreglib"},
					// 	Type: &typesystem.Class{
					// 		BaseType: typesystem.BaseType{
					// 			GirName: "ParamSpec",
					// 			GoTyp:   "ParamSpec",
					// 			CTyp:    "GParamSpec",
					// 			CGoTyp:  "C.GParamSpec",
					// 		},
					// 	},
					// },
				},
				IgnoredDefinitions: []typesystem.IgnoreFunc{
					// manually implemented, but hidden from the user
					typesystem.IgnoreMatching(typesystem.GIRRecordPattern("ParamSpec")),
				},
			},
		},
	}

	ts := typesystem.FromRepositories(tsCfg, repos)

	// TODO: add a hook stage here, where the user can modify the chosen names of the typesystem

	var reposToGenerate []*typesystem.Repository

	for _, repo := range ts.Repositories {
		for _, pkg := range data.Packages {
			if pkg.Name == repo.Pkg {
				reposToGenerate = append(reposToGenerate, repo)
				break
			}
		}
	}

	var gen []generators.Generator

	if !CgoLink {
		gen = generators.WithDynamicLinking(reposToGenerate)
	} else {
		gen = generators.WithRuntimeLinking(reposToGenerate)
	}

	// TODO: add a hook stage here, where the user can modify all generators in "gen"

	for _, g := range gen {
		w := file.NewWriter(Output)

		g.Generate(w)

		// in theory we can add a pre commit stage here, but I don't know if this is useful

		err := w.Commit()

		if err != nil {
			panic(err)
		}
	}
}
