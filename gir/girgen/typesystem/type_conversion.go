package typesystem

import "fmt"

type ConvertFromGlibFullType interface {
	GoUnsafeFromGlibFullFunction() string
}

type ConvertFromGlibNoneType interface {
	GoUnsafeFromGlibNoneFunction() string
}

type ConvertFromGlibBorrowType interface {
	GoUnsafeFromGlibBorrowFunction() string
}

type ConvertToGlibFullType interface {
	GoUnsafeToGlibFullFunction() string
}

type ConvertToGlibNoneType interface {
	GoUnsafeToGlibNoneFunction() string
}

type allConversionsType interface {
	ConvertFromGlibFullType
	ConvertFromGlibNoneType
	ConvertFromGlibBorrowType
	ConvertToGlibFullType
	ConvertToGlibNoneType
}

type BaseConversions struct {
	FromGlibBorrowFunction string
	FromGlibFullFunction   string
	FromGlibNoneFunction   string
	ToGlibNoneFunction     string
	ToGlibFullFunction     string
}

func newDefaultBaseConversions(name string) BaseConversions {
	return BaseConversions{
		FromGlibBorrowFunction: fmt.Sprintf("Unsafe%sFromGlibBorrow", name),
		FromGlibNoneFunction:   fmt.Sprintf("Unsafe%sFromGlibNone", name),
		FromGlibFullFunction:   fmt.Sprintf("Unsafe%sFromGlibFull", name),

		ToGlibNoneFunction: fmt.Sprintf("Unsafe%sToGlibNone", name),
		ToGlibFullFunction: fmt.Sprintf("Unsafe%sToGlibFull", name),
	}
}

// GoUnsafeFromGlibBorrowFunction implements ConvertibleType.
func (b BaseConversions) GoUnsafeFromGlibBorrowFunction() string {
	return b.FromGlibBorrowFunction
}

// GoUnsafeFromGlibFullFunction implements ConvertibleType.
func (b BaseConversions) GoUnsafeFromGlibFullFunction() string {
	return b.FromGlibFullFunction
}

// GoUnsafeFromGlibNoneFunction implements ConvertibleType.
func (b BaseConversions) GoUnsafeFromGlibNoneFunction() string {
	return b.FromGlibNoneFunction
}

// GoUnsafeToGlibFullFunction implements ConvertibleType.
func (b BaseConversions) GoUnsafeToGlibFullFunction() string {
	return b.ToGlibFullFunction
}

// GoUnsafeToGlibNoneFunction implements ConvertibleType.
func (b BaseConversions) GoUnsafeToGlibNoneFunction() string {
	return b.ToGlibNoneFunction
}

var _ allConversionsType = BaseConversions{}
