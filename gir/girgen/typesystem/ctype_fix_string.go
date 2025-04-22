package typesystem

// OverriddenCTypeString is a type that wraps a type if the ctype differs from the ctype
// of the gir type that is resolved. This produces an implicit cast in the generated code.
type OverriddenCTypeString struct {
	Actual *StringPrimitive

	// These types already contain pointers here:
	Ctype   string
	Cgotype string
}

// maxPointersAllowed implements maxPointerConstrainedType.
func (i *OverriddenCTypeString) maxPointersAllowed() int {
	return i.Actual.maxPointersAllowed()
}

// minPointersRequired implements minPointerConstrainedType.
func (i *OverriddenCTypeString) minPointersRequired() int {
	return i.Actual.minPointersRequired()
}

// OriginalCGoType implements OverriddenCType.
func (i *OverriddenCTypeString) OriginalCGoType(pointers int) string {
	return i.Actual.CGoType(pointers)
}

// CGoType implements Type.
func (i *OverriddenCTypeString) CGoType(pointers int) string {
	return i.Cgotype
}

// CType implements Type.
func (i *OverriddenCTypeString) CType(pointers int) string {
	return i.Ctype
}

// GIRName implements Type.
func (i *OverriddenCTypeString) GIRName() string {
	return i.Actual.GIRName()
}

// GoType implements Type.
func (i *OverriddenCTypeString) GoType(pointers int) string {
	return i.Actual.GoType(pointers)
}

// GoTypeRequiredImport implements Type.
func (i *OverriddenCTypeString) GoTypeRequiredImport() (alias string, module string) {
	return i.Actual.GoTypeRequiredImport()
}

var _ Type = (*OverriddenCTypeString)(nil)
var _ OverriddenCType = (*OverriddenCTypeString)(nil)
var _ minPointerConstrainedType = (*OverriddenCTypeString)(nil)
var _ maxPointerConstrainedType = (*OverriddenCTypeString)(nil)
