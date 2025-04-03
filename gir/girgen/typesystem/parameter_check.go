package typesystem

type checkedParameterType interface {
	allowedTypeForParam(P *Param) bool
}
