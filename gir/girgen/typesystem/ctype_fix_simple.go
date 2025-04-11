package typesystem

// OverriddenCTypeSimple is a type that wraps a type if the ctype differs from the ctype
// of the gir type that is resolved. This produces an implicit cast in the generated code.
type OverriddenCTypeSimple struct {
	Actual Type // Class or Interface

	Ctype   string
	Cgotype string
}

// OriginalCGoType implements OverriddenCType.
func (i *OverriddenCTypeSimple) OriginalCGoType(pointers int) string {
	return i.Actual.CGoType(pointers)
}

// CGoType implements Type.
func (i *OverriddenCTypeSimple) CGoType(pointers int) string {
	return GetPointers(pointers) + i.Cgotype
}

// CType implements Type.
func (i *OverriddenCTypeSimple) CType(pointers int) string {
	return i.Ctype + GetPointers(pointers)
}

// GIRName implements Type.
func (i *OverriddenCTypeSimple) GIRName() string {
	return i.Actual.GIRName()
}

// GoType implements Type.
func (i *OverriddenCTypeSimple) GoType(pointers int) string {
	return i.Actual.GoType(pointers)
}

// GoTypeRequiredImport implements Type.
func (i *OverriddenCTypeSimple) GoTypeRequiredImport() (alias string, module string) {
	return i.Actual.GoTypeRequiredImport()
}

var _ Type = (*OverriddenCTypeSimple)(nil)
var _ OverriddenCType = (*OverriddenCTypeSimple)(nil)
