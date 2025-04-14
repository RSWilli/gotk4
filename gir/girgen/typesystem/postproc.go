package typesystem

import (
	"fmt"

	"github.com/diamondburned/gotk4/gir"
)

type PostProcessor func(*Registry) error

func MarkAsManuallyExtended(namespace string, majorversion int, typ string) PostProcessor {
	return func(r *Registry) error {
		ns := r.FindNamespace(versionedName{
			name: namespace,
			version: gir.Version{
				Major: majorversion,
			},
		})
		if ns == nil {
			return fmt.Errorf("could not find namespace %s", namespace)
		}

		t := ns.FindLocalTypeByGIRName(typ)

		if t == nil {
			return fmt.Errorf("could not find type %s in namespace %s", typ, namespace)
		}

		switch t := t.(type) {
		case *Class:
			t.ManuallyExtended = true
		case *Interface:
			t.ManuallyExtended = true
		default:
			return fmt.Errorf("type %s is not a class or interface, but instead %T", typ, t)
		}

		return nil
	}

}
