package typesystem

type Manual struct {
	BaseType

	GoImportAlias  string
	GoImportModule string

	FromGLibTransferFull   string
	FromGLibTransferBorrow string
	FromGLibTransferNone   string

	ToGLibTransferFull   string
	ToGLibTransferBorrow string
	ToGLibTransferNone   string
}
