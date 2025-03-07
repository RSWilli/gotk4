package typesystem

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type Member struct {
	Doc    Doc
	GoName string
	Value  string
}

func GetMembers(ms []gir.Member) []*Member {
	var mems []*Member

	for _, girM := range ms {
		if m := NewMember(girM); m != nil {
			mems = append(mems, m)
		}
	}

	return mems
}

func NewMember(m gir.Member) *Member {
	if !m.IsIntrospectable() {
		return nil
	}

	return &Member{
		Doc:    NewDoc(&m.InfoAttrs, &m.InfoElements),
		GoName: formatMember(m),
		Value:  "C." + m.CIdentifier, // using the C identifier directly to avoid string quoting issues etc.
	}
}

var numberMap = map[rune]string{
	'0': "Zero",
	'1': "One",
	'2': "Two",
	'3': "Three",
	'4': "Four",
	'5': "Five",
	'6': "Six",
	'7': "Seven",
	'8': "Eight",
	'9': "Nine",
}

// formatMember returns the enum member's Go name.
func formatMember(member gir.Member) string {
	// Pop the namespace off. Probably works only most of the time.
	if parts := strings.SplitN(member.CIdentifier, "_", 2); len(parts) == 2 {
		member.CIdentifier = parts[1]
	}

	memberName := strcases.SnakeToGo(true, strings.ToLower(member.CIdentifier))

	// TODO: prepend GoName instead.
	r, sz := utf8.DecodeRuneInString(memberName)
	if sz > 0 && unicode.IsNumber(r) {
		memberName = numberMap[r] + memberName[sz:]
	}

	return memberName
}
