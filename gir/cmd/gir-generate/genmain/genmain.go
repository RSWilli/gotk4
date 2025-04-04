package genmain

import (
	"flag"
	"fmt"
	"log"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen"
	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/generators"
	"github.com/diamondburned/gotk4/gir/girgen/logger"
	"github.com/diamondburned/gotk4/gir/girgen/types"
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
	// ExternOverrides adds into ImportOverrides packages that were generated
	// from the given GIR repositories, with the map key being the Go module
	// root for those packages. It internally invokes LoadExternOverrides.
	ExternOverrides map[string]gir.Repositories
	// Preprocessors defines a list of preprocessors that the main generator
	// will use. It's mostly used for renaming colliding types/identifiers.
	Preprocessors []types.Preprocessor

	Config typesystem.Config
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
		last.Config = last.Config.Combine(datum.Config)
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
	err := CleanGeneratedFiles(Output)

	if err != nil {
		log.Fatalln("failed to clean output directory:", err)
	}

	importBaseURIs := map[string]string{}

	for _, r := range repos {
		for _, ns := range r.Namespaces {
			importBaseURIs[fmt.Sprintf("%s-%d", ns.Name, ns.Version.Major)] = data.Module
		}
	}

	for mod, extern := range data.ExternOverrides {
		for _, r := range extern {
			for _, ns := range r.Namespaces {
				importBaseURIs[fmt.Sprintf("%s-%d", ns.Name, ns.Version.Major)] = mod
			}
		}
	}

	types.ApplyPreprocessors(repos, data.Preprocessors)

	ts := typesystem.FromRepositories(data.Config, repos)

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
		w := file.NewPackage(Output, importBaseURIs)

		g.Generate(w)

		// in theory we can add a pre commit stage here, but I don't know if this is useful

		err := w.Commit()

		if err != nil {
			panic(err)
		}
	}
}
