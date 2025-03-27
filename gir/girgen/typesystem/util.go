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

// decreaseAnyTypePointers removes one pointer from the ctype of the [gir.AnyType], but doesn't modify the original
// value
func decreaseAnyTypePointers(t gir.AnyType) gir.AnyType {
	var newType *gir.Type
	var newArray *gir.Array

	if t.Type != nil {
		newType = &gir.Type{
			XMLName:        t.Type.XMLName,
			Name:           t.Type.Name,
			CType:          strings.TrimSuffix(t.Type.CType, "*"),
			Introspectable: t.Type.Introspectable,
			DocElements:    t.Type.DocElements,
			Types:          t.Type.Types,
		}
	}

	if t.Array != nil {
		newArray = &gir.Array{
			XMLName:        t.Array.XMLName,
			Name:           t.Array.Name,
			CType:          strings.TrimSuffix(t.Array.CType, "*"),
			Length:         t.Array.Length,
			ZeroTerminated: t.Array.ZeroTerminated,
			FixedSize:      t.Array.FixedSize,
			Introspectable: t.Array.Introspectable,
			Type:           t.Array.Type,
		}
	}

	return gir.AnyType{
		Type:  newType,
		Array: newArray,
	}
}

func debugCTypeFromAnytype(t gir.AnyType) string {
	switch {
	case t.Array != nil:
		return "(array) " + t.Array.CType
	case t.Type != nil:
		return t.Type.CType
	default:
		panic("invalid anytype")
	}
}

func infoFromAnyGir(girAny any) (string, GIRKind) {
	switch t := girAny.(type) {
	case gir.Class:
		return t.Name, GIRKindClass
	case gir.Interface:
		return t.Name, GIRKindInterface
	case gir.Callback:
		return t.Name, GIRKindCallback
	case gir.Field:
		return t.Name, GIRKindField
	case gir.Enum:
		return t.Name, GIRKindEnum
	case gir.Bitfield:
		return t.Name, GIRKindBitfield
	case gir.Member:
		return t.Name(), GIRKindMember
	case gir.Record:
		return t.Name, GIRKindRecord
	case gir.Constant:
		return t.Name, GIRKindConstant
	case gir.Constructor:
		return t.Name, GIRKindConstructor
	case gir.Method:
		return t.Name, GIRKindMethod
	case gir.VirtualMethod:
		return t.Name, GIRKindVirtualMethod
	case gir.Union:
		return t.Name, GIRKindUnion
	case gir.Alias:
		return t.Name, GIRKindAlias
	case gir.Function:
		return t.Name, GIRKindFunction
	case gir.Signal:
		return t.Name, GIRKindSignal
	default:
		panic(fmt.Sprintf("received unhandled type: %T", t))
	}
}
