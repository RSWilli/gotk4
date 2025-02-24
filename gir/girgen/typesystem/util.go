package typesystem

import (
	"fmt"
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

func goImportPath(base, name, version string) string {
	folder := goPackageName(name)

	if version := parseMajorVersion(version); version > 1 {
		folder += fmt.Sprintf("/v%d", version)
	}

	return base + "/" + folder
}
