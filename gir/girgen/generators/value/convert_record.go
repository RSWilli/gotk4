package value

import (
	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/gencontext"
)

type RecordConverter struct {
	Direction ConversionDirection
	InIdent   string
	InTyp     string
	OutIdent  string
	OutTyp    string
}

// AddImports implements Converter.
func (r *RecordConverter) AddImports(importer) {}

// Conversion implements Converter.
func (r *RecordConverter) Conversion() string {
	return ""
}

// ConversionDirection implements Converter.
func (r *RecordConverter) ConversionDirection() ConversionDirection {
	return r.Direction
}

// InIdentifier implements Converter.
func (r *RecordConverter) InIdentifier() string {
	return r.InIdent
}

// InType implements Converter.
func (r *RecordConverter) InType() string {
	return r.InTyp
}

// OutIdentifier implements Converter.
func (r *RecordConverter) OutIdentifier() string {
	return r.OutIdent
}

// OutType implements Converter.
func (r *RecordConverter) OutType() string {
	return r.OutTyp
}

var _ Converter = &RecordConverter{}

func NewRecordConverter(ctx gencontext.GenerationContext, r *gir.Record, direction ConversionDirection, inIdent string, inTyp string, outIdent string, outTyp string) *RecordConverter {
	return &RecordConverter{
		Direction: direction,
		InIdent:   inIdent,
		InTyp:     inTyp,
		OutIdent:  outIdent,
		OutTyp:    outTyp,
	}
}
