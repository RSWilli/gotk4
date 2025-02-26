package value

type PrimitiveConverter struct {
	Direction ConversionDirection
	InIdent   string
	InTyp     string
	OutIdent  string
	OutTyp    string
}

// AddImports implements Converter.
func (p *PrimitiveConverter) AddImports(importer) {}

// Conversion implements Converter.
func (p *PrimitiveConverter) Conversion() string {
	return ""
}

// ConversionDirection implements Converter.
func (p *PrimitiveConverter) ConversionDirection() ConversionDirection {
	return p.Direction
}

// InIdentifier implements Converter.
func (p *PrimitiveConverter) InIdentifier() string {
	return p.InIdent
}

// InType implements Converter.
func (p *PrimitiveConverter) InType() string {
	return p.InTyp
}

// OutIdentifier implements Converter.
func (p *PrimitiveConverter) OutIdentifier() string {
	return p.OutIdent
}

// OutType implements Converter.
func (p *PrimitiveConverter) OutType() string {
	return p.OutTyp
}

var _ Converter = &PrimitiveConverter{}
