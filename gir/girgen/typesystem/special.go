package typesystem

type Error struct {
	BaseType
}

var TypeError = &Error{
	BaseType: BaseType{
		GirName: "GLib.Error", // this doesn't resolve correctly from the GLib namespace
		GoTyp:   "error",
		CGoTyp:  "**C.GError",
		CTyp:    "GError**",
	},
}
