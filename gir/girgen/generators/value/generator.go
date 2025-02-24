package value

import "strings"

// Generator writes the parameter or return value conversion into sections
type Generator interface {
	Generate(importer, FunctionCallSections)
}

// Converter descibes a cgo<->go value conversion.
//
// The names of the accessors are derived from https://go.dev/ref/spec#Function_types The Converter itself
// should return empty strings when not applicable
type Converter interface {
	Generator
	// CGoParameterDecl returns the declaration of the cgo variable, e.g. `attr1 *C.GObject`
	CGoParameterDecl() string
	// CGoParameterIdentifier returns the identifier of the cgo variable, e.g. `attr1`
	CGoParameterIdentifier() string

	// CGoReturnDecl returns the declaration of the cgo return variable, e.g. `cret C.gboolean`
	CGoReturnDecl() string
	// CGoReturnIdentifier returns the identifier of the cgo variable, e.g. `cret`
	CGoReturnIdentifier() string

	// GoParameterDecl returns the declaration of the go variable, e.g. `attr1 *coreglib.Object`
	GoParameterDecl() string
	// GoParameterIdentifier returns the identifier of the go variable, e.g. `attr1`
	GoParameterIdentifier() string

	// GoReturnDecl returns the declaration of the go return variable, e.g. `ok bool`
	GoReturnDecl() string
	// GoReturnIdentifier returns the identifier of the go variable, e.g. `ok`
	GoReturnIdentifier() string
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

var _ Generator = ConverterList{}

func (list ConverterList) Generate(f importer, s FunctionCallSections) {
	for _, g := range list {
		if g == nil {
			continue
		}

		g.Generate(f, s)
	}
}

// CGoParameterDeclList returns all cgo parameter declarations comma separated
func (list ConverterList) CGoParameterDeclList() string {
	return list.makeCommaSeparated((Converter).CGoParameterDecl)
}

// CGoParameterIdentifierList returns all cgo parameter identifiers comma separated
func (list ConverterList) CGoParameterIdentifierList() string {
	return list.makeCommaSeparated((Converter).CGoParameterIdentifier)
}

// CGoReturnDeclList returns all cgo parameter declarations comma separated
func (list ConverterList) CGoReturnDeclList() string {
	return list.makeCommaSeparated((Converter).CGoReturnDecl)
}

// CGoReturnIdentifierList returns all cgo parameter identifiers comma separated
func (list ConverterList) CGoReturnIdentifierList() string {
	return list.makeCommaSeparated((Converter).CGoReturnIdentifier)
}

// GoParameterDeclList returns all Go parameter declarations comma separated
func (list ConverterList) GoParameterDeclList() string {
	return list.makeCommaSeparated((Converter).GoParameterDecl)
}

// GoParameterIdentifierList returns all Go parameter identifiers comma separated
func (list ConverterList) GoParameterIdentifierList() string {
	return list.makeCommaSeparated((Converter).GoParameterIdentifier)
}

// GoReturnDeclList returns all Go return declarations comma separated and bracketed if needed
func (list ConverterList) GoReturnDeclList() string {
	return list.makeCommaSeparatedAutoBracket((Converter).GoReturnDecl)
}

// GoReturnIdentifierList returns all go return identifiers comma separated
func (list ConverterList) GoReturnIdentifierList() string {
	return list.makeCommaSeparated((Converter).GoReturnIdentifier)
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

// makeCommaSeparatedAutoBracket returns a comma separated list of the non empty strings returned by the accessor function, and
// encloses <list> like " (<list>)" if the list contained more than one element. It adds a leading space if not empty.
func (list ConverterList) makeCommaSeparatedAutoBracket(f func(Converter) string) string {
	var parts []string

	for _, v := range list {
		s := f(v)

		if s == "" {
			continue
		}

		parts = append(parts, s)
	}

	if len(parts) == 0 {
		return ""
	}

	if len(parts) > 1 {

		return " (" + strings.Join(parts, ", ") + ")"
	}

	return " " + strings.Join(parts, ", ")
}
