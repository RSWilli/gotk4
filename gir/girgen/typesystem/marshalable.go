package typesystem

// Marshalable describes a
type Marshalable interface {
	Type
	GLibGetType() string

	// Value is used to know from where we need to import GValue for the marshaler/SetValue method
	Value() CouldBeForeign[*Record]
}

type Marshaler struct {
	GlibGetType string

	Gvalue CouldBeForeign[*Record]
}

// CanMarshal implements Marshalable.
func (m Marshaler) CanMarshal() bool {
	return m.GlibGetType != ""
}

// GLibGetType implements Marshalable.
func (m Marshaler) GLibGetType() string {
	return m.GlibGetType
}

// Value implements Marshalable.
func (m Marshaler) Value() CouldBeForeign[*Record] {
	return m.Gvalue
}

func newDefaultMarshaler(gettype string, value CouldBeForeign[*Record]) Marshaler {
	return Marshaler{
		GlibGetType: gettype,
		Gvalue:      value,
	}
}
