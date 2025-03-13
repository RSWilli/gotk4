package typesystem

import (
	"fmt"
	"log"
	"strings"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type Record struct {
	Doc
	BaseType

	Fields []*Field

	IsTypeStructFor Type

	// PrivateGoType is the inner struct that contains the C pointer and gets the finalizer attached
	PrivateGoType string

	// unsafe constructors names:
	GoUnsafeBorrowFunction       string
	GoUnsafeTransferFullFunction string
	GoUnsafeTransferNoneFunction string

	GoUnsafeToGlibNoneMethod string
	GoUnsafeToGlibFullMethod string

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

func DeclareRecord(ctx context, v gir.Record) *Record {
	if ctx.skipType(v) {
		return nil
	}

	if strings.HasSuffix(v.Name, "Private") {
		return nil
	}

	r := &Record{
		Doc:           NewDoc(&v.InfoAttrs, &v.InfoElements),
		PrivateGoType: strcases.UnexportPascal(v.Name),

		GoUnsafeBorrowFunction:       fmt.Sprintf("Unsafe%sFromGlibBorrow", v.Name),
		GoUnsafeTransferNoneFunction: fmt.Sprintf("Unsafe%sFromGlibNone", v.Name),
		GoUnsafeTransferFullFunction: fmt.Sprintf("Unsafe%sFromGlibFull", v.Name),

		GoUnsafeToGlibNoneMethod: fmt.Sprintf("Unsafe%sToGlibNone", v.Name),
		GoUnsafeToGlibFullMethod: fmt.Sprintf("Unsafe%sToGlibFull", v.Name),

		GoUnsafeUnrefFunction: "UnsafeFree",
		CgoUnrefFunction:      "C.free", // replaced below if an unref method is found

		BaseType: BaseType{
			GirName: v.Name,
			GoTyp:   v.Name,
			CGoTyp:  "C." + v.CType,
			CTyp:    v.CType,

			GlibGetTypeFn: v.GLibGetType,
		},
	}

	return r
}

func (r *Record) resolveNested(ctx context, v gir.Record) {
	if v.GLibIsGTypeStructFor != "" {
		classType := ctx.findType(&gir.Type{Name: v.GLibIsGTypeStructFor})

		if classType == nil {
			return
		}

		switch classType.(type) {
		case *Class:
		case *Interface:
		default:
			log.Printf("%s should be a type struct for %s, but %s is %T and not class or interface\n", v.Name, v.GLibIsGTypeStructFor, v.GLibIsGTypeStructFor, classType)
			return
		}

		r.IsTypeStructFor = classType
	}

	for _, v := range v.Functions {
		if t := DeclareFunction(ctx, v); t != nil {
			r.Functions = append(r.Functions, t)
		}
	}

	for _, v := range v.Methods {
		if v.Name == "weak_ref" || v.Name == "weak_unref" {
			continue
		}

		if v.Name == "ref" {
			r.GoUnsafeRefFunction = "UnsafeRef"
			r.CgoRefFunction = "C." + v.CIdentifier
			continue
		}

		if v.Name == "unref" {
			r.GoUnsafeUnrefFunction = "UnsafeUnref"
			r.CgoUnrefFunction = "C." + v.CIdentifier
			continue
		}

		if t := NewMethod(ctx, v); t != nil {
			r.Methods = append(r.Methods, t)
		}
	}

	for _, v := range v.Constructors {
		if t := DeclareConstructor(ctx, r, v); t != nil {
			r.Constructors = append(r.Constructors, t)
		}
	}

	// Disguised means opaque, so we're not supposed to access these fields.
	if !v.Disguised {
		for _, v := range v.Fields {
			if t := NewField(ctx, r, v); t != nil {
				r.Fields = append(r.Fields, t)
			}
		}
	}
}
