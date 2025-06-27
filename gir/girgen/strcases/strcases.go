// Package strcases provides helper functions to convert between string cases,
// such as Pascal Case, snake_case and Go's Mixed Caps, along with various
// special cases.
package strcases

import (
	"log"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	_ "embed"
)

//go:embed capitalized.txt
var capitalizedTXT string

//go:embed replaced.txt
var replacedTXT string

var (
	// goIdentRegex matches valid go identifiers (must not start with a number)
	goIdentRegex   = regexp.MustCompile(`^[_A-Za-z]\w+`)
	snakeRegex     = regexp.MustCompile(`[_0-9]+\w`)
	pascalSpecials = strings.Split(capitalizedTXT, "\n")
	pascalWords    = map[string]string{}

	pascalRegex        *regexp.Regexp
	pascalPostReplacer *strings.Replacer
)

func initPascalWords() {
	for _, line := range strings.Split(replacedTXT, "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		words := strings.Split(line, "->")
		if len(words) != 2 {
			log.Fatalf("invalid replace %q", line)
		}

		words[0] = strings.TrimSpace(words[0])
		words[1] = strings.TrimSpace(words[1])
		pascalWords[words[0]] = words[1]
	}
}

func initPascalRegex() {
	fullRegex := strings.Builder{}
	fullRegex.Grow(256)
	fullRegex.WriteByte('(')

	for i, special := range pascalSpecials {
		if special == "" {
			continue
		}
		if i > 0 {
			fullRegex.WriteByte('|')
		}
		fullRegex.WriteString(special)
	}

	fullRegex.WriteByte(')')

	// Must account for the next character being either EOF or a capitalized
	// letter to avoid cases like "IDentifier".
	fullRegex.WriteString("([A-Z0-9]|$)")

	pascalRegex = regexp.MustCompile(fullRegex.String())
}

func initPascalPostReplacer() {
	postReplacerArgs := make([]string, len(pascalWords)*2)
	for from, to := range pascalWords {
		postReplacerArgs = append(postReplacerArgs, from, to)
	}

	pascalPostReplacer = strings.NewReplacer(postReplacerArgs...)
}

func init() {
	initPascalWords()
	initPascalRegex()
	initPascalPostReplacer()
}

// isLower returns true if the string is all lower-cased.
func isLower(s string) bool {
	return strings.IndexFunc(s, unicode.IsUpper) == -1
}

// guessSnake guesses if the given name is snake-cased or not.
func guessSnake(name string) (snake bool) {
	return strings.Contains(name, "_") || isLower(name)
}

// Go converts either pascal or snake case to the Go name. The original casing
// is inferred from the given name.
func Go(name string) string {
	if guessSnake(name) {
		return SnakeToGo(true, name)
	} else {
		return PascalToGo(name)
	}
}

// PascalToGo converts regular Pascal case to Go.
func PascalToGo(pascal string) string {
	// Use a for loop so that we can handle cases where acronyms are next to
	// each other, such as "SkuId" -> "SKUId" -> "SKUID".
	for {
		pascal2 := pascalRegex.ReplaceAllStringFunc(pascal, strings.ToUpper)
		if pascal2 == pascal {
			break
		}
		pascal = pascal2
	}

	pascal = pascalPostReplacer.Replace(pascal)

	if pascal == "" {
		panic("empty pascal string")
	}

	if !goIdentRegex.MatchString(pascal) {
		// This is a last resort to ensure that the string is a valid Go
		pascal = "Gotk" + pascal
	}

	return pascal
}

// ReceiverName returns the first letter in lower-case.
func ParamNameToGo(p string) string {
	return SnakeToGo(false, p)
}

// ReceiverName returns the first letter in lower-case.
func ReceiverName(p string) string {
	r, sz := utf8.DecodeRuneInString(p)
	if sz > 0 && r != utf8.RuneError {
		// FIXME: this could return "_" which is not a valid receiver
		return string(unicode.ToLower(r))
	}

	return string(p[0]) // fallback
}

// UnexportPascal converts the PascalToGo string to be unexported.
func UnexportPascal(pascal string) string {
	runes := []rune(pascal)
	if len(runes) < 1 {
		return snakeNoGo(strings.ToLower(pascal))
	}

	var i int
	for i < len(runes) && unicode.IsUpper(runes[i]) {
		i++
	}

	if i > 1 {
		i--
	}

	pascal = strings.ToLower(string(runes[:i])) + string(runes[i:])
	pascal = snakeNoGo(pascal)

	return pascal
}

// SnakeToGo converts snake case to Go's special case. If Pascal is true, then
// the first letter is capitalized.
func SnakeToGo(pascal bool, snakeString string) string {
	if pascal {
		snakeString = "_" + snakeString
	}

	snakeString = snakeRegex.ReplaceAllStringFunc(snakeString,
		func(orig string) string {
			orig = strings.ToUpper(orig)
			orig = strings.Replace(orig, "_", "", 2)
			return orig
		},
	)

	if !pascal {
		return snakeNoGo(snakeString)
	}

	return PascalToGo(snakeString)
}

// KebabToGo converts kebab case to Go's special case. See SnakeToGo.
func KebabToGo(pascal bool, kebabString string) string {
	return SnakeToGo(pascal, strings.ReplaceAll(kebabString, "-", "_"))
}

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

	// words that may collide with go stdlib packages
	"context": "_context",
	"strings": "_strings",
	"fmt":     "_fmt",
}

// GoBuiltinTypes contains Go built-in types.
var GoBuiltinTypes = map[string]string{
	// Types.
	"bool":       "",
	"byte":       "",
	"complex128": "cmplx",
	"complex64":  "cmplx",
	"error":      "err",
	"float32":    "",
	"float64":    "",
	"int":        "",
	"int16":      "",
	"int32":      "",
	"int64":      "",
	"int8":       "",
	"rune":       "",
	"string":     "str",
	"uint":       "",
	"uint16":     "",
	"uint32":     "",
	"uint64":     "",
	"uint8":      "",
	"uintptr":    "",
}

// CGoField formats the C field name to not be confused with a Go keyword.
// See https://golang.org/cmd/cgo/#hdr-Go_references_to_C.
func CGoField(field string) string {
	_, keyword := GoKeywords[field]
	if keyword {
		return "_" + field
	}
	return field
}

// snakeNoGo ensures the snake-case string is never a Go keyword.
func snakeNoGo(snake string) string {
	s, isKeyword := GoKeywords[snake]
	if isKeyword {
		if s != "" {
			return s
		}
		return "_" + snake
	}

	s, isType := GoBuiltinTypes[snake]
	if isType {
		if s != "" {
			return s
		}
		return "_" + snake
	}

	return snake
}
