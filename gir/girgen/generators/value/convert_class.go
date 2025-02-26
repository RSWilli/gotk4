package value

type ClassConverter struct {
	Direction ConversionDirection
	InIdent   string
	InTyp     string
	OutIdent  string
	OutTyp    string
}

// AddImports implements Converter.
func (c *ClassConverter) AddImports(importer) {}

// Conversion implements Converter.
func (c *ClassConverter) Conversion() string {
	return ""
}

// ConversionDirection implements Converter.
func (c *ClassConverter) ConversionDirection() ConversionDirection {
	return c.Direction
}

// InIdentifier implements Converter.
func (c *ClassConverter) InIdentifier() string {
	return c.InIdent
}

// InType implements Converter.
func (c *ClassConverter) InType() string {
	return c.InTyp
}

// OutIdentifier implements Converter.
func (c *ClassConverter) OutIdentifier() string {
	return c.OutIdent
}

// OutType implements Converter.
func (c *ClassConverter) OutType() string {
	return c.OutTyp
}

var _ Converter = &ClassConverter{}
