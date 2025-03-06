package typesystem

import (
	"fmt"
	"log"
	"strings"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/girgen/strcases"
)

func (r *Registry) GetNamespace(ns *gir.Namespace) *Namespace {
	found, ok := r.namespaces[VersionedNamespace{Name: ns.Name, MajorVersion: ns.MajorVersion()}]

	if !ok {
		return nil
	}

	return found
}

type TypeMetadata struct {
	GoPointers      int
	CGoPointers     int
	GoBaseType      string
	CGoBaseType     string
	RequiredImports []string

	IsCastable bool

	GirType   any
	Namespace *Namespace
}

func (tm TypeMetadata) GoType() string {
	return addPointers(tm.GoBaseType, tm.GoPointers)
}

func (tm TypeMetadata) CGoType() string {
	return addPointers(tm.CGoBaseType, tm.CGoPointers)
}

type TypeSystem interface {
	LookupType(typname, ctype string) *TypeMetadata
}

// LookupType finds a type in the registry and attaches metadata needed for generation. The typestring must be versioned or primitive.
func (r *Registry) LookupType(typname, ctype string) *TypeMetadata {
	ctype = strings.TrimPrefix(ctype, "const ")

	parts := strings.Split(typname, ".")

	if len(parts) > 2 {
		panic(fmt.Sprintf("received invalid type name %s", typname))
	}

	if len(parts) == 1 {
		return lookupPrimitive(ctype)
	}

	// TODO: GObject.Object does not suffice for a lookup, because we might have multiple versions of gobject

	panic("Lookup for versioned types unimplemented")
}

// LookupType finds a type from within a namespace. Will only resolve the types known to the namespace. The type must not be versioned,
// instead the correct version of the include from the namespace is taken.
func (n *Namespace) LookupType(typname, ctype string) *TypeMetadata {
	ctype = strings.TrimPrefix(ctype, "const ")

	// some types with pointers are not pointer types in go, e.g. gchar*, resolve them before trimming the pointers
	if t := lookupPrimitive(ctype); t != nil {
		return t
	}

	_, pointers := trimPointers(ctype)

	parts := strings.Split(typname, ".")

	if len(parts) > 2 {
		panic(fmt.Sprintf("received invalid type name %s", typname))
	}

	var meta *TypeMetadata

	switch len(parts) {
	case 1:
		// look in current namespace
		meta = n.findByName(parts[0])

		if meta != nil {
			meta.Namespace = n
		}
	case 2:
		// look in referenced namespace
		if reffedNS, ok := n.includes[parts[0]]; ok && reffedNS != nil {
			meta = reffedNS.findByName(parts[1])

			if meta != nil {
				// and attach the required import:
				meta.RequiredImports = append(meta.RequiredImports, reffedNS.goImportPath)
				// the imported package is also needed for the type name:
				meta.GoBaseType = reffedNS.GoPackageName + "." + meta.GoBaseType

				meta.Namespace = reffedNS
			}
		}
	default:
		panic(fmt.Sprintf("received invalid type name %s", typname))
	}

	if meta != nil {
		meta.CGoPointers = pointers
		meta.GoPointers = pointers
	} else {
		log.Printf("type lookup not found for %s\n", typname)
	}

	return meta
}

// findByName returns the [TypeMetdaata] for the named type
func (n *Namespace) findByName(typeName string) *TypeMetadata {
	var meta *TypeMetadata

	for name, girType := range n.aliasesByName {
		if name == typeName {
			meta = &TypeMetadata{
				GoBaseType:  strcases.PascalToGo(girType.Name),
				CGoBaseType: ctypeToCGo(girType.CType),
				GirType:     &girType,
				IsCastable:  true,
			}
		}
	}

	for name, girType := range n.classesByName {
		if name == typeName {
			meta = &TypeMetadata{
				GoBaseType:  strcases.PascalToGo(girType.Name),
				CGoBaseType: ctypeToCGo(girType.CType),
				GirType:     &girType,
			}
		}
	}

	for name, girType := range n.interfacesByName {
		if name == typeName {
			meta = &TypeMetadata{
				GoBaseType:  strcases.PascalToGo(girType.Name),
				CGoBaseType: ctypeToCGo(girType.CType),
				GirType:     &girType,
			}
		}
	}

	for name, girType := range n.recordsByName {
		if name == typeName {
			meta = &TypeMetadata{
				GoBaseType:  strcases.PascalToGo(girType.Name),
				CGoBaseType: ctypeToCGo(girType.CType),
				GirType:     &girType,
			}
		}
	}

	for name, girType := range n.enumsByName {
		if name == typeName {
			meta = &TypeMetadata{
				GoBaseType:  strcases.PascalToGo(girType.Name),
				CGoBaseType: ctypeToCGo(girType.CType),
				GirType:     &girType,
				IsCastable:  true,
			}
		}
	}

	for name, girType := range n.functionsByName {
		if name == typeName {
			meta = &TypeMetadata{
				GoBaseType:  strcases.PascalToGo(girType.Name),
				CGoBaseType: ctypeToCGo(girType.CIdentifier),
				GirType:     &girType,
			}
		}
	}

	for name, girType := range n.unionsByName {
		if name == typeName {
			meta = &TypeMetadata{
				GoBaseType:  strcases.PascalToGo(girType.Name),
				CGoBaseType: ctypeToCGo(girType.CType),
				GirType:     &girType,
			}
		}
	}

	for name, girType := range n.bitfieldsByName {
		if name == typeName {
			meta = &TypeMetadata{
				GoBaseType:  strcases.PascalToGo(girType.Name),
				CGoBaseType: ctypeToCGo(girType.CType),
				GirType:     &girType,
			}
		}
	}

	for name, girType := range n.callbacksByName {
		if name == typeName {
			goBaseType := strcases.PascalToGo(girType.Name)
			meta = &TypeMetadata{
				GoBaseType:  goBaseType,
				CGoBaseType: fmt.Sprintf("_gotk4_%s%d_%s", n.GoPackageName, n.VersionedName.MajorVersion, goBaseType),
				GirType:     &girType,
			}
		}
	}

	for name, girType := range n.constantsByName {
		if name == typeName {
			meta = &TypeMetadata{
				GoBaseType:  strcases.PascalToGo(girType.Name),
				CGoBaseType: ctypeToCGo(girType.CType),
				GirType:     &girType,
			}
		}
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
	"int":      "int",
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

func lookupPrimitive(ctype string) *TypeMetadata {
	base, ptrs := trimPointers(ctype)

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

	switch ctype {
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
