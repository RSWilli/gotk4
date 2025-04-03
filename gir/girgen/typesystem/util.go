package typesystem

import (
	"fmt"
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

// this is an invalid type indentifier, that can be used for types that do not have a
// c or go type, e.g. callbacks. This is chosen becaus it will also break go/cgo compilation when used
const typeInvalid = "// invalid type"

func cleanCType(ctype string) string {
	// use spaces to prevent valid infix replaces

	ctype = strings.ReplaceAll(ctype, "const ", "")
	ctype = strings.ReplaceAll(ctype, " const", "")

	ctype = strings.ReplaceAll(ctype, "volatile ", "")
	ctype = strings.ReplaceAll(ctype, " volatile", "")

	return ctype
}

func CTypeFromAnytype(t gir.AnyType) string {
	switch {
	case t.Array != nil:
		return t.Array.CType
	case t.Type != nil:
		return t.Type.CType
	default:
		panic("invalid anytype")
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
