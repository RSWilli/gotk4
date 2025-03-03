package value

import (
	"regexp"
	"strings"

	"github.com/diamondburned/gotk4/gir"
)

// guessParameterOutput guesses the parameter output using various clues to make
// up for GIR's painful shortcomings.
func guessParameterOutput(param gir.ParameterAttrs) string {
	switch param.Direction {
	case "out", "in", "inout":
		return param.Direction
	}

	if param.Doc != nil && strings.HasPrefix(param.Doc.String, "Return location") {
		panic("should this be a out param?")
		// param.Direction = "out"
		// return "out"
	}

	return "in"
}

func AnyTypeC(anyT gir.AnyType) (ctype string, is_array bool) {
	switch {
	case anyT.Array != nil:
		return "", true // TODO: return correct type
	case anyT.Type != nil:
		return anyT.Type.CType, false
	default:
		panic("invalid anytype")
	}
}

func AnyTypeName(anyT gir.AnyType) (typename string, is_array bool) {
	switch {
	case anyT.Array != nil:
		return "arrayOf...", true // TODO: return correct type
	case anyT.Type != nil:
		return anyT.Type.Name, false
	default:
		panic("invalid anytype")
	}
}

func IsVoid(anyT gir.AnyType) bool {
	t, isArray := AnyTypeC(anyT)

	if isArray {
		return false
	}

	return t == "void"
}

var validGoIndentRegex = regexp.MustCompile("^[a-zA-Z_][a-zA-Z_0-9]*$")

func isValidGoIndent(indent string) bool {
	return validGoIndentRegex.Match([]byte(indent))
}
