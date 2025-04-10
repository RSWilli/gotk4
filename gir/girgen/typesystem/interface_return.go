package typesystem

import "github.com/diamondburned/gotk4/gir"

// InterfaceReturn is a type that wraps an interface type if the return type is a pointer
// and an interface. This is needed if the Ctype does not match the interface's Ctype.
type InterfaceReturn struct {
	Interface *Interface

	Ctype   string
	Cgotype string
}

// CanTransferFromGlib implements ConvertibleType.
func (i *InterfaceReturn) CanTransferFromGlib(transfer TransferOwnership) bool {
	return i.Interface.CanTransferFromGlib(transfer)
}

// CanTransferToGlib implements ConvertibleType.
func (i *InterfaceReturn) CanTransferToGlib(transfer TransferOwnership) bool {
	return i.Interface.CanTransferToGlib(transfer)
}

// GetTransferFromGlibFunction implements ConvertibleType.
func (i *InterfaceReturn) GetTransferFromGlibFunction(transfer TransferOwnership) string {
	return i.Interface.GetTransferFromGlibFunction(transfer)
}

// GetTransferToGlibFunction implements ConvertibleType.
func (i *InterfaceReturn) GetTransferToGlibFunction(transfer TransferOwnership) string {
	return i.Interface.GetTransferToGlibFunction(transfer)
}

// GoUnsafeFromGlibBorrowFunction implements ConvertibleType.
func (i *InterfaceReturn) GoUnsafeFromGlibBorrowFunction() string {
	return i.Interface.GoUnsafeFromGlibBorrowFunction()
}

// GoUnsafeFromGlibFullFunction implements ConvertibleType.
func (i *InterfaceReturn) GoUnsafeFromGlibFullFunction() string {
	return i.Interface.GoUnsafeFromGlibFullFunction()
}

// GoUnsafeFromGlibNoneFunction implements ConvertibleType.
func (i *InterfaceReturn) GoUnsafeFromGlibNoneFunction() string {
	return i.Interface.GoUnsafeFromGlibNoneFunction()
}

// GoUnsafeToGlibFullFunction implements ConvertibleType.
func (i *InterfaceReturn) GoUnsafeToGlibFullFunction() string {
	return i.Interface.GoUnsafeToGlibFullFunction()
}

// GoUnsafeToGlibNoneFunction implements ConvertibleType.
func (i *InterfaceReturn) GoUnsafeToGlibNoneFunction() string {
	return i.Interface.GoUnsafeToGlibNoneFunction()
}

// CGoType implements Type.
func (i *InterfaceReturn) CGoType(pointers int) string {
	return GetPointers(pointers) + i.Cgotype
}

// CType implements Type.
func (i *InterfaceReturn) CType(pointers int) string {
	return i.Ctype + GetPointers(pointers)
}

// GIRName implements Type.
func (i *InterfaceReturn) GIRName() string {
	return i.Interface.GIRName()
}

// GoType implements Type.
func (i *InterfaceReturn) GoType(pointers int) string {
	return i.Interface.GoType(pointers)
}

// GoTypeRequiredImport implements Type.
func (i *InterfaceReturn) GoTypeRequiredImport() (alias string, module string) {
	return i.Interface.GoTypeRequiredImport()
}

var _ Type = (*InterfaceReturn)(nil)
var _ ConvertibleType = (*InterfaceReturn)(nil)

func wrapInterfaceReturnIfNeeded(resolved Type, requested gir.AnyType) Type {
	inter, ok := resolved.(*Interface)

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

	return &InterfaceReturn{
		Interface: inter,
		Ctype:     ctype,
		Cgotype:   "C." + ctype,
	}
}
