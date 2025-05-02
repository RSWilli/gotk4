package generators

import (
	"fmt"
	"strings"

	"github.com/diamondburned/gotk4/gir/girgen/file"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
)

type VirtualMethodGenerator struct {
	Doc SubGenerator
	*typesystem.VirtualMethod
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

func (c *VirtualMethodGenerator) generateExport(pkg *file.Package) {
	w := &pkg.Exported

	fmt.Fprintf(w.Go(), "//export %s\n", c.TrampolineName)

	// TODO: the out params here are missing a pointer because it was stripped in the type resolution

	fmt.Fprintf(w.Go(), "%s {\n", c.CGoTrampolineSignature())

	w.Go().Indent()

	fmt.Fprintf(w.Go(), "panic(\"unimplemented\")\n")
	// fmt.Fprintf(w.Go(), "var fn %s\n", c.GoType(0)) // declare fn as the callback itself

	// w.GoImport("unsafe")
	// w.GoImportCore("userdata")

	// fmt.Fprintf(w.Go(), "{\n")
	// w.Go().Indent()
	// fmt.Fprintf(w.Go(), "v := userdata.Load(unsafe.Pointer(%s))\n", c.UserdataParam.CName)
	// fmt.Fprintf(w.Go(), "if v == nil {\n")
	// fmt.Fprintf(w.Go(), "\tpanic(`callback not found`)\n")
	// fmt.Fprintf(w.Go(), "}\n")
	// fmt.Fprintf(w.Go(), "fn = v.(%s)\n", c.GoType(0))
	// w.Go().Unindent()
	// fmt.Fprintf(w.Go(), "}\n")

	// w.Go().NewSection()

	// var decls file.DeclarationWriter

	// for i, param := range c.Callback.GoParameters {
	// 	if param.Implicit || param.Skip {
	// 		continue
	// 	}
	// 	conv := c.ParamConverters[i]
	// 	fmt.Fprintf(&decls, "var\t%s\t%s\t// %s\n", param.GoName, param.GoType(), conv.Metadata())

	// 	w.GoImportType(param.Type)
	// }
	// for i, ret := range c.Callback.GoReturns {
	// 	if ret.Implicit || ret.Skip {
	// 		continue
	// 	}

	// 	conv := c.ReturnConverters[i]
	// 	fmt.Fprintf(&decls, "var\t%s\t%s\t// %s\n", ret.GoName, ret.GoType(), conv.Metadata())

	// 	w.GoImportType(ret.Type)
	// }

	// decls.WriteTo(w.Go())

	// w.Go().NewSection()

	// for _, c := range c.ParamConverters {
	// 	c.Convert(w)
	// }

	// w.Go().NewSection()

	// goReturns := c.GoReturns.GoIdentifiers()

	// if goReturns == "" {
	// 	fmt.Fprintf(w.Go(), "fn(%s)\n", c.GoParameters.GoIdentifiers())
	// } else {
	// 	fmt.Fprintf(w.Go(), "%s = fn(%s)\n", goReturns, c.GoParameters.GoIdentifiers())
	// }

	// w.Go().NewSection()

	// for _, c := range c.ReturnConverters {
	// 	c.Convert(w)
	// }

	// w.Go().NewSection()

	// if c.CReturn != nil && c.CReturn.Type.Type != typesystem.Void {
	// 	fmt.Fprintf(w.Go(), "return %s\n", c.CReturn.CName)
	// }

	w.Go().Unindent()

	fmt.Fprintln(w.Go(), "}")
	fmt.Fprintln(w.Go())
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

func NewVirtualMethodGenerator(vfunc *typesystem.VirtualMethod) *VirtualMethodGenerator {
	return &VirtualMethodGenerator{
		Doc:           NewParametersGoDocGenerator(vfunc),
		VirtualMethod: vfunc,
	}
}

// CGoTrampolineSignature returns a string of the cgo trampoline function signature.
func (m *VirtualMethodGenerator) CGoTrampolineSignature() string {
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

	return fmt.Sprintf("func %s(%s)%s", m.VirtualMethod.TrampolineName, strings.Join(paramDecls, ", "), ret)
}
