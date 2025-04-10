package typesystem

import "github.com/diamondburned/gotk4/gir"

// ClassReturn is a type that wraps a Class type if the return type is a pointer
// and differs from the classes ctype. This is needed if the Ctype does not match the Class's Ctype.
type ClassReturn struct {
	Class *Class

	Ctype   string
	Cgotype string
}

// CanTransferFromGlib implements ConvertibleType.
func (i *ClassReturn) CanTransferFromGlib(transfer TransferOwnership) bool {
	return i.Class.CanTransferFromGlib(transfer)
}

// CanTransferToGlib implements ConvertibleType.
func (i *ClassReturn) CanTransferToGlib(transfer TransferOwnership) bool {
	return i.Class.CanTransferToGlib(transfer)
}

// GetTransferFromGlibFunction implements ConvertibleType.
func (i *ClassReturn) GetTransferFromGlibFunction(transfer TransferOwnership) string {
	return i.Class.GetTransferFromGlibFunction(transfer)
}

// GetTransferToGlibFunction implements ConvertibleType.
func (i *ClassReturn) GetTransferToGlibFunction(transfer TransferOwnership) string {
	return i.Class.GetTransferToGlibFunction(transfer)
}

// GoUnsafeFromGlibBorrowFunction implements ConvertibleType.
func (i *ClassReturn) GoUnsafeFromGlibBorrowFunction() string {
	return i.Class.GoUnsafeFromGlibBorrowFunction()
}

// GoUnsafeFromGlibFullFunction implements ConvertibleType.
func (i *ClassReturn) GoUnsafeFromGlibFullFunction() string {
	return i.Class.GoUnsafeFromGlibFullFunction()
}

// GoUnsafeFromGlibNoneFunction implements ConvertibleType.
func (i *ClassReturn) GoUnsafeFromGlibNoneFunction() string {
	return i.Class.GoUnsafeFromGlibNoneFunction()
}

// GoUnsafeToGlibFullFunction implements ConvertibleType.
func (i *ClassReturn) GoUnsafeToGlibFullFunction() string {
	return i.Class.GoUnsafeToGlibFullFunction()
}

// GoUnsafeToGlibNoneFunction implements ConvertibleType.
func (i *ClassReturn) GoUnsafeToGlibNoneFunction() string {
	return i.Class.GoUnsafeToGlibNoneFunction()
}

// CGoType implements Type.
func (i *ClassReturn) CGoType(pointers int) string {
	return GetPointers(pointers) + i.Cgotype
}

// CType implements Type.
func (i *ClassReturn) CType(pointers int) string {
	return i.Ctype + GetPointers(pointers)
}

// GIRName implements Type.
func (i *ClassReturn) GIRName() string {
	return i.Class.GIRName()
}

// GoType implements Type.
func (i *ClassReturn) GoType(pointers int) string {
	return i.Class.GoType(pointers)
}

// GoTypeRequiredImport implements Type.
func (i *ClassReturn) GoTypeRequiredImport() (alias string, module string) {
	return i.Class.GoTypeRequiredImport()
}

var _ Type = (*ClassReturn)(nil)
var _ ConvertibleType = (*ClassReturn)(nil)

func wrapClassReturnIfNeeded(resolved Type, requested gir.AnyType) Type {
	inter, ok := resolved.(*Class)

	if !ok {
		return resolved
	}

	if requested.Type == nil {
		panic("requested type is nil")
	}

	req := requested.Type

	ctype := trimCTypePointers(req.CType)

	if ctype == inter.CType(0) {
		return resolved
	}

	return &ClassReturn{
		Class:   inter,
		Ctype:   ctype,
		Cgotype: "C." + ctype,
	}
}
