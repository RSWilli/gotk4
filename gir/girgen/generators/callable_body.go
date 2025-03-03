package generators

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/generators/value"
)

// CallableBodyGenerator concerns itself with the generation of go function bodies that call a single C function.
// It unifies the generation of the bodies of Constructors, Methods, Virtual Methods and Functions.
//
// It does NOT concern itself with the go method signature, meaning that this wont generate docs, go function signatures
// or return statements. These are expected to be provided by the parent generator
type CallableBodyGenerator struct {
	// CgoFunction is the function that should be called
	CgoFunction string

	// CgoParameters contains the converters as Converters used to generate the cgo function signature.
	//
	// It is not recommended to reorder them, because this will most likely break the type of the called function
	CGoParameters value.ConverterList

	// CGoReturn contains the return param of the cgo function. C can only return one value.
	CGoReturn value.Converter
}

// Generate implements Generator.
func (r *CallableBodyGenerator) Generate(w *file.Writer) {
	if len(r.CGoParameters) > 0 {
		w.GoImport("runtime")
	}

	for _, param := range r.CGoParameters {
		fmt.Fprintf(w.Go(), "\tvar %s %s // out\n", param.OutIdentifier(), param.OutType())
	}

	if r.CGoReturn != nil {
		fmt.Fprintf(w.Go(), "\tvar %s %s // in\n", r.CGoReturn.InIdentifier(), r.CGoReturn.InType())
		fmt.Fprintln(w.Go())
	}

	for _, param := range r.CGoParameters {
		fmt.Fprintln(w.Go(), param.Conversion())
	}

	// func call:
	if r.CGoReturn != nil {
		fmt.Fprintf(w.Go(), "\t%s = %s(%s)\n", r.CGoReturn.InIdentifier(), r.CgoFunction, r.CGoParameters.OutIdentifierList())
	} else {
		fmt.Fprintf(w.Go(), "\t%s(%s)\n", r.CgoFunction, r.CGoParameters.OutIdentifierList())
	}

	for _, param := range r.CGoParameters {
		fmt.Fprintf(w.Go(), "\truntime.KeepAlive(%s)\n", param.InIdentifier())
	}

	fmt.Fprintln(w.Go())

	if r.CGoReturn != nil {
		fmt.Fprintf(w.Go(), "\tvar %s %s // out\n\n", r.CGoReturn.OutIdentifier(), r.CGoReturn.OutType())

		fmt.Fprintln(w.Go(), r.CGoReturn.Conversion())
	}
}

// NewCallableBodyGenerator creates a new CallableBodyGenerator generator. It does not do any type converter lookup, and is only a way
// to generate the bodies of Constructors, Methods, Virtual Methods and Functions in a unified way.
func NewCallableBodyGenerator(function string, params value.ConverterList, _return value.Converter) *CallableBodyGenerator {
	return &CallableBodyGenerator{
		CgoFunction:   function,
		CGoParameters: params,
		CGoReturn:     _return,
	}
}
