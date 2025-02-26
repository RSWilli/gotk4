package value

import (
	"fmt"
	"strings"
)

// ConversionDirection is the conversion direction between Go and C.
type ConversionDirection uint8

const (
	_ ConversionDirection = iota
	ConvertGoToC
	ConvertCToGo
)

func (cd ConversionDirection) Switch() ConversionDirection {
	switch cd {
	case ConvertCToGo:
		return ConvertGoToC
	case ConvertGoToC:
		return ConvertCToGo
	default:
		panic(fmt.Sprintf("unexpected value.ConversionDirection: %#v", cd))
	}
}

// Converter descibes a cgo<->go value conversion.
type Converter interface {
	AddImports(importer)
	ConversionDirection() ConversionDirection
	InIdentifier() string
	InType() string
	OutIdentifier() string
	OutType() string
	Conversion() string
}

// importer is used by the [Generator] to allow it to import packages and C headers
type importer interface {
	CInclude(header string)
	GoImportCore(pkg string)
	GoImportCoreGlib()
	GoImport(pkg string)
	GoImportAnonymous(pkg string)
	GoImportAliased(pkg string, alias string)
}

// ConverterList is a utility type to ease the use of multiple ValueConverters.
type ConverterList []Converter

func (list ConverterList) AddImports(f importer) {
	for _, g := range list {
		if g == nil {
			continue
		}

		g.AddImports(f)
	}
}

// CDeclList returns the c declarations comma separated
func (list ConverterList) CDeclList() string {
	return list.makeCommaSeparated(func(c Converter) string {
		var ident, typ string
		if c.ConversionDirection() == ConvertCToGo {
			ident = c.InIdentifier()
			typ = c.InType()
		} else {
			ident = c.OutIdentifier()
			typ = c.OutType()
		}
		if ident == "" || typ == "" {
			return ""
		}
		return fmt.Sprintf("%s %s", ident, typ)
	})
}

// CDeclList returns the c declarations comma separated
func (list ConverterList) GoDeclList() string {
	return list.makeCommaSeparated(func(c Converter) string {
		var ident, typ string
		if c.ConversionDirection() == ConvertGoToC {
			ident = c.InIdentifier()
			typ = c.InType()
		} else {
			ident = c.OutIdentifier()
			typ = c.OutType()
		}
		if ident == "" || typ == "" {
			return ""
		}
		return fmt.Sprintf("%s %s", ident, typ)
	})
}

// InIdentifierList returns all "In" identifiers comma separated
func (list ConverterList) InIdentifierList() string {
	return list.makeCommaSeparated((Converter).InIdentifier)
}

// OutIdentifierList returns all "Out" identifiers comma separated
func (list ConverterList) OutIdentifierList() string {
	return list.makeCommaSeparated((Converter).OutIdentifier)
}

// makeCommaSeparated returns a comma separated list of the non empty strings returned by the accessor function
func (list ConverterList) makeCommaSeparated(f func(Converter) string) string {
	var parts []string

	for _, v := range list {
		s := f(v)

		if s == "" {
			continue
		}

		parts = append(parts, s)
	}

	return strings.Join(parts, ", ")
}
