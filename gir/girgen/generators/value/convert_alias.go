package value

type AliasConverter struct {
	Direction ConversionDirection
	InIdent   string
	InTyp     string
	OutIdent  string
	OutTyp    string
}

// AddImports implements Converter.
func (c *AliasConverter) AddImports(importer) {}

// Conversion implements Converter.
func (c *AliasConverter) Conversion() string {
	return ""
}

// ConversionDirection implements Converter.
func (c *AliasConverter) ConversionDirection() ConversionDirection {
	return c.Direction
}

// InIdentifier implements Converter.
func (c *AliasConverter) InIdentifier() string {
	return c.InIdent
}

// InType implements Converter.
func (c *AliasConverter) InType() string {
	return c.InTyp
}

// OutIdentifier implements Converter.
func (c *AliasConverter) OutIdentifier() string {
	return c.OutIdent
}

// OutType implements Converter.
func (c *AliasConverter) OutType() string {
	return c.OutTyp
}

var _ Converter = &AliasConverter{}
