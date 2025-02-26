package value

type BitfieldConverter struct {
	Direction ConversionDirection
	InIdent   string
	InTyp     string
	OutIdent  string
	OutTyp    string
}

// AddImports implements Converter.
func (c *BitfieldConverter) AddImports(importer) {}

// Conversion implements Converter.
func (c *BitfieldConverter) Conversion() string {
	return ""
}

// ConversionDirection implements Converter.
func (c *BitfieldConverter) ConversionDirection() ConversionDirection {
	return c.Direction
}

// InIdentifier implements Converter.
func (c *BitfieldConverter) InIdentifier() string {
	return c.InIdent
}

// InType implements Converter.
func (c *BitfieldConverter) InType() string {
	return c.InTyp
}

// OutIdentifier implements Converter.
func (c *BitfieldConverter) OutIdentifier() string {
	return c.OutIdent
}

// OutType implements Converter.
func (c *BitfieldConverter) OutType() string {
	return c.OutTyp
}

var _ Converter = &BitfieldConverter{}
