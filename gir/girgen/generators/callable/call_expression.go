package callable

import (
	"strings"

	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type CallExpressionList []CallExpression

// CallExpression is responsible for generating the correct call expression required for the c call
//
// this goes hand-in-hand with the way out parameters are converted
type CallExpression struct {
	Param *typesystem.Param
}

func (ce CallExpression) CallExpression() string {
	if ce.Param.Direction == "out" {
		return "&" + ce.Param.CName
	}

	return ce.Param.CName
}

func (l CallExpressionList) Call() string {
	var expressions []string

	for _, c := range l {
		expressions = append(expressions, c.CallExpression())
	}

	return strings.Join(expressions, ", ")
}
