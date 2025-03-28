package typesystem

import (
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type Record struct {
	Doc
	BaseType

	gir gir.Record

	Fields []*Field

	// PrivateGoType is the inner struct that contains the C pointer and gets the finalizer attached
	PrivateGoType string

	// unsafe constructors names:
	GoUnsafeFromGlibBorrowFunction string
	GoUnsafeFromGlibFullFunction   string
	GoUnsafeFromGlibNoneFunction   string

	GoUnsafeToGlibNoneFunction string
	GoUnsafeToGlibFullFunction string

	GoUnsafeRefFunction string
	CgoRefFunction      string

	GoUnsafeUnrefFunction string
	CgoUnrefFunction      string

	Functions    []*CallableSignature
	Methods      []*CallableSignature
	Constructors []*CallableSignature

	// TODO:
	Unions     []*Union
	Properties []*struct{}
}

func DeclareRecord(e *env, v gir.Record) *Record {
	if !v.IsIntrospectable() {
		return nil
	}

	if e.skip(nil, v) {
		return nil
	}

	if strings.HasSuffix(v.Name, "Private") {
		return nil
	}

	return &Record{
		Doc:           NewDoc(&v.InfoAttrs, &v.InfoElements),
		PrivateGoType: strcases.UnexportPascal(v.Name),

		GoUnsafeFromGlibBorrowFunction: fmt.Sprintf("Unsafe%sFromGlibBorrow", v.Name),
		GoUnsafeFromGlibNoneFunction:   fmt.Sprintf("Unsafe%sFromGlibNone", v.Name),
		GoUnsafeFromGlibFullFunction:   fmt.Sprintf("Unsafe%sFromGlibFull", v.Name),

		GoUnsafeToGlibNoneFunction: fmt.Sprintf("Unsafe%sToGlibNone", v.Name),
		GoUnsafeToGlibFullFunction: fmt.Sprintf("Unsafe%sToGlibFull", v.Name),

		GoUnsafeUnrefFunction: fmt.Sprintf("Unsafe%sFree", v.Name),
		CgoUnrefFunction:      "C.free", // replaced below if an unref or custom free method is found

		BaseType: BaseType{
			GirName: v.Name,
			GoTyp:   v.Name,
			CGoTyp:  "C." + v.CType,
			CTyp:    v.CType,

			GlibGetTypeFn: v.GLibGetType,
		},
		gir: v,
	}
}

func (r *Record) declareNested(e *env) {

	e = e.sub("record", r.gir.Name)

	for _, v := range r.gir.Functions {
		if t := DeclareFunction(e, r, v); t != nil {
			r.Functions = append(r.Functions, t)
		}
	}

	for _, v := range r.gir.Methods {
		if v.Name == "weak_ref" || v.Name == "weak_unref" {
			// we don't want the user to be able to weakly reference the object
			// as there are better tools for this and this will only cause problems
			continue
		}

		// see https://github.com/gtk-rs/gir/blob/87cddb70c739f25edd8047e6780e3934af8ff474/src/library.rs#L459-L481

		if v.Name == "ref" {
			r.GoUnsafeRefFunction = fmt.Sprintf("Unsafe%sRef", v.Name)
			r.CgoRefFunction = "C." + v.CIdentifier
			continue
		}

		if v.Name == "unref" {
			r.GoUnsafeUnrefFunction = fmt.Sprintf("Unsafe%sUnref", v.Name)
			r.CgoUnrefFunction = "C." + v.CIdentifier
			continue
		}

		if v.Name == "copy" {
			// r.GoUnsafeRefFunction = "UnsafeCopy"
			// TODO: how to handle this? Maybe add a "ref mode" to the struct?
			continue
		}

		if v.Name == "copy_into" {
			// TODO: how to handle this? It sounds like we need to allocate a copy struct before
			// beeing able to copy into it
			continue
		}

		if v.Name == "free" {
			r.GoUnsafeUnrefFunction = fmt.Sprintf("Unsafe%sFree", v.Name)
			r.CgoUnrefFunction = "C." + v.CIdentifier
			continue
		}

		if v.Name == "destroy" {
			r.GoUnsafeUnrefFunction = fmt.Sprintf("Unsafe%sDestroy", v.Name)
			r.CgoUnrefFunction = "C." + v.CIdentifier
			continue
		}

		if t := NewMethod(e, r, v); t != nil {
			r.Methods = append(r.Methods, t)
		}
	}

	for _, v := range r.gir.Constructors {
		if t := DeclareConstructor(e, r, v); t != nil {
			r.Constructors = append(r.Constructors, t)
		}
	}

	// Disguised means opaque, so we're not supposed to access these fields.
	if !r.gir.Disguised {
		for _, v := range r.gir.Fields {
			if t := NewField(e, r, v); t != nil {
				r.Fields = append(r.Fields, t)
			}
		}
	}
}
