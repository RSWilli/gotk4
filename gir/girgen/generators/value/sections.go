package value

import (
	"bytes"
	"io"
	"strings"
)

type FunctionCallSections interface {
	io.WriterTo
	Preamble() io.Writer
	InputDecl() io.Writer
	InputConv() io.Writer
	FnCall() io.Writer
	OutputDecl() io.Writer
	OutputConv() io.Writer
	Return() io.Writer
}

func NewFunctionCallSections() FunctionCallSections {
	return &functionCallSections{}
}

type functionCallSections struct {
	secPreamble   bytes.Buffer
	secInputDecl  bytes.Buffer
	secInputConv  bytes.Buffer
	secFnCall     bytes.Buffer
	secOutputDecl bytes.Buffer
	secOutputConv bytes.Buffer
	secReturn     bytes.Buffer
}

// FnCall implements FunctionCallSections.
func (s *functionCallSections) Preamble() io.Writer {
	return &s.secPreamble
}

// FnCall implements FunctionCallSections.
func (s *functionCallSections) FnCall() io.Writer {
	return &s.secFnCall
}

// InputConv implements FunctionCallSections.
func (s *functionCallSections) InputConv() io.Writer {
	return &s.secInputConv
}

// InputDecl implements FunctionCallSections.
func (s *functionCallSections) InputDecl() io.Writer {
	return &s.secInputDecl
}

// OutputConv implements FunctionCallSections.
func (s *functionCallSections) OutputConv() io.Writer {
	return &s.secOutputConv
}

// OutputDecl implements FunctionCallSections.
func (s *functionCallSections) OutputDecl() io.Writer {
	return &s.secOutputDecl
}

// Return implements FunctionCallSections.
func (s *functionCallSections) Return() io.Writer {
	return &s.secReturn
}

var _ FunctionCallSections = &functionCallSections{}

func (s *functionCallSections) WriteTo(w io.Writer) (int64, error) {
	r := io.MultiReader(
		&s.secInputDecl,
		strings.NewReader("\n"),
		&s.secInputConv,
		strings.NewReader("\n"),
		&s.secFnCall,
		strings.NewReader("\n"),
		&s.secOutputDecl,
		strings.NewReader("\n"),
		&s.secOutputConv,
		strings.NewReader("\n"),
		&s.secReturn,
	)

	return io.Copy(w, r)
}
