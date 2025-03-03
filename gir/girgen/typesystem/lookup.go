package typesystem

import (
	"fmt"
	"log"
	"strings"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

type NamespaceMetadata struct {
	MajorVersion int

	GoPackageName string

	Namespace *namespace
}

func (r *Registry) GetNamespaceMetadata(ns *gir.Namespace) *NamespaceMetadata {
	found, ok := r.namespaces[versionedNamespace{name: ns.Name, majorversion: gir.MajorVersion(ns.Version)}]

	if !ok {
		return nil
	}

	return &NamespaceMetadata{
		MajorVersion:  parseMajorVersion(found.version),
		GoPackageName: found.goPackageName,
		Namespace:     found,
	}
}

type TypeMetadata struct {
	GoPointers      int
	CGoPointers     int
	GoBaseType      string
	CGoBaseType     string
	RequiredImports []string

	IsCastable bool

	GirType any
}

func (tm TypeMetadata) GoType() string {
	return addPointers(tm.GoBaseType, tm.GoPointers)
}

func (tm TypeMetadata) CGoType() string {
	return addPointers(tm.CGoBaseType, tm.CGoPointers)
}

// LookupType finds a type in the registry and attaches metadata needed for generation. The typestring must be versioned or primitive.
func (r *Registry) LookupType(typ string) *TypeMetadata {
	typ = strings.TrimPrefix(typ, "const ")

	parts := strings.Split(typ, ".")

	if len(parts) > 1 {
		panic("searching for versioned type in whole registry not yet implemented")
	}

	if t := lookupPrimitive(parts[0]); t != nil {
		return t
	}

	return nil
}

// LookupType finds a type from within a namespace. Will only resolve the types known to the namespace. The type must not be versioned,
// instead the correct version of the include from the namespace is taken.
func (n *namespace) LookupType(typ string) *TypeMetadata {
	typ = strings.TrimPrefix(typ, "const ")

	// some types with pointers are not pointer types in go, e.g. gchar*
	if t := lookupPrimitive(typ); t != nil {
		return t
	}

	baseType, pointers := trimPointers(typ)

	parts := strings.Split(baseType, ".")

	if len(parts) > 1 {
		// lookup in referenced namespace again
		typ := strings.Join(parts[1:], ".")

		// search in the correct included namespace.
		reffedNS := n.includes[parts[0]]

		if reffedNS == nil {
			return nil
		}

		t := reffedNS.LookupType(typ)

		if t == nil {
			return nil
		}

		// and attach the required import:
		t.RequiredImports = append(t.RequiredImports, reffedNS.goImportPath)
		// the imported package is also needed for the type name:
		t.GoBaseType = reffedNS.goPackageName + "." + t.GoBaseType

		return t
	}

	if t := lookupPrimitive(parts[0]); t != nil {
		t.CGoPointers += pointers
		t.GoPointers += pointers
		return t
	}

	var meta *TypeMetadata

	for name, girType := range n.aliasesByName {
		if name == baseType {
			meta = &TypeMetadata{
				GoBaseType:  strcases.PascalToGo(girType.Name),
				CGoBaseType: ctypeToCGo(baseType),
				GirType:     &girType,
				IsCastable:  true,
			}
		}
	}

	for name, girType := range n.classesByName {
		if name == baseType {
			meta = &TypeMetadata{
				GoBaseType:  strcases.PascalToGo(girType.Name),
				CGoBaseType: ctypeToCGo(baseType),
				GirType:     &girType,
			}
		}
	}

	for name, girType := range n.interfacesByName {
		if name == baseType {
			meta = &TypeMetadata{
				GoBaseType:  strcases.PascalToGo(girType.Name),
				CGoBaseType: ctypeToCGo(baseType),
				GirType:     &girType,
			}
		}
	}

	for name, girType := range n.recordsByName {
		if name == baseType {
			meta = &TypeMetadata{
				GoBaseType:  strcases.PascalToGo(girType.Name),
				CGoBaseType: ctypeToCGo(baseType),
				GirType:     &girType,
			}
		}
	}

	for name, girType := range n.enumsByName {
		if name == baseType {
			meta = &TypeMetadata{
				GoBaseType:  strcases.PascalToGo(girType.Name),
				CGoBaseType: ctypeToCGo(baseType),
				GirType:     &girType,
				IsCastable:  true,
			}
		}
	}

	for name, girType := range n.functionsByName {
		if name == baseType {
			meta = &TypeMetadata{
				GoBaseType:  strcases.PascalToGo(girType.Name),
				CGoBaseType: ctypeToCGo(baseType),
				GirType:     &girType,
			}
		}
	}

	for name, girType := range n.unionsByName {
		if name == baseType {
			meta = &TypeMetadata{
				GoBaseType:  strcases.PascalToGo(girType.Name),
				CGoBaseType: ctypeToCGo(baseType),
				GirType:     &girType,
			}
		}
	}

	for name, girType := range n.bitfieldsByName {
		if name == baseType {
			meta = &TypeMetadata{
				GoBaseType:  strcases.PascalToGo(girType.Name),
				CGoBaseType: ctypeToCGo(baseType),
				GirType:     &girType,
			}
		}
	}

	for name, girType := range n.callbacksByName {
		if name == baseType {
			goBaseType := strcases.PascalToGo(girType.Name)
			meta = &TypeMetadata{
				GoBaseType:  goBaseType,
				CGoBaseType: fmt.Sprintf("_gotk4_%s%s_%s", n.goPackageName, gir.MajorVersion(n.version), goBaseType),
				GirType:     &girType,
			}
		}
	}

	for name, girType := range n.constantsByName {
		if name == baseType {
			meta = &TypeMetadata{
				GoBaseType:  strcases.PascalToGo(girType.Name),
				CGoBaseType: ctypeToCGo(baseType),
				GirType:     &girType,
			}
		}
	}

	if meta == nil && n.name != "GObject" {

		// special case: GObject types are referenced without a namespace prefix
		if reffedNS, ok := n.includes["GObject"]; ok {
			log.Printf("trying fallback to GObject.%s\n", typ)

			t := reffedNS.LookupType(typ)

			if t == nil {
				return nil
			}

			// and attach the required import:
			t.RequiredImports = append(t.RequiredImports, reffedNS.goImportPath)
			// the imported package is also needed for the type name:
			t.GoBaseType = reffedNS.goPackageName + "." + t.GoBaseType

			return t
		}
	}

	if meta != nil {
		meta.CGoPointers = pointers
		meta.GoPointers = pointers
	} else {
		log.Printf("type lookup not found: %s\n", typ)
	}

	return meta
}

var cgoSpecialTypes = map[string]string{
	"long long":          "C.longlong",
	"unsigned char":      "C.uchar",
	"unsigned int":       "C.uint",
	"unsigned short":     "C.ushort",
	"unsigned long":      "C.ulong",
	"unsigned long long": "C.ulonglong",
}

func ctypeToCGo(t string) string {
	if mapped, ok := cgoSpecialTypes[t]; ok {
		return mapped
	}

	return "C." + t
}

var builtinTypeMap = map[string]string{
	"long long": "C.longlong",

	"unsigned char":      "C.uchar",
	"unsigned int":       "C.uint",
	"unsigned short":     "C.ushort",
	"unsigned long":      "C.ulong",
	"unsigned long long": "C.ulonglong",

	"none":     "",
	"void":     "",
	"gboolean": "bool",
	"gfloat":   "float32",
	"gdouble":  "float64",
	"gint":     "int",
	"gssize":   "int",
	"gint8":    "int8",
	"gint16":   "int16",
	"gshort":   "int16",
	"gint32":   "int32",
	"glong":    "int32",
	"int32":    "int32",
	"gint64":   "int64",
	"guint":    "uint",
	"gsize":    "uint",
	"guchar":   "byte",
	"gchar":    "byte",
	"guint8":   "byte", // some weird cases
	"guint16":  "uint16",
	"gushort":  "uint16",
	"guint32":  "uint32",
	"gulong":   "uint32",
	"gunichar": "uint32",
	"guint64":  "uint64",
	"guintptr": "uintptr",
	"utf8":     "string",
	"filename": "string",
}

func lookupPrimitive(t string) *TypeMetadata {
	base, ptrs := trimPointers(t)

	if base == "gchar" && ptrs >= 1 {
		return &TypeMetadata{
			GoBaseType:  "string",
			CGoBaseType: "C." + base,
			GoPointers:  ptrs - 1,
			CGoPointers: ptrs,
			IsCastable:  false,
		}
	}

	if goType, ok := builtinTypeMap[base]; ok {
		meta := &TypeMetadata{
			GoBaseType:  goType,
			CGoBaseType: ctypeToCGo(base),
			GoPointers:  ptrs,
			CGoPointers: ptrs,
			IsCastable:  goType != "string",
		}

		return meta
	}

	switch t {
	case "gpointer":
		return &TypeMetadata{
			GoBaseType:      "unsafe.Pointer",
			CGoBaseType:     "C.gpointer",
			RequiredImports: []string{"unsafe"},
			IsCastable:      true,
		}
	case "gconstpointer":
		return &TypeMetadata{
			GoBaseType:      "unsafe.Pointer",
			CGoBaseType:     "C.gconstpointer",
			RequiredImports: []string{"unsafe"},
			IsCastable:      true,
		}
	}

	return nil
}

func trimPointers(t string) (basetype string, pointercount int) {
	count := strings.Count(t, "*")

	return strings.TrimRight(t, "*"), count
}

func addPointers(t string, pointers int) string {
	if t == "" {
		return ""
	}
	pointerstr := strings.Repeat("*", pointers)

	return pointerstr + t
}
