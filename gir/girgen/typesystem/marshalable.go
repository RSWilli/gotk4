package typesystem

// Marshalable describes a
type Marshalable interface {
	Type
	GLibGetType() string
}

type Marshaler struct {
	GlibGetType string
}

// CanMarshal implements Marshalable.
func (m Marshaler) CanMarshal() bool {
	return m.GlibGetType != ""
}

// GLibGetType implements Marshalable.
func (m Marshaler) GLibGetType() string {
	return m.GlibGetType
}

func newDefaultMarshaler(gettype string) Marshaler {
	return Marshaler{
		GlibGetType: gettype,
	}
}
