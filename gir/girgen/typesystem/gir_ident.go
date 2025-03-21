package typesystem

import (
	"fmt"
	"strings"
)

type GIRKind string

const (
	GIRKindAny           GIRKind = ""
	GIRKindClass         GIRKind = "class"
	GIRKindInterface     GIRKind = "interface"
	GIRKindEnum          GIRKind = "enum"
	GIRKindBitfield      GIRKind = "bitfield"
	GIRKindMember        GIRKind = "member"
	GIRKindField         GIRKind = "field"
	GIRKindMethod        GIRKind = "method"
	GIRKindFunction      GIRKind = "function"
	GIRKindCallback      GIRKind = "callback"
	GIRKindRecord        GIRKind = "record"
	GIRKindSignal        GIRKind = "signal"
	GIRKindConstructor   GIRKind = "constructor"
	GIRKindVirtualMethod GIRKind = "virtual_method"
	GIRKindConstant      GIRKind = "constant"
	GIRKindUnion         GIRKind = "union"
	GIRKindAlias         GIRKind = "alias"
)

const GIRWildCard = "*"

type GIRIdentifier struct {
	Parent string
	Name   string
	Kind   GIRKind
}

func GIRAnyPattern(str string) GIRIdentifier {
	return ParseGIRIdentifier(str, GIRKindAny)
}
func GIRClassPattern(str string) GIRIdentifier {
	return ParseGIRIdentifier(str, GIRKindClass)
}
func GIRInterfacePattern(str string) GIRIdentifier {
	return ParseGIRIdentifier(str, GIRKindInterface)
}
func GIREnumPattern(str string) GIRIdentifier {
	return ParseGIRIdentifier(str, GIRKindEnum)
}
func GIRBitfieldPattern(str string) GIRIdentifier {
	return ParseGIRIdentifier(str, GIRKindBitfield)
}
func GIRMemberPattern(str string) GIRIdentifier {
	return ParseGIRIdentifier(str, GIRKindMember)
}
func GIRFieldPattern(str string) GIRIdentifier {
	return ParseGIRIdentifier(str, GIRKindField)
}
func GIRMethodPattern(str string) GIRIdentifier {
	return ParseGIRIdentifier(str, GIRKindMethod)
}
func GIRFunctionPattern(str string) GIRIdentifier {
	return ParseGIRIdentifier(str, GIRKindFunction)
}
func GIRCallbackPattern(str string) GIRIdentifier {
	return ParseGIRIdentifier(str, GIRKindCallback)
}
func GIRRecordPattern(str string) GIRIdentifier {
	return ParseGIRIdentifier(str, GIRKindRecord)
}
func GIRSignalPattern(str string) GIRIdentifier {
	return ParseGIRIdentifier(str, GIRKindSignal)
}
func GIRConstructorPattern(str string) GIRIdentifier {
	return ParseGIRIdentifier(str, GIRKindConstructor)
}
func GIRVirtualMethodPattern(str string) GIRIdentifier {
	return ParseGIRIdentifier(str, GIRKindVirtualMethod)
}
func GIRConstantPattern(str string) GIRIdentifier {
	return ParseGIRIdentifier(str, GIRKindConstant)
}
func GIRUnionPattern(str string) GIRIdentifier {
	return ParseGIRIdentifier(str, GIRKindUnion)
}
func GIRAliasPattern(str string) GIRIdentifier {
	return ParseGIRIdentifier(str, GIRKindAlias)
}

func ParseGIRIdentifier(str string, kind GIRKind) GIRIdentifier {
	parts := strings.Split(str, ".")

	switch len(parts) {
	case 1:
		return GIRIdentifier{
			Parent: "",
			Name:   parts[0],
			Kind:   kind,
		}
	case 2:
		return GIRIdentifier{
			Parent: parts[0],
			Name:   parts[1],
			Kind:   kind,
		}
	default:
		panic("invalid GIR identifier")
	}
}

func (id GIRIdentifier) String() string {
	if id.Parent == "" {
		return id.Name
	}
	// Dot notation: Parent.Name
	return fmt.Sprintf("%s.%s", id.Parent, id.Name)
}

// Matches compares the concrete identifier id (without wildcards) with the pattern identifier
// (potentially with wildcards or kind any)
func (id GIRIdentifier) Matches(pattern GIRIdentifier) bool {
	// Invalid pattern: name must be provided
	if pattern.Name == "" {
		return false
	}

	if pattern.Kind != GIRKindAny && id.Kind != pattern.Kind {
		return false
	}

	if pattern.Parent != "" && pattern.Parent != GIRWildCard && id.Parent != pattern.Parent {
		return false
	}

	if pattern.Name != GIRWildCard && id.Name != pattern.Name {
		return false
	}

	return true
}
