package typesystem

type GoStdLibType struct {
	BaseType

	StdPackages []string
}

// StdImports implements StdLibType.
func (g *GoStdLibType) StdImports() []string {
	return g.StdPackages
}

var Gpointer = &GoStdLibType{
	BaseType: BaseType{
		GirName: "gpointer",
		CTyp:    "gpointer",
		CGoTyp:  "C.gpointer",
		GoTyp:   "unsafe.Pointer",
	},
	StdPackages: []string{"unsafe"},
}

func IsGoStdLibType(t Type) bool {
	switch t.(type) {
	case *GoStdLibType:
		return true
	default:
		// should we handle pointer types?
		return false
	}
}
