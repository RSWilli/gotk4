package typesystem

import (
	"strings"
)

func CountCTypePointers(ctype string) int {
	return strings.Count(ctype, "*")
}

func GetPointers(count int) string {
	return strings.Repeat("*", count)
}
