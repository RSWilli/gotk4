package file

import (
	"io"
)

// goImports maps the package name to its import alias
type goImports map[string]string

func (gi goImports) formatted() io.Reader {
	if len(gi) == 0 {
		return empty
	}

	parts := make([]io.Reader, 0, len(gi)+2)

	parts = append(parts, str("import (\n"))

	for pkg, alias := range gi {
		parts = append(parts, str("\t"))

		if alias != "" {
			parts = append(parts,
				str(alias),
				str(" "),
			)
		}

		parts = append(parts,
			str(`"`),
			str(pkg),
			str("\"\n"),
		)
	}

	parts = append(parts, str(")\n"))

	return io.MultiReader(parts...)
}
