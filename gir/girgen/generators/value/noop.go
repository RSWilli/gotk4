package value

import "fmt"

// NoopConverter will declare the incoming type but will drop it
type NoopConverter struct {
	dir    ConversionDirection
	inName string
	inType string
}

// Conversion implements Converter.
func (n NoopConverter) Conversion() string {
	return fmt.Sprintf("_ = %s // no-op conversion\n", n.inName)
}

// AddImports implements Converter.
func (n NoopConverter) AddImports(importer) {}

// ConversionDirection implements Converter.
func (n NoopConverter) ConversionDirection() ConversionDirection { return n.dir }

// In implements Converter.
func (n NoopConverter) InIdentifier() string { return n.inName }

// InType implements Converter.
func (n NoopConverter) InType() string { return n.inType }

// Out implements Converter.
func (n NoopConverter) OutIdentifier() string { return "_" }

// OutType implements Converter.
func (n NoopConverter) OutType() string { return "struct{}" }

var _ Converter = NoopConverter{}
