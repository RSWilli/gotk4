package value

type CallbackConverter struct {
	Direction ConversionDirection
	InIdent   string
	InTyp     string
	OutIdent  string
	OutTyp    string
}

// AddImports implements Converter.
func (b *CallbackConverter) AddImports(importer) {}

// Conversion implements Converter.
func (b *CallbackConverter) Conversion() string {
	return ""
}

// ConversionDirection implements Converter.
func (b *CallbackConverter) ConversionDirection() ConversionDirection {
	return b.Direction
}

// InIdentifier implements Converter.
func (b *CallbackConverter) InIdentifier() string {
	return b.InIdent
}

// InType implements Converter.
func (b *CallbackConverter) InType() string {
	return b.InTyp
}

// OutIdentifier implements Converter.
func (b *CallbackConverter) OutIdentifier() string {
	return b.OutIdent
}

// OutType implements Converter.
func (b *CallbackConverter) OutType() string {
	return b.OutTyp
}
