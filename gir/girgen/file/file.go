package file

import (
	"io"

	"github.com/diamondburned/gotk4/gir/girgen/file/internal"
)

// File contains the shared logic between the file.go and file_export.go
type File struct {
	cPreamble internal.CodeWriter

	goContents internal.CodeWriter

	goImports goImports
}

var ImportAnonymous = "_"

var coreglibPkg = "github.com/diamondburned/gotk4/pkg/core"

func (d *File) GoImportCore(pkg string) {
	d.GoImport(coreglibPkg + "/" + pkg)
}

func (d *File) GoImport(pkg string) {
	d.GoImportAliased(pkg, "")
}

func (d *File) GoImportAnonymous(pkg string) {
	d.GoImportAliased(pkg, ImportAnonymous)
}

func (d *File) GoImportAliased(pkg string, alias string) {
	if d.goImports == nil {
		d.goImports = make(goImports)
	}

	currentImport, ok := d.goImports[pkg]

	if ok && alias == ImportAnonymous {
		return // keep already imported name
	}

	if ok && currentImport != alias {
		panic("tried to import the same module twice with different aliases")
	}

	if ok {
		return // already in map with the same name
	}

	d.goImports[pkg] = alias
}

type CodeWriter interface {
	io.Writer
	Indent()
	Unindent()
	NewSection()
}

func (d *File) Go() CodeWriter {
	return &d.goContents
}

func (d *File) C() CodeWriter {
	return &d.cPreamble
}

func (d *File) empty() bool {
	return len(d.goImports) == 0 &&
		d.cPreamble.Len() == 0 &&
		d.goContents.Len() == 0
}

func (d *File) c() io.Reader {
	if d.cPreamble.Len() == 0 {
		return empty
	}

	return internal.NewPrependLinesReader("// ", &d.cPreamble)
}
