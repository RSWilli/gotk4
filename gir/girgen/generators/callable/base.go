package callable

import (
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

// BaseConverter wraps a [typesystem.Param] and partially implements [Converter]. This is useful for the
// converters that don't need to do anything special with the identifiers/expressions
type BaseConverter struct {
	Param *typesystem.Param
}

// CGoIdentifier implements Converter.
func (p BaseConverter) CGoIdentifier() string {
	return p.Param.CName
}

// CGoType implements Converter.
func (p BaseConverter) CGoType() string {
	return p.Param.Type.CGoType()
}

// CallExpression implements Converter.
func (p BaseConverter) CallExpression() string {
	return p.CGoIdentifier()
}

// GoIdentifier implements Converter.
func (p BaseConverter) GoIdentifier() string {
	return p.Param.GoName
}

// GoType implements Converter.
func (p BaseConverter) GoType() string {
	return p.Param.Type.GoType()
}
