package typesystem

import (
	"fmt"
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
	GoType          string
	CGoType         string
	RequiredImports []string

	GirType any
}

// LookupType finds a type in the registry and attaches metadata needed for generation. The typestring must be versioned or primitive.
func (r *Registry) LookupType(typ string) *TypeMetadata {
	baseType, pointers := trimPointers(typ)

	parts := strings.Split(baseType, ".")

	if len(parts) > 1 {
		panic("searching for versioned type in whole registry not yet implemented")
	}

	if t := lookupPrimitive(parts[0]); t != nil {
		addGoPointers(t, pointers)
		return t
	}

	return nil
}

// LookupType finds a type from within a namespace. Will only resolve the types known to the namespace. The type must not be versioned,
// instead the correct version of the include from the namespace is taken.
func (n *namespace) LookupType(typ string) *TypeMetadata {
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
		t.GoType = reffedNS.goPackageName + "." + t.GoType

		addGoPointers(t, pointers)

		return t
	}

	if t := lookupPrimitive(parts[0]); t != nil {
		return t
	}

	var meta *TypeMetadata

	for name, girType := range n.aliasesByName {
		if name == baseType {
			meta = &TypeMetadata{
				GoType:  strcases.PascalToGo(girType.Name),
				CGoType: ctypeToCGo(baseType),
				GirType: &girType,
			}
		}
	}

	for name, girType := range n.classesByName {
		if name == baseType {
			meta = &TypeMetadata{
				GoType:  strcases.PascalToGo(girType.Name),
				CGoType: ctypeToCGo(baseType),
				GirType: &girType,
			}
		}
	}

	for name, girType := range n.interfacesByName {
		if name == baseType {
			meta = &TypeMetadata{
				GoType:  strcases.PascalToGo(girType.Name),
				CGoType: ctypeToCGo(baseType),
				GirType: &girType,
			}
		}
	}

	for name, girType := range n.recordsByName {
		if name == baseType {
			meta = &TypeMetadata{
				GoType:  strcases.PascalToGo(girType.Name),
				CGoType: ctypeToCGo(baseType),
				GirType: &girType,
			}
		}
	}

	for name, girType := range n.enumsByName {
		if name == baseType {
			meta = &TypeMetadata{
				GoType:  strcases.PascalToGo(girType.Name),
				CGoType: ctypeToCGo(baseType),
				GirType: &girType,
			}
		}
	}

	for name, girType := range n.functionsByName {
		if name == baseType {
			meta = &TypeMetadata{
				GoType:  strcases.PascalToGo(girType.Name),
				CGoType: ctypeToCGo(baseType),
				GirType: &girType,
			}
		}
	}

	for name, girType := range n.unionsByName {
		if name == baseType {
			meta = &TypeMetadata{
				GoType:  strcases.PascalToGo(girType.Name),
				CGoType: ctypeToCGo(baseType),
				GirType: &girType,
			}
		}
	}

	for name, girType := range n.bitfieldsByName {
		if name == baseType {
			meta = &TypeMetadata{
				GoType:  strcases.PascalToGo(girType.Name),
				CGoType: ctypeToCGo(baseType),
				GirType: &girType,
			}
		}
	}

	for name, girType := range n.callbacksByName {
		if name == baseType {
			goType := strcases.PascalToGo(girType.Name)
			meta = &TypeMetadata{
				GoType:  goType,
				CGoType: fmt.Sprintf("_gotk4_%s%s_%s", n.goPackageName, gir.MajorVersion(n.version), goType),
				GirType: &girType,
			}
		}
	}

	for name, girType := range n.constantsByName {
		if name == baseType {
			meta = &TypeMetadata{
				GoType:  strcases.PascalToGo(girType.Name),
				CGoType: ctypeToCGo(baseType),
				GirType: &girType,
			}
		}
	}

	addGoPointers(meta, pointers)

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
	// special pointer types
	"gchar*": "string",
}

func lookupPrimitive(t string) *TypeMetadata {
	if goType, ok := builtinTypeMap[t]; ok {
		base, ptrs := trimPointers(t)
		meta := &TypeMetadata{
			GoType:  goType,
			CGoType: ctypeToCGo(base),
		}

		addGoPointers(meta, ptrs)

		return meta
	}

	switch t {
	case "gpointer":
		return &TypeMetadata{
			GoType:          "unsafe.Pointer",
			CGoType:         "C.gpointer",
			RequiredImports: []string{"unsafe"},
		}
	case "gconstpointer":
		return &TypeMetadata{
			GoType:          "unsafe.Pointer",
			CGoType:         "C.gconstpointer",
			RequiredImports: []string{"unsafe"},
		}
	}

	return nil
}

func trimPointers(t string) (basetype string, pointercount int) {
	count := strings.Count(t, "*")

	return strings.TrimRight(t, "*"), count
}

func addGoPointers(t *TypeMetadata, pointers int) {
	if pointers == 0 || t == nil {
		return
	}
	pointerstr := strings.Repeat("*", pointers)

	t.GoType = pointerstr + t.GoType
	t.CGoType = pointerstr + t.CGoType
}
