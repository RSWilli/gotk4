package typesystem

type OverriddenCType interface {
	OriginalCGoType(pointers int) string
}
