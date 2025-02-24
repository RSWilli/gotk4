package file

import (
	"io"
	"strings"
)

type cIncludes map[string]struct{}

func (ci cIncludes) Reader() io.Reader {
	var b strings.Builder

	for i := range ci {
		b.WriteString("// #include <")
		b.WriteString(i)
		b.WriteString(">\n")
	}

	return strings.NewReader(b.String())
}
