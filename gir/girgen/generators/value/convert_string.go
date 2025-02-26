package value

type StringConverter struct {
	Direction ConversionDirection
	InIdent   string
	InTyp     string
	OutIdent  string
	OutTyp    string
}

// AddImports implements Converter.
func (c *StringConverter) AddImports(importer) {}

// Conversion implements Converter.
func (c *StringConverter) Conversion() string {
	return ""
}

// ConversionDirection implements Converter.
func (c *StringConverter) ConversionDirection() ConversionDirection {
	return c.Direction
}

// InIdentifier implements Converter.
func (c *StringConverter) InIdentifier() string {
	return c.InIdent
}

// InType implements Converter.
func (c *StringConverter) InType() string {
	return c.InTyp
}

// OutIdentifier implements Converter.
func (c *StringConverter) OutIdentifier() string {
	return c.OutIdent
}

// OutType implements Converter.
func (c *StringConverter) OutType() string {
	return c.OutTyp
}

var _ Converter = &StringConverter{}
