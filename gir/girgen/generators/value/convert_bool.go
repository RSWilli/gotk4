package value

import "fmt"

type BooleanConverter struct {
	Direction ConversionDirection
	InIdent   string
	InTyp     string
	OutIdent  string
	OutTyp    string
}

// AddImports implements Converter.
func (b *BooleanConverter) AddImports(importer) {}

// Conversion implements Converter.
func (b *BooleanConverter) Conversion() string {
	return fmt.Sprintf("\tif %s {\n\t\t%s = C.TRUE\n\t}\n", b.InIdent, b.OutIdent)
}

// ConversionDirection implements Converter.
func (b *BooleanConverter) ConversionDirection() ConversionDirection {
	return b.Direction
}

// InIdentifier implements Converter.
func (b *BooleanConverter) InIdentifier() string {
	return b.InIdent
}

// InType implements Converter.
func (b *BooleanConverter) InType() string {
	return b.InTyp
}

// OutIdentifier implements Converter.
func (b *BooleanConverter) OutIdentifier() string {
	return b.OutIdent
}

// OutType implements Converter.
func (b *BooleanConverter) OutType() string {
	return b.OutTyp
}
