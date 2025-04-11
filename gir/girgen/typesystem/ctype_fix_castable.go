package typesystem

type OverriddenCTypeCastable struct {
	Actual CastableType

	Ctype   string
	Cgotype string
}

// canBeCasted implements CastableType.
func (o *OverriddenCTypeCastable) canBeCasted() {}

// OriginalCGoType implements OverriddenCType.
func (o *OverriddenCTypeCastable) OriginalCGoType(pointers int) string {
	return o.Actual.CGoType(pointers)
}

// CGoType implements Type.
func (o *OverriddenCTypeCastable) CGoType(pointers int) string {
	return GetPointers(pointers) + o.Cgotype
}

// CType implements Type.
func (o *OverriddenCTypeCastable) CType(pointers int) string {
	return o.Ctype + GetPointers(pointers)
}

// GIRName implements Type.
func (o *OverriddenCTypeCastable) GIRName() string {
	return o.Actual.GIRName()
}

// GoType implements Type.
func (o *OverriddenCTypeCastable) GoType(pointers int) string {
	return o.Actual.GoType(pointers)
}

// GoTypeRequiredImport implements Type.
func (o *OverriddenCTypeCastable) GoTypeRequiredImport() (alias string, module string) {
	return o.Actual.GoTypeRequiredImport()
}

var _ CastableType = (*OverriddenCTypeCastable)(nil)
var _ OverriddenCType = (*OverriddenCTypeCastable)(nil)
