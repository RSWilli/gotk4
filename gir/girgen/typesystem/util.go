package typesystem

import (
	"log"
	"strconv"
	"strings"
	"unicode"

	"github.com/diamondburned/gotk4/gir"
)

func goPackageNameRuneAllowed(r rune) bool {
	return unicode.IsLetter(r) ||
		unicode.IsDigit(r)
}

// goPackageName converts a GIR package name to a Go package name. It's only
// tested against a known set of GIR files.
func goPackageName(girPkgName string) string {
	return strings.Map(func(r rune) rune {
		if goPackageNameRuneAllowed(r) {
			return unicode.ToLower(r)
		}

		return -1
	}, girPkgName)
}

// parseMajorVersion returns the major version of the GIR namespace in int.
func parseMajorVersion(version string) int {
	major := gir.MajorVersion(version)

	v, err := strconv.Atoi(major)
	if err != nil {
		log.Panicf("invalid major %q", major)
	}

	return v
}

// this is an invalid type indentifier, that can be used for types that do not have a
// c or go type, e.g. callbacks. This is chosen becaus it will also break go/cgo compilation when used
const typeInvalid = "// invalid type"

// GoKeywords includes Go keywords. This is primarily to prevent collisions with
// meaningful Go words.
var GoKeywords = map[string]string{
	// Keywords.
	"break":       "",
	"default":     "",
	"func":        "fn",
	"interface":   "iface",
	"select":      "sel",
	"case":        "",
	"defer":       "",
	"go":          "",
	"map":         "",
	"struct":      "",
	"chan":        "ch",
	"else":        "",
	"goto":        "",
	"package":     "pkg",
	"switch":      "",
	"const":       "",
	"fallthrough": "",
	"if":          "",
	"range":       "",
	"type":        "typ",
	"continue":    "",
	"for":         "",
	"import":      "",
	"return":      "ret",
	"var":         "",
}

func cleanCType(ctype string) string {
	ctype = strings.TrimPrefix(ctype, "const ")
	ctype = strings.TrimPrefix(ctype, "volatile ")

	return ctype
}
