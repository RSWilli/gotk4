package typesystem

type Type interface {
	GIRName() string
	GoType(pointers int) string
	CGoType(pointers int) string
	CType(pointers int) string

	pointersAllowed(pointers int) bool
}

// BaseType partially implements the [Type] interface
type BaseType struct {
	GirName string
	GoTyp   string
	CGoTyp  string
	CTyp    string

	IsGoBuiltin bool
}

// GIRName implements Type.
func (b BaseType) GIRName() string {
	return b.GirName
}

// CGoType implements Type.
func (b BaseType) CGoType(pointers int) string {
	return GetPointers(pointers) + b.CGoTyp
}

// CType implements Type.
func (b BaseType) CType(pointers int) string {
	return b.CTyp + GetPointers(pointers)
}

// GoType implements Type.
func (b BaseType) GoType(pointers int) string {
	if b.IsGoBuiltin {
		return b.GoTyp
	}
	return GetPointers(pointers) + b.GoTyp
}
