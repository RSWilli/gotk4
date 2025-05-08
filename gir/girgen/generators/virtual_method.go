package generators

import (
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/generators/convert"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type VirtualMethodGenerator struct {
	Doc SubGenerator
	*typesystem.VirtualMethod

	// ParamConverters contains all the c->go in param converters needed for the go call
	ParamConverters convert.ConverterList

	// ReturnConverters contains all the go->c converters needed for the returns and out params
	ReturnConverters convert.ConverterList
}

// Generate implements VirtualMethodGenerator.
func (v *VirtualMethodGenerator) Generate(w *file.Package) {
	v.generateExport(w)
	v.generateParentCall(w)
}

func (v *VirtualMethodGenerator) generateParentCall(w *file.Package) {
	v.generateParentCPreamble(w)
}

func (v *VirtualMethodGenerator) generateParentCPreamble(w *file.Package) {

	var paramDecls []string
	// args is the list of arguments needed to call the casted function pointer
	var args []string

	// paramTypes is the list of arguments in the function pointer
	var paramTypes []string

	// add the function pointer as the first parameter which we will need to cast
	paramDecls = append(paramDecls, "void* fnptr")

	for _, p := range v.VirtualMethod.CParameters() {
		ctype := p.CType()
		if p.Direction == "out" {
			ctype += "*" // add the trimmed pointer back
		}

		args = append(args, p.CName)
		paramTypes = append(paramTypes, ctype)
		paramDecls = append(
			paramDecls,
			fmt.Sprintf("%s %s", ctype, p.CName),
		)
	}

	fmt.Fprintf(w.C(), "%s %s(%s) {\n", v.VirtualMethod.CReturn.CType(), v.VirtualMethod.ParentTrampolineName, strings.Join(paramDecls, ", "))
	w.C().Indent()
	fmt.Fprintf(w.C(), "return ((%s (*) (%s))(fnptr))(%s);\n", v.CReturn.CType(), strings.Join(paramTypes, ", "), strings.Join(args, ", "))
	w.C().Unindent()
	fmt.Fprintf(w.C(), "}\n")
}

// generateExport generates the exported Cgo trampoline function for the virtual method. We do not do much here, because we do not have the necessary
// type information of the go subclass to call the function directly. Instead we lookup a function pointer that accepts the same paremeters as the
// trampoline, which handles all the necessary conversions and calls the go function pointer. See [VirtualMethodGenerator.generateTrampoline] for more information.
func (c *VirtualMethodGenerator) generateExport(pkg *file.Package) {
	w := &pkg.Exported

	w.GoImportCore("classdata")
	w.GoImport("unsafe")

	fmt.Fprintf(w.Go(), "//export %s\n", c.TrampolineName)

	fmt.Fprintf(w.Go(), "%s {\n", c.CGoTrampolineSignature())

	w.Go().Indent()

	fmt.Fprintf(w.Go(), "var fn func%s\n", c.CGoTrampolineTail())
	fmt.Fprintf(w.Go(), "{\n")
	w.Go().Indent()
	fmt.Fprintf(w.Go(), "fn = classdata.LoadVirtualMethodFromInstance(unsafe.Pointer(%s), \"%s\").(func%s)\n", c.InstanceParam.CName, c.TrampolineName, c.CGoTrampolineTail())

	fmt.Fprintf(w.Go(), "if fn == nil {\n")
	w.Go().Indent()
	fmt.Fprintf(w.Go(), "panic(\"%s: no function pointer found\")\n", c.TrampolineName)
	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n")
	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "}\n")

	if c.CReturn != nil && c.CReturn.Type.Type != typesystem.Void {
		fmt.Fprintf(w.Go(), "return fn(%s)\n", c.CParameters().CIdentifiers())
	} else {
		fmt.Fprintf(w.Go(), "fn(%s)\n", c.CParameters().CIdentifiers())
	}

	w.Go().Unindent()

	fmt.Fprintln(w.Go(), "}")
	fmt.Fprintln(w.Go())
}

// generateTrampoline is called from the overrides generator to generate the virtual function in a place where we have enough type information
// to correctly call the overloaded function. This is done by using the generic Instance type from the apply overloads scope.
//
// this must end with a comma before the newline, as the overrides generator will call this from a function parameter list.
func (vf *VirtualMethodGenerator) generateTrampoline(w file.File, genericType, goFn string) {
	fmt.Fprintf(w.Go(), "func%s {\n", vf.CGoTrampolineTail())
	w.Go().Indent()

	var decls file.DeclarationWriter

	fmt.Fprintf(&decls, "var\t%s\t%s\t// go %s subclass\n", vf.InstanceParam.GoName, genericType, vf.InstanceParam.Type.Type.CType(0))

	for i, param := range vf.GoParameters {
		if param.Implicit || param.Skip {
			continue
		}
		conv := vf.ParamConverters[i]
		fmt.Fprintf(&decls, "var\t%s\t%s\t// %s\n", param.GoName, param.GoType(), conv.Metadata())

		w.GoImportType(param.Type)
	}

	for i, ret := range vf.GoReturns {
		if ret.Implicit || ret.Skip {
			continue
		}

		conv := vf.ReturnConverters[i]
		fmt.Fprintf(&decls, "var\t%s\t%s\t// %s\n", ret.GoName, ret.GoType(), conv.Metadata())

		w.GoImportType(ret.Type)
	}

	decls.WriteTo(w.Go())

	w.Go().NewSection()

	// convert the instanceparam manually here, because we need another cast:
	fmt.Fprintf(w.Go(), "%s = %s(unsafe.Pointer(%s)).(%s)\n", vf.InstanceParam.GoName, vf.Parent.GoUnsafeFromGlibBorrowFunction(), vf.InstanceParam.CName, genericType)

	for _, c := range vf.ParamConverters {
		c.Convert(w)
	}

	w.Go().NewSection()

	goReturns := vf.GoReturns.GoIdentifiers()

	if goReturns == "" {
		fmt.Fprintf(w.Go(), "%s(%s)\n", goFn, vf.GoIdentifiers())
	} else {
		fmt.Fprintf(w.Go(), "%s = %s(%s)\n", goReturns, goFn, vf.GoIdentifiers())
	}

	w.Go().NewSection()

	for _, c := range vf.ReturnConverters {
		c.Convert(w)
	}

	w.Go().NewSection()

	if vf.CReturn != nil && vf.CReturn.Type.Type != typesystem.Void {
		fmt.Fprintf(w.Go(), "return %s\n", vf.CReturn.CName)
	}

	w.Go().Unindent()
	fmt.Fprintf(w.Go(), "},\n")
}

// GenerateClassOverrideField generates the field name and type for the class override structs field.
//
// as opposed to interface signatures this must include the instance parameter as a first parameter of
// the method signature. The type of the instance parameter is always OverridesInstanceGenericType, as that is the
// generic type of the struct.
func (m *VirtualMethodGenerator) GenerateClassOverrideField(w file.File) {
	var ret string
	if len(m.GoReturns) == 1 {
		ret = " " + m.GoReturns[0].GoType()
	} else if len(m.GoReturns) > 1 {
		ret = " (" + m.GoReturns.GoTypes() + ")"
	}

	var vparams = []string{
		overridesInstanceGenericType, // first param is the generic instance
	}

	for _, param := range m.GoParameters {
		if param.Skip || param.Implicit {
			continue
		}

		w.GoImportType(param.Type)

		vparams = append(vparams, param.GoType())
	}

	fmt.Fprintf(w.Go(), "// %s allows you to override the implementation of the virtual method %s.\n", m.GoName, m.Invoker.CIndentifier())
	m.Doc.Generate(w.Go())
	fmt.Fprintf(w.Go(), "%s func(%s)%s\n", m.GoName, strings.Join(vparams, ", "), ret)
}

// CGoTrampolineSignature returns a string of the cgo trampoline function signature.
func (m *VirtualMethodGenerator) CGoTrampolineSignature() string {
	return fmt.Sprintf("func %s%s", m.VirtualMethod.TrampolineName, m.CGoTrampolineTail())
}

func (m *VirtualMethodGenerator) CGoTrampolineTail() string {
	var ret string
	if m.VirtualMethod.CReturn != nil && m.VirtualMethod.CReturn.Type.Type != typesystem.Void {
		ret = fmt.Sprintf(" (%s %s)", m.VirtualMethod.CReturn.CName, m.VirtualMethod.CReturn.CGoType())
	}

	paramDecls := make([]string, 0, len(m.VirtualMethod.CParameters()))

	for _, p := range m.VirtualMethod.CParameters() {
		additionalPointer := ""
		if p.Direction == "out" {
			additionalPointer = "*"
		}

		paramDecls = append(
			paramDecls,
			fmt.Sprintf("%s %s%s", p.CName, additionalPointer, p.CGoType()),
		)
	}

	return fmt.Sprintf("(%s)%s", strings.Join(paramDecls, ", "), ret)
}

func (m *VirtualMethodGenerator) GoIdentifiers() string {
	var params []string

	params = append(params, m.VirtualMethod.InstanceParam.GoName)

	for _, param := range m.VirtualMethod.GoParameters {
		if param.Skip || param.Implicit {
			continue
		}

		params = append(params, param.GoName)
	}
	return strings.Join(params, ", ")
}

func NewVirtualMethodGenerator(vfunc *typesystem.VirtualMethod) *VirtualMethodGenerator {
	g := &VirtualMethodGenerator{
		Doc:           NewParametersGoDocGenerator(vfunc),
		VirtualMethod: vfunc,
	}

	for _, param := range vfunc.GoParameters {
		g.ParamConverters = append(g.ParamConverters, convert.NewCToGoConverter(param))
	}

	for _, param := range vfunc.GoReturns {
		g.ReturnConverters = append(g.ReturnConverters, convert.NewGoToCConverter(param))
	}

	return g
}
