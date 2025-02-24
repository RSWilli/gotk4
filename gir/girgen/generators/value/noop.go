package value

type NoopConverter struct{}

// CGoReturnDecl implements ValueConverter.
func (NoopConverter) CGoReturnDecl() string { return "" }

// CGoReturnIdentifier implements ValueConverter.
func (NoopConverter) CGoReturnIdentifier() string { return "" }

// GoReturnDecl implements ValueConverter.
func (NoopConverter) GoReturnDecl() string { return "" }

// GoReturnIdentifier implements ValueConverter.
func (NoopConverter) GoReturnIdentifier() string { return "" }

var _ Converter = NoopConverter{}

func (NoopConverter) Generate(importer, FunctionCallSections) {}

// CGoParameterDecl implements ValueConverter.
func (NoopConverter) CGoParameterDecl() string { return "" }

// CGoParameterIdentifier implements ValueConverter.
func (NoopConverter) CGoParameterIdentifier() string { return "" }

// GoParameterDecl implements ValueConverter.
func (NoopConverter) GoParameterDecl() string { return "" }

// GoParameterIdentifier implements ValueConverter.
func (NoopConverter) GoParameterIdentifier() string { return "" }
