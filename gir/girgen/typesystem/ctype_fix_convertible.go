package typesystem

// OverriddenCTypeConvertible is a type that wraps a type if the ctype differs from the ctype
// of the gir type that is resolved. This produces an implicit cast in the generated code.
type OverriddenCTypeConvertible struct {
	Actual ConvertibleType // Class or Interface

	Ctype   string
	Cgotype string
}

// OriginalCGoType implements OverriddenCType.
func (i *OverriddenCTypeConvertible) OriginalCGoType(pointers int) string {
	return i.Actual.CGoType(pointers)
}

// CanTransferFromGlib implements ConvertibleType.
func (i *OverriddenCTypeConvertible) CanTransferFromGlib(transfer TransferOwnership) bool {
	return i.Actual.CanTransferFromGlib(transfer)
}

// CanTransferToGlib implements ConvertibleType.
func (i *OverriddenCTypeConvertible) CanTransferToGlib(transfer TransferOwnership) bool {
	return i.Actual.CanTransferToGlib(transfer)
}

// GetTransferFromGlibFunction implements ConvertibleType.
func (i *OverriddenCTypeConvertible) GetTransferFromGlibFunction(transfer TransferOwnership) string {
	return i.Actual.GetTransferFromGlibFunction(transfer)
}

// GetTransferToGlibFunction implements ConvertibleType.
func (i *OverriddenCTypeConvertible) GetTransferToGlibFunction(transfer TransferOwnership) string {
	return i.Actual.GetTransferToGlibFunction(transfer)
}

// GoUnsafeFromGlibBorrowFunction implements ConvertibleType.
func (i *OverriddenCTypeConvertible) GoUnsafeFromGlibBorrowFunction() string {
	return i.Actual.GoUnsafeFromGlibBorrowFunction()
}

// GoUnsafeFromGlibFullFunction implements ConvertibleType.
func (i *OverriddenCTypeConvertible) GoUnsafeFromGlibFullFunction() string {
	return i.Actual.GoUnsafeFromGlibFullFunction()
}

// GoUnsafeFromGlibNoneFunction implements ConvertibleType.
func (i *OverriddenCTypeConvertible) GoUnsafeFromGlibNoneFunction() string {
	return i.Actual.GoUnsafeFromGlibNoneFunction()
}

// GoUnsafeToGlibFullFunction implements ConvertibleType.
func (i *OverriddenCTypeConvertible) GoUnsafeToGlibFullFunction() string {
	return i.Actual.GoUnsafeToGlibFullFunction()
}

// GoUnsafeToGlibNoneFunction implements ConvertibleType.
func (i *OverriddenCTypeConvertible) GoUnsafeToGlibNoneFunction() string {
	return i.Actual.GoUnsafeToGlibNoneFunction()
}

// CGoType implements Type.
func (i *OverriddenCTypeConvertible) CGoType(pointers int) string {
	return GetPointers(pointers) + i.Cgotype
}

// CType implements Type.
func (i *OverriddenCTypeConvertible) CType(pointers int) string {
	return i.Ctype + GetPointers(pointers)
}

// GIRName implements Type.
func (i *OverriddenCTypeConvertible) GIRName() string {
	return i.Actual.GIRName()
}

// GoType implements Type.
func (i *OverriddenCTypeConvertible) GoType(pointers int) string {
	return i.Actual.GoType(pointers)
}

// GoTypeRequiredImport implements Type.
func (i *OverriddenCTypeConvertible) GoTypeRequiredImport() (alias string, module string) {
	return i.Actual.GoTypeRequiredImport()
}

var _ Type = (*OverriddenCTypeConvertible)(nil)
var _ ConvertibleType = (*OverriddenCTypeConvertible)(nil)
var _ OverriddenCType = (*OverriddenCTypeConvertible)(nil)
