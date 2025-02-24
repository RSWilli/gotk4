package file_test

import (
	"fmt"
	"testing"

	"github.com/diamondburned/gotk4/gir/girgen/file"
)

func Test(t *testing.T) {
	w := file.NewWriter("/home/wbartel/projects/go-gst/gotk4/pkg")

	w.SetGoPackageName("test", 1)

	fmt.Fprintf(w.C(), "extern foobar()\n")
	fmt.Fprintf(w.Exported.Go(), "func test() {}\n")

	w.GoImportAnonymous("runtime")

	w.Commit()
}
