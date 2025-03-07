package typesystem

import "github.com/diamondburned/gotk4/gir"

type Primitive struct {
	baseType
}

func prim(girName, cType, cGoType, goType string) *Primitive {
	return &Primitive{
		baseType: baseType{
			girName: girName,
			cType:   cType,
			cGoType: cGoType,
			goType:  goType,
		},
	}
}

var Primitives = []*Primitive{
	prim("guint", "guint", "C.guint", "uint"),
	prim("guint8", "guint8", "C.guint8", "uint8"),
	prim("guint16", "guint16", "C.guint16", "uint16"),
	prim("guint32", "guint32", "C.guint32", "uint32"),
	prim("guint64", "guint64", "C.guint64", "uint64"),

	prim("gint", "gint", "C.gint", "int"),
	prim("gint8", "gint8", "C.gint8", "int8"),
	prim("gint16", "gint16", "C.gint16", "int16"),
	prim("gint32", "gint32", "C.gint32", "int32"),
	prim("gint64", "gint64", "C.gint64", "int64"),

	prim("gshort", "gshort", "C.gshort", "int16"),
	prim("gushort", "gushort", "C.gushort", "uint16"),

	prim("gsize", "gsize", "C.gsize", "uint"),
	prim("gssize", "gssize", "C.gssize", "int"),

	prim("gchar", "gchar", "C.char", "byte"),
	prim("gunichar", "gunichar", "C.gunichar", "uint32"),

	prim("gboolean", "gboolean", "C.gboolean", "bool"),

	prim("gfloat", "gfloat", "C.gfloat", "float32"),
	prim("gdouble", "gdouble", "C.gdouble", "float64"),

	prim("utf8", "gchar*", "*C.gchar", "string"),
	prim("filename", "gchar*", "*C.gchar", "string"),

	prim("gintptr", "gintptr", "C.gintptr", "uintptr"),
	prim("guintptr", "guintptr", "C.guintptr", "uintptr"),
	prim("gpointer", "gpointer", "C.gpointer", "unsafe.Pointer"),

	prim("glong", "glong", "C.glong", "int32"),
	prim("gulong", "gulong", "C.gulong", "uint32"),

	prim("time_t", "time_t", "C.time_t", "uint64"), // TODO: check go type

	prim("GType", "GType", "C.GType", "GType"), // TODO: check go type

	prim("pid_t", "pid_t", "C.pid_t", "int"),  // process ids
	prim("ino_t", "ino_t", "C.ino_t", "uint"), // file serial ids
	prim("uid_t", "uid_t", "C.uid_t", "uint"), // user ids, may be signed on some platforms
	prim("gid_t", "gid_t", "C.gid_t", "uint"), // group ids, may be signed on some platforms
}

type VoidType struct {
	baseType
}

var Void = &VoidType{
	baseType: baseType{
		girName: "none",
		goType:  typeInvalid,
		cGoType: "C.void",
		cType:   "void",
	},
}

func findPrimitiveType(t *gir.Type) Type {
	if t.Name == Void.GIRName() {
		return Void
	}

	for _, p := range Primitives {
		if p.GIRName() == t.Name {
			return WithPointers(t, p)
		}
	}

	return nil
}

type Error struct {
	baseType
}

var TypeError = &Error{
	baseType: baseType{
		girName: "GLib.Error", // this doesn't resolve correctly from the GLib namespace
		goType:  "error",
		cGoType: "**C.GError",
		cType:   "GError**",
	},
}

var IgnoredTypes = []string{
	"long double", // may be more precise than float64
}

func isIgnoredType(t *gir.Type) bool {
	for _, ign := range IgnoredTypes {
		if ign == t.Name {
			return true
		}
	}

	return false
}
