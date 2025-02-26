package value

type EnumConverter struct {
	Direction ConversionDirection
	InIdent   string
	InTyp     string
	OutIdent  string
	OutTyp    string
}

// AddImports implements Converter.
func (c *EnumConverter) AddImports(importer) {}

// Conversion implements Converter.
func (c *EnumConverter) Conversion() string {
	return ""
}

// ConversionDirection implements Converter.
func (c *EnumConverter) ConversionDirection() ConversionDirection {
	return c.Direction
}

// InIdentifier implements Converter.
func (c *EnumConverter) InIdentifier() string {
	return c.InIdent
}

// InType implements Converter.
func (c *EnumConverter) InType() string {
	return c.InTyp
}

// OutIdentifier implements Converter.
func (c *EnumConverter) OutIdentifier() string {
	return c.OutIdent
}

// OutType implements Converter.
func (c *EnumConverter) OutType() string {
	return c.OutTyp
}

var _ Converter = &EnumConverter{}
