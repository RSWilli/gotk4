// Package gendata contains data used to generate GTK4 bindings for Go. It
// exists primarily to be used externally.
package gendata

import (
	"slices"

	"github.com/diamondburned/gotk4/gir"
	"github.com/diamondburned/gotk4/gir/cmd/gir-generate/genmain"
	"github.com/diamondburned/gotk4/gir/girgen/typesystem"
	girfiles_gotk4 "github.com/diamondburned/gotk4/girs"
)

const Module = "github.com/diamondburned/gotk4/pkg"

var Main = genmain.Data{
	Module:        Module,
	GirFiles:      girfiles_gotk4.GirFiles,
	Preprocessors: Preprocessors,
	// Postprocessors: []typesystem.PostProcessor{
	// 	// remove all virtual methods from the Atk, Gtk classes and interfaces.
	// 	// FIXME: this is a workaround to get it to compile. Correct fix would be to rename
	// 	// all collisions
	// 	func(r *typesystem.Registry) error {
	// 		nss := []string{"Atk-1", "Gtk-3"}

	// 		for _, nsident := range nss {
	// 			ns := r.FindNamespaceByName(nsident)

	// 			for _, class := range ns.Classes {
	// 				class.VirtualMethods = nil
	// 			}
	// 			for _, inter := range ns.Interfaces {
	// 				inter.VirtualMethods = nil
	// 			}
	// 		}

	// 		return nil
	// 	},
	// },

	Config: typesystem.Config{
		GIRReplacements: map[string]string{
			"GType": "GObject.Type", // GType is often referred to as a global type instead of GObject scoped
		},
		Namespaces: map[string]typesystem.NamespaceConfig{
			// "Graphene-1": {
			// 	ManualTypes: []typesystem.Type{
			// 		// graphenes gbooleans are not the same as glib's
			// 		// these are castable to bool, see below preprocessor
			// 		&typesystem.CastablePrimitive{
			// 			BaseType: typesystem.BaseType{
			// 				GirName: "gotk4-graphene-gboolean",
			// 				GoTyp:   "bool",
			// 				CGoTyp:  "C._Bool",
			// 				CTyp:    "_Bool",
			// 			},
			// 		},
			// 	},
			// },
			// "cairo-1": {
			// 	IgnoredDefinitions: []typesystem.IgnoreFunc{
			// 		// These are not in gotk3/cairo.
			// 		typesystem.IgnoreMatching("ScaledFont"),
			// 		typesystem.IgnoreMatching("FontType"),
			// 	},
			// },
			// "Atspi-2": {
			// 	Ignored: true, // Missing AtspiDevice
			// },
			"GLib-2": {
				MinVersion: "2.80",
				ManualTypes: []typesystem.Type{
					&typesystem.Container{
						GirName: "List",
						C:       "GList",
						CGo:     "C.GList",
						MakeGoType: func(innerTypes []typesystem.CouldBeForeign[typesystem.Type]) string {
							return "[]" + innerTypes[0].NamespacedGoType(1)
						},
						NumInnerTypes:        1,
						FromGlibFullFunction: "UnsafeListFromGlibFull",
						FromGlibNoneFunction: "UnsafeListFromGlibNone",
					},
					&typesystem.Container{
						GirName: "SList",
						C:       "GSList",
						CGo:     "C.GSList",
						MakeGoType: func(innerTypes []typesystem.CouldBeForeign[typesystem.Type]) string {
							return "[]" + innerTypes[0].NamespacedGoType(1)
						},
						NumInnerTypes:        1,
						FromGlibFullFunction: "UnsafeSListFromGlibFull",
						FromGlibNoneFunction: "UnsafeSListFromGlibNone",
					},
					&typesystem.Callback{
						BaseType: typesystem.BaseType{
							GirName: "DestroyNotify",
							GoTyp:   "DestroyNotify",
							CGoTyp:  "C.GDestroyNotify",
							CTyp:    "GDestroyNotify",
						},
						Parameters: &typesystem.Parameters{
							CReturn: typesystem.NewManualParam("cret", "goret", typesystem.Void, 0),
							GIRParameters: typesystem.ParamList{
								typesystem.NewManualParam("arg0", "goarg0", typesystem.Gpointer, 0),
							},
						},
						TrampolineName: "destroyUserdata",
					},
					// &typesystem.Record{
					// 	BaseType: typesystem.BaseType{
					// 		GirName: "Variant",
					// 		GoTyp:   "Variant",
					// 		CGoTyp:  "C.GVariant",
					// 		CTyp:    "GVariant",
					// 	},
					// 	BaseConversions: typesystem.BaseConversions{
					// 		FromGlibBorrowFunction: "UnsafeVariantFromGlibBorrow",
					// 		FromGlibFullFunction:   "UnsafeVariantFromGlibFull",
					// 		FromGlibNoneFunction:   "UnsafeVariantFromGlibNone",
					// 		ToGlibNoneFunction:     "UnsafeVariantToGlibNone",
					// 		ToGlibFullFunction:     "UnsafeVariantToGlibFull",
					// 	},
					// },
					&typesystem.Record{
						BaseType: typesystem.BaseType{
							GirName: "Error",
							CGoTyp:  "C.GError",
							CTyp:    "GError",
							GoTyp:   "error",
						},
						BaseConversions: typesystem.BaseConversions{
							FromGlibBorrowFunction: "UnsafeErrorFromGlibBorrow",
							FromGlibFullFunction:   "UnsafeErrorFromGlibFull",
							FromGlibNoneFunction:   "UnsafeErrorFromGlibNone",
							ToGlibNoneFunction:     "UnsafeErrorToGlibNone",
							ToGlibFullFunction:     "UnsafeErrorToGlibFull",
						},
					},
				},
				IgnoredDefinitions: []typesystem.IgnoreFunc{
					// not found:
					typesystem.IgnoreMatching("StatBuf"),
					typesystem.IgnoreMatching("access"),
					typesystem.IgnoreMatching("chdir"),
					typesystem.IgnoreMatching("chmod"),
					typesystem.IgnoreMatching("close"),
					typesystem.IgnoreMatching("closefrom"),
					typesystem.IgnoreMatching("creat"),
					typesystem.IgnoreMatching("date_get_week_of_year"),
					typesystem.IgnoreMatching("date_get_weeks_in_year"),
					typesystem.IgnoreMatching("fdwalk_set_cloexec"),
					typesystem.IgnoreMatching("fsync"),
					typesystem.IgnoreMatching("log_get_always_fatal"),
					typesystem.IgnoreMatching("lstat"),
					typesystem.IgnoreMatching("mkdir"),
					typesystem.IgnoreMatching("open"),
					typesystem.IgnoreMatching("remove"),
					typesystem.IgnoreMatching("rename"),
					typesystem.IgnoreMatching("rmdir"),
					typesystem.IgnoreMatching("source_dup_context"),
					typesystem.IgnoreMatching("stat"),
					typesystem.IgnoreMatching("string_copy"),
					typesystem.IgnoreMatching("unlink"),

					typesystem.IgnoreByRegex("Date.*"),
					typesystem.IgnoreMatching("Source"),
					typesystem.IgnoreMatching("TestLogMsg"),
					typesystem.IgnoreMatching("String"),
					typesystem.IgnoreMatching("ThreadPool"),

					typesystem.IgnoreMatching("HookFlagMask"), // Has a member of the same name

					typesystem.IgnoreMatching("Variant"),           // TODO: implement manually
					typesystem.IgnoreMatching("variant_get_gtype"), // implemented with gvalue in gobject

					typesystem.IgnoreMatching("ucs4_to_utf16"), // returns a pointer instead of an array

					// Nothing "Unix" is going to be available on Windows.
					typesystem.IgnoreByRegex(".*[Uu]nix.*"),
					// Useless
					typesystem.IgnoreMatching("NullifyPointer"),
					typesystem.IgnoreByRegex("[Aa]tomic.*"),
					typesystem.IgnoreByRegex("ATOMIC.*"),
					// Dangerous.
					typesystem.IgnoreMatching("IOChannel.read"),
					typesystem.IgnoreMatching("Bytes.new_take"),
					typesystem.IgnoreMatching("Bytes.new_static"),
					typesystem.IgnoreMatching("Bytes.unref_to_data"),
					typesystem.IgnoreMatching("Bytes.unref_to_array"),

					typesystem.IgnoreMatching("G_WIN32_MSG_HANDLE"),
					typesystem.IgnoreMatching("GLIB_VERSION_MIN_REQUIRED"),

					typesystem.IgnoreMatching("strv_get_type"), // requires gobject

					// see https://gitlab.gnome.org/GNOME/gobject-introspection/-/issues/305#note_981623
					// Container structures that are unused:
					typesystem.IgnoreMatching("Array"),
					typesystem.IgnoreMatching("Queue"),
					typesystem.IgnoreMatching("Tree"),
					typesystem.IgnoreMatching("PtrArray"),
					typesystem.IgnoreMatching("HashTable"),
				},
			},
			"Gio-2": {
				ManualTypes: []typesystem.Type{
					&typesystem.Record{
						BaseType: typesystem.BaseType{
							GirName: "Cancellable",
							CGoTyp:  "C.GCancellable",
							CTyp:    "GCancellable",
							GoTyp:   "context.Context",

							GoImport: "context",
						},
						BaseConversions: typesystem.BaseConversions{
							FromGlibBorrowFunction: "",
							FromGlibFullFunction:   "",
							FromGlibNoneFunction:   "NewCancellableContext",
							ToGlibNoneFunction:     "UnsafeGCancellableToGlibNone",
							ToGlibFullFunction:     "",
						},
					},
				},
				IgnoredDefinitions: []typesystem.IgnoreFunc{
					// Nothing "Unix" is going to be available on Windows.
					typesystem.IgnoreByRegex(".*[Uu]nix.*"),
					typesystem.IgnoreByRegex(".*Subprocess.*"),

					typesystem.IgnoreByRegex("FileDescriptorBased"),
					typesystem.IgnoreByRegex("SettingsBackend"),
					typesystem.IgnoreByRegex("DesktopAppInfo.*"),
					typesystem.IgnoreByRegex("DBus.*"),
					typesystem.IgnoreByRegex("ThreadedResolver.*"),

					typesystem.IgnoreMatching("ZlibCompressor"),

					typesystem.IgnoreMatching("networking_init"),

					typesystem.IgnoreMatching("DataInputStream.read_byte"), // collides with BufferedInputStream.read_byte
				},
			},
			"GObject-2": {
				MinVersion: "2.80",
				ManualTypes: func() []typesystem.Type { // use an immediately invoked function to create the circular references
					object := &typesystem.Class{
						BaseType: typesystem.BaseType{
							GirName: "Object",
							GoTyp:   "ObjectInstance",
							CTyp:    "GObject",
							CGoTyp:  "C.GObject",
						},
						GoInterfaceName: "Object",
						Doc:             typesystem.Doc{},
						BaseConversions: typesystem.BaseConversions{
							FromGlibBorrowFunction: "UnsafeObjectFromGlibBorrow", // borrow is needed for subclassing
							FromGlibFullFunction:   "UnsafeObjectFromGlibFull",
							FromGlibNoneFunction:   "UnsafeObjectFromGlibNone",
							ToGlibNoneFunction:     "UnsafeObjectToGlibNone",
							ToGlibFullFunction:     "UnsafeObjectToGlibFull",
						},
						GoExtendOverrideStructName: "ObjectOverrides",
						GoUnsafeApplyOverridesName: "UnsafeApplyObjectOverrides",
					}
					objectClass := &typesystem.Record{
						BaseType: typesystem.BaseType{
							GirName: "ObjectClass",
							GoTyp:   "ObjectClass",
							CTyp:    "GObjectClass",
							CGoTyp:  "C.GObjectClass",
						},
						BaseConversions: typesystem.BaseConversions{
							FromGlibBorrowFunction: "UnsafeObjectClassFromGlibBorrow",
							// not transferable, we don't want any methods with this
						},
					}

					object.TypeStruct = objectClass
					objectClass.IsTypeStructFor = object

					return []typesystem.Type{
						&typesystem.Alias{
							BaseType: typesystem.BaseType{
								GirName: "Type",
								CTyp:    "GType",
								CGoTyp:  "C.GType",
								GoTyp:   "Type",
							},
							AliasedType: typesystem.CouldBeForeign[typesystem.Type]{
								Namespace: nil,
								Type:      typesystem.Guint64,
							},
						},
						object,
						objectClass,
						&typesystem.Record{
							BaseType: typesystem.BaseType{
								GirName: "Value",
								GoTyp:   "Value",
								CTyp:    "GValue",
								CGoTyp:  "C.GValue",
							},
							BaseConversions: typesystem.BaseConversions{
								FromGlibBorrowFunction: "ValueFromNative",

								// these should get implemented manually, because "any" would be a better match
								FromGlibFullFunction: "UnsafeValueFromGlibUseAnyInstead",
								FromGlibNoneFunction: "UnsafeValueFromGlibUseAnyInstead",
								ToGlibNoneFunction:   "UnsafeValueToGlibUseAnyInstead",
								ToGlibFullFunction:   "UnsafeValueToGlibUseAnyInstead",
							},
						},
						&typesystem.Record{
							BaseType: typesystem.BaseType{
								GirName: "ParamSpec",
								GoTyp:   "ParamSpec",
								CTyp:    "GParamSpec",
								CGoTyp:  "C.GParamSpec",
							},
							BaseConversions: typesystem.BaseConversions{
								FromGlibBorrowFunction: "UnsafeParamSpecFromGlibBorrow",
								FromGlibFullFunction:   "UnsafeParamSpecFromGlibFull",
								FromGlibNoneFunction:   "UnsafeParamSpecFromGlibNone",
								ToGlibNoneFunction:     "UnsafeParamSpecToGlibNone",
								ToGlibFullFunction:     "UnsafeParamSpecToGlibFull",
							},
						},
					}
				}(),
				IgnoredDefinitions: []typesystem.IgnoreFunc{
					// manually implemented, but hidden from the user
					typesystem.IgnoreMatching("ParamSpecClass"),
					typesystem.IgnoreMatching("Closure"),
					typesystem.IgnoreMatching("SignalGroup"),
					typesystem.IgnoreMatching("SignalQuery"),
					typesystem.IgnoreMatching("TypeQuery"),

					// maybe something for later, not needed now:
					typesystem.IgnoreMatching("TypeModule"),
					typesystem.IgnoreMatching("TypeModuleClass"),
					typesystem.IgnoreMatching("ParamSpecPool"),
					typesystem.IgnoreMatching("TypePlugin"),
					typesystem.IgnoreMatching("TypePluginClass"),
					typesystem.IgnoreMatching("ParamSpecTypeInfo"), // needed for registering param types

					typesystem.IgnoreMatching("Binding"), // is this needed?

					// signal handler accumulators, maybe implement them manually?
					typesystem.IgnoreMatching("signal_accumulator_first_wins"),
					typesystem.IgnoreMatching("signal_accumulator_true_handled"),

					// type registration is implemented manually but hidden from the user
					typesystem.IgnoreMatching("type_register_fundamental"),
					typesystem.IgnoreMatching("type_register_static"),

					typesystem.IgnoreMatching("type_check_value_holds"),
					typesystem.IgnoreMatching("type_check_value"),
					typesystem.IgnoreMatching("strdup_value_contents"),

					typesystem.IgnoreMatching("TypeInterface"), // base struct for interfaces, not needed
					typesystem.IgnoreMatching("TypeClass"),     // base struct for classes, not needed
				},
			},
			// "Atk-1": {
			// 	MinVersion: "2.50",
			// },
			// "Gdk-3": {
			// 	MinVersion: "3.24",
			// 	IgnoredDefinitions: []typesystem.IgnoreFunc{
			// 		typesystem.IgnoreByFileNameSubstring("gdkprivate"), // not found

			// 		// These return instance owned GValues, maybe implement them manually?
			// 		typesystem.IgnoreMatching("Clipboard.read_value_finish"),
			// 		typesystem.IgnoreMatching("ContentDeserializer.get_value"),
			// 		typesystem.IgnoreMatching("ContentSerializer.get_value"),
			// 		typesystem.IgnoreMatching("Drop.read_value_finish"),
			// 	},
			// },
			// "Gdk-4": {
			// 	MinVersion: "4.19",
			// 	IgnoredDefinitions: []typesystem.IgnoreFunc{
			// 		// These return instance owned GValues, maybe implement them manually?
			// 		typesystem.IgnoreMatching("Clipboard.read_value_finish"),
			// 		typesystem.IgnoreMatching("ContentDeserializer.get_value"),
			// 		typesystem.IgnoreMatching("ContentSerializer.get_value"),
			// 		typesystem.IgnoreMatching("Drop.read_value_finish"),
			// 	},
			// },
			// "GdkPixbuf-2": {
			// 	MinVersion: "2.42",
			// 	IgnoredDefinitions: []typesystem.IgnoreFunc{
			// 		// these are not found:
			// 		typesystem.IgnoreMatching("PixbufModule"),
			// 		typesystem.IgnoreMatching("PixbufNonAnim"),
			// 		typesystem.IgnoreMatching("PixbufModulePattern"),
			// 		typesystem.IgnoreMatching("PixbufFormat.domain"),
			// 		typesystem.IgnoreMatching("PixbufFormat.flags"),
			// 		typesystem.IgnoreMatching("PixbufFormat.disabled"),
			// 		typesystem.IgnoreMatching("PixbufAnimationClass"),
			// 		typesystem.IgnoreMatching("PixbufAnimationIterClass"),
			// 	},
			// },
			// "GdkX11-4": {
			// 	MinVersion: "4.19",
			// },
			// "GdkWayland-4": {
			// 	MinVersion: "4.19",
			// 	IgnoredDefinitions: []typesystem.IgnoreFunc{
			// 		// FIXME: returned type is converted to *unsafe.Pointer? https://docs.gtk.org/gdk4-wayland/method.WaylandDevice.get_xkb_keymap.html
			// 		typesystem.IgnoreMatching("WaylandDevice.get_xkb_keymap"),
			// 	},
			// },
			// "Gsk-4": {
			// 	MinVersion: "4.19",
			// 	IgnoredDefinitions: []typesystem.IgnoreFunc{
			// 		typesystem.IgnoreByFileNameSubstring("gsk/broadway/gskbroadwayrenderer.h"),
			// 	},
			// },
			// "Gtk-3": {
			// 	MinVersion: "3.24",
			// 	IgnoredDefinitions: []typesystem.IgnoreFunc{
			// 		// These are not found.
			// 		typesystem.IgnoreMatching("HeaderBarAccessibleClass"),
			// 		typesystem.IgnoreMatching("FileChooserWidgetAccessibleClass"),
			// 		typesystem.IgnoreMatching("_MountOperationHandler"),
			// 		typesystem.IgnoreMatching("_MountOperationHandlerIface"),
			// 		typesystem.IgnoreMatching("_MountOperationHandlerSkeleton"),
			// 		typesystem.IgnoreMatching("_MountOperationHandlerSkeletonClass"),
			// 		typesystem.IgnoreMatching("_MountOperationHandlerProxy"),
			// 		typesystem.IgnoreMatching("_MountOperationHandlerProxyClass"),
			// 	},
			// },
			// "Gtk-4": {
			// 	MinVersion: "4.19",
			// 	IgnoredDefinitions: []typesystem.IgnoreFunc{
			// 		// These are not found.
			// 		typesystem.IgnoreByFileNameSubstring("gtkpagesetupunixdialog"),
			// 		typesystem.IgnoreByFileNameSubstring("gtkprintunixdialog"),
			// 		typesystem.IgnoreByFileNameSubstring("gtkprinter"),
			// 		typesystem.IgnoreByFileNameSubstring("gtkprintjob"),
			// 		typesystem.IgnoreByRegex("Print.*"),
			// 	},
			// },
		},
	},
}

// Preprocessors defines a list of preprocessors that the main generator will
// use. It's mostly used for renaming colliding types/identifiers.
var Preprocessors = []gir.Preprocessor{
	// // Collision due to case conversions.
	gir.TypeRenamer("GLib-2.file_test", "test_file"),
	// // This collides with Native().
	// TypeRenamer("Gtk-4.Native", "NativeSurface"),
	// // This collides with Editable()
	// TypeRenamer("Gtk-4.Editable", "EditableTextWidget"),
	// // These collide with structs of the same names.
	// RenameEnumMembers("Pango-1.AttrType", "ATTR_(.*)", "ATTR_TYPE_$1"),
	// RenameEnumMembers("Gsk-4.RenderNodeType", ".*", "${0}_TYPE"),
	// RenameEnumMembers("Gdk-3.EventType", ".*", "${0}_TYPE"),
	// RenameEnumMembers("Gtk-4.GraphicsOffloadEnabled", ".*", "${0}_TYPE"),
	// See #28.
	gir.RemoveCIncludes("Gio-2.0.gir", "gio/gdesktopappinfo.h"),
	// These probably shouldn't be built on Windows.
	gir.RemovePkgconfig("Gio-2.0.gir", "gio-unix-2.0"),
	gir.RemoveCIncludes("Gio-2.0.gir", "gio/gfiledescriptorbased.h", `gio/gunix.*\.h`),

	// // Length and Value are invalid in Go. We manually handle them in GLibLogs.
	gir.RemoveRecordFields("GLib-2.LogField", "length", "value"),

	gir.ModifyParamDirections("Gio-2.InputStream.read", map[string]string{
		"buffer": "in",
		"count":  "in",
	}),
	gir.ModifyParamDirections("Gio-2.InputStream.read_async", map[string]string{
		"buffer": "in",
		"count":  "in",
	}),
	gir.ModifyParamDirections("Gio-2.InputStream.read_all", map[string]string{
		"buffer": "in",
		"count":  "in",
	}),
	gir.ModifyParamDirections("Gio-2.InputStream.read_all_async", map[string]string{
		"buffer": "in",
		"count":  "in",
	}),
	gir.ModifyParamDirections("Gio-2.Socket.receive", map[string]string{
		"buffer": "in",
		"size":   "in",
	}),
	gir.ModifyParamDirections("Gio-2.Socket.receive_from", map[string]string{
		"buffer": "in",
		"size":   "in",
	}),
	gir.ModifyParamDirections("Gio-2.Socket.receive_with_blocking", map[string]string{
		"buffer": "in",
		"size":   "in",
	}),
	gir.ModifyParamDirections("Gio-2.DBusInterfaceGetPropertyFunc", map[string]string{
		"error": "out",
	}),

	// ModifyCallable("Gdk-4.Clipboard.read_async", func(c *gir.CallableAttrs) {
	// 	// Fix this parameter's type not being a proper array.
	// 	p := FindParameter(c, "mime_types")
	// 	p.Array = &gir.Array{
	// 		CType: "const char**",
	// 		Type:  &gir.Type{Name: "utf8"},
	// 	}
	// }),

	// // These are not introspectable for some reason, even though their
	// // signatures look correct.
	// MustIntrospect("Gdk-4.Clipboard.set_text"),
	// MustIntrospect("Gdk-4.Clipboard.set_texture"),

	// // Fix up the return array type for (*Variant).String().
	// ModifyCallable("GLib-2.Variant.get_string", func(c *gir.CallableAttrs) {
	// 	c.ReturnValue.Type = nil
	// 	c.ReturnValue.Array = &gir.Array{
	// 		CType:          "const gchar*",
	// 		Type:           &gir.Type{Name: "gchar"},
	// 		Length:         new(int),  // 0
	// 		ZeroTerminated: new(bool), // false
	// 	}
	// }),

	// // Fix up Application::open's File type. It's supposed to be a GFile** from
	// // the source code, but that's missing from the GIR data.
	// ModifySignal("Gio-2.Application::open", func(sig *gir.Signal) {
	// 	param := FindParameterFromSlice(sig.Parameters.Parameters, "files")
	// 	param.Array.CType = "GFile**"
	// }),

	// // Fix up GVariant methods to have nullable returns.
	// PreprocessorFunc(func(repos gir.Repositories) {
	// 	variant := repos.FindFullType("GLib-2.Variant").Type.(*gir.Record)
	// 	for _, method := range variant.Methods {
	// 		returnsGVariant := true &&
	// 			method.ReturnValue != nil &&
	// 			method.ReturnValue.Type != nil &&
	// 			method.ReturnValue.Type.CType == "GVariant*"

	// 		if returnsGVariant && !method.ReturnValue.Nullable {
	// 			// GVariant pointers can be null.
	// 			method.ReturnValue.Nullable = true
	// 		}
	// 	}
	// }),

	// Fix GAsyncReadyCallback missing the closure bit for the user_data
	// parameter.
	gir.PreprocessorFunc(func(repos gir.Repositories) {
		callback := repos.FindFullType("Gio-2.AsyncReadyCallback").(*gir.Callback)

		userDataIx := slices.IndexFunc(
			callback.Parameters.Parameters,
			func(p *gir.Parameter) bool { return p.Name == "data" },
		)

		userData := callback.Parameters.Parameters[userDataIx]
		userData.Closure = &userDataIx
	}),

	// // Collisions on NoOpObject due to interface implementations:
	// gir.RenameCallable("Atk-1.Action.get_name", "get_action_name"),
	// gir.RenameCallable("Atk-1.Action.get_description", "get_action_description"),
	// gir.RenameCallable("Atk-1.Action.set_description", "set_action_description"),
	// gir.RenameCallable("Atk-1.Text.add_selection", "add_text_selection"),
	// gir.RenameCallable("Atk-1.Text.remove_selection", "remove_text_selection"),

	// Collide in other namespaces (e.g. Gio) when implementing TypePlugin and TypeModule
	gir.RenameCallable("GObject-2.TypePlugin.use", "use_plugin"),
	gir.RenameCallable("GObject-2.TypePlugin.unuse", "unuse_plugin"),

	// // Collide in Gtk-3 when implementing interface:
	// gir.RenameCallable("Gtk-3.Buildable.get_name", "get_buildable_name"),
	// gir.RenameCallable("Gtk-3.Buildable.set_name", "set_buildable_name"),
	// gir.RenameCallable("Gtk-3.ToolShell.get_orientation", "get_tool_shell_orientation"),
	// gir.RenameCallable("Gtk-3.ToolShell.get_icon_size", "get_tool_shell_icon_size"),
	// gir.RenameCallable("Gtk-3.Widget.child_notify", "widget_child_notify"),
	// gir.RenameCallable("Gtk-3.TextView.get_window", "get_text_view_window"),
	// gir.RenameCallable("Gtk-3.ComboBoxText.remove", "remove_combo_box_text"),
	// gir.RenameCallable("Gtk-3.Menu.set_accel_path", "set_menu_accel_path"),
	// gir.RenameCallable("Gtk-3.MenuItem.set_accel_path", "set_menu_item_accel_path"),
	// gir.RenameCallable("Gtk-3.Statusbar.remove", "remove_statusbar"),
	// gir.RenameCallable("Gtk-3.MenuItem.activate", "activate_menu_item"),
	// gir.RenameCallable("Gtk-3.Window.mnemonic_activate", "window_mnemonic_activate"),
	// gir.RenameCallable("Gtk-3.MenuButton.get_direction", "get_menu_button_direction"),
	// gir.RenameCallable("Gtk-3.MenuButton.set_direction", "set_menu_button_direction"),

	// // must rename to allow atk interface to be implemented
	// gir.RenameCallable("Gtk-3.CellAccessibleParent.grab_focus", "cell_accessible_parent_grab_focus"),

	// // Collide in Gtk-4 when implementing interface:
	// gir.RenameCallable("Gtk-4.MenuButton.get_direction", "get_menu_button_direction"),
	// gir.RenameCallable("Gtk-4.MenuButton.set_direction", "set_menu_button_direction"),

	// Collide with GObject.Connect:
	gir.RenameCallable("Gio-2.Socket.connect", "connect_socket"),
	gir.RenameCallable("Gio-2.SocketClient.connect", "connect_socket_client"),
	gir.RenameCallable("Gio-2.SocketConnection.connect", "connect_socket_connection"),
	gir.RenameCallable("Gio-2.Proxy.connect", "connect_proxy"),

	// Less confusing because C.int differs from int in Go.
	gir.RenameCallable("GObject-2.param_spec_int", "param_spec_int32"),

	// // Fix Graphenes _Bool return values to be castable to bool.
	// gir.PreprocessorFunc(func(r gir.Repositories) {
	// 	graphene := r.Find("Graphene-1").(*gir.Namespace)

	// 	switchBoolReturnType := func(c *gir.CallableAttrs) {
	// 		if c.ReturnValue == nil || c.ReturnValue.Type == nil {
	// 			return
	// 		}

	// 		if c.ReturnValue.Type.CType == "_Bool" {
	// 			c.ReturnValue.Type.Name = "gotk4-graphene-gboolean" // referenced above in ManualTypes
	// 		}
	// 	}

	// 	for _, f := range graphene.Functions {
	// 		switchBoolReturnType(&f.CallableAttrs)
	// 	}

	// 	for _, r := range graphene.Records {
	// 		for _, m := range r.Methods {
	// 			switchBoolReturnType(&m.CallableAttrs)
	// 		}
	// 		for _, f := range r.Functions {
	// 			switchBoolReturnType(&f.CallableAttrs)
	// 		}
	// 	}
	// 	for _, c := range graphene.Classes {
	// 		for _, m := range c.Methods {
	// 			switchBoolReturnType(&m.CallableAttrs)
	// 		}
	// 		for _, f := range c.Functions {
	// 			switchBoolReturnType(&f.CallableAttrs)
	// 		}
	// 	}
	// }),
}
