package file

import (
	"io"
	"sync"

	"github.com/diamondburned/gotk4/gir/girgen/file/internal"
)

// filedata is the file filedata embedded by [Writer] and contains the shared logic between the file.go and file_export.go
type filedata struct {
	used bool

	m sync.Mutex

	// cIncludes contains the deduped c #include directives that will be placed at the top
	cIncludes cIncludes
	cPreamble internal.CodeWriter

	goContents internal.CodeWriter

	goImports goImports
}

func (d *filedata) CInclude(header string) {
	d.m.Lock()
	defer d.m.Unlock()
	d.used = true

	if d.cIncludes == nil {
		d.cIncludes = make(cIncludes)
	}

	d.cIncludes[header] = struct{}{}
}

var ImportAnonymous = "_"

var coreglibPkg = "github.com/diamondburned/gotk4/pkg/core"

func (d *filedata) GoImportCore(pkg string) {
	d.GoImport(coreglibPkg + "/" + pkg)
}

func (d *filedata) GoImport(pkg string) {
	d.GoImportAliased(pkg, "")
}

func (d *filedata) GoImportAnonymous(pkg string) {
	d.GoImportAliased(pkg, ImportAnonymous)
}

func (d *filedata) GoImportAliased(pkg string, alias string) {
	d.m.Lock()
	defer d.m.Unlock()
	d.used = true

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
}

func (d *filedata) Go() CodeWriter {
	d.used = true

	return &d.goContents
}

func (d *filedata) C() CodeWriter {
	d.used = true

	return &d.cPreamble
}

func (d *filedata) empty() bool {
	d.used = true

	return !d.used
}

func (d *filedata) c() io.Reader {
	if d.cPreamble.Len() == 0 {
		return empty
	}

	return internal.NewPrependLinesReader("// ", &d.cPreamble)
}
