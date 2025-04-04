package gobject

// _gotk4_goMarshal is called by the GLib runtime when a closure needs to be
// invoked. The closure will be invoked with as many arguments as it can take,
// from 0 to the full amount provided by the call. If the closure asks for more
// parameters than there are to give, then a runtime panic will occur.
//
//export _gotk4_goMarshal
func _gotk4_goMarshal(
	gclosure *C.GClosure,
	retValue *C.GValue,
	nParams C.guint,
	params *C.GValue,
	invocationHint C.gpointer,
	gobject *C.GObject) {

	// Get the function value associated with this callback closure.
	box := intern.TryGet(unsafe.Pointer(gobject))
	if box == nil {
		log.Printf(
			"warning: object %s %v cannot be resurrected",
			typeFromObject(unsafe.Pointer(gobject)),
			unsafe.Pointer(gobject),
		)
		return
	}

	fs := box.Closures().Load(unsafe.Pointer(gclosure))
	if fs == nil {
		log.Printf(
			"warning: object %s %v missing closure %v",
			typeFromObject(unsafe.Pointer(gobject)),
			unsafe.Pointer(gobject),
			unsafe.Pointer(gclosure),
		)
		return
	}

	fn := fs.Func
	var skip int

	if box, ok := fn.(GeneratedClosure); ok {
		fn = box.Func
		skip++
	}

	// Fast path for an empty function.
	if fn, ok := fn.(func()); ok {
		fn()
		return
	}

	fsValue := reflect.ValueOf(fn)
	fsType := fsValue.Type()

	// Get number of parameters passed in.
	nGLibParams := int(nParams)
	nTotalParams := nGLibParams

	// Reflect may panic, so we defer recover here to re-panic with our trace.
	defer fs.TryRepanic()

	// Get number of parameters from the callback closure. If this exceeds
	// the total number of marshaled parameters, trigger a runtime panic.
	nCbParams := fsType.NumIn()
	if nCbParams > nTotalParams {
		fs.Panicf("too many closure args: have %d, max %d", nCbParams, nTotalParams)
	}

	// Create a slice of reflect.Values as arguments to call the function.
	gValues := unsafe.Slice(params, nGLibParams)
	args := make([]reflect.Value, 0, nCbParams)

	// Fill beginning of args, up to the minimum of the total number of callback
	// parameters and parameters from the glib runtime.
	for i := skip; (i-skip) < nCbParams && i < nGLibParams; i++ {
		val := goGValue(&gValues[i])

		// Parameters that are descendants of GObject come wrapped in another
		// GObject. For C applications, the default marshaller
		// (g_cclosure_marshal_VOID__VOID in gmarshal.c in the GTK glib
		// library) 'peeks' into the enclosing object and passes the wrapped
		// object to the handler. Use the Object.Cast() function
		// to emulate that for Go signal handlers.
		switch v := val.(type) {
		case *Object:
			val = v.Cast()
		case *Variant:
			if g := v.GoValue(); g != nil {
				val = g
			}
		}

		// Allow callbacks to omit the first type.
		paramType := fsType.In(i - skip)
		if i == 0 && !reflect.TypeOf(val).ConvertibleTo(paramType) {
			// The first type does not match. Allow skipping this first
			// parameter and try plugging again into the second one.
			skip++
			// Ideally, we would have a function to check if the next type
			// actually matches as well instead of this, but whatever.
			continue
		}

		rval := reflect.ValueOf(val)
		rtyp := rval.Type()
		if rtyp != paramType {
			rval = rval.Convert(paramType)
		}

		args = append(args, rval)
	}

	// Call closure with args. If the callback returns one or more values, save
	// the GValue equivalent of the first.
	rv := fsValue.Call(args)
	if retValue != nil && len(rv) > 0 {
		gv := allocateValue()
		gv.InitGoValue(rv[0].Interface())
		defer gv.unset()

		ok := C.g_value_transform(gv.native(), retValue) != 0

		if !ok {
			fs.Panicf(
				"failed to transform return value from %s to %s",
				gv.Type(), (&Value{&value{retValue}}).Type())
		}
	}
}
