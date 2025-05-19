package gir

import (
	"encoding"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

// https://gitlab.gnome.org/GNOME/gobject-introspection/-/blob/HEAD/docs/gir-1.2.rnc

type Alias struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 alias"`

	Name  string `xml:"name,attr"`
	CType string `xml:"http://www.gtk.org/introspection/c/1.0 type,attr"`

	InfoAttrs
	InfoElements

	Type Type `xml:"http://www.gtk.org/introspection/core/1.0 type"`
}

type AnyType struct {
	// Possible variants.
	Type  *Type  `xml:"http://www.gtk.org/introspection/core/1.0 type"`
	Array *Array `xml:"http://www.gtk.org/introspection/core/1.0 array"`
}

type Annotation struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 attribute"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value,attr"`
}

type Array struct {
	XMLName        xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 array"`
	Name           string   `xml:"name,attr"`
	CType          string   `xml:"http://www.gtk.org/introspection/c/1.0 type,attr"`
	Length         *int     `xml:"length,attr"` // ix of .Parameters
	ZeroTerminated *bool    `xml:"zero-terminated,attr"`
	FixedSize      int      `xml:"fixed-size,attr"`
	Introspectable bool     `xml:"introspectable,attr"`
	// Type is the array's inner type.
	Type *Type `xml:"http://www.gtk.org/introspection/core/1.0 type"`
}

// IsZeroTerminated returns true if the Array is zero-terminated. It accounts
// for edge cases of the structure.
func (a Array) IsZeroTerminated() bool {
	return a.Name == "" && (a.ZeroTerminated == nil || *a.ZeroTerminated)
}

type Bitfield struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 bitfield"`

	Name         string `xml:"name,attr"` // Go case
	CType        string `xml:"http://www.gtk.org/introspection/c/1.0 type,attr"`
	GLibTypeName string `xml:"http://www.gtk.org/introspection/glib/1.0 type-name,attr"`
	GLibGetType  string `xml:"http://www.gtk.org/introspection/glib/1.0 get-type,attr"`

	Members   []Member   `xml:"http://www.gtk.org/introspection/core/1.0 member"`
	Functions []Function `xml:"http://www.gtk.org/introspection/core/1.0 function"`

	InfoAttrs
	InfoElements
}

type Boxed struct{}

type CInclude struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/c/1.0 include"`
	Name    string   `xml:"name,attr"`
}

type CallableAttrs struct {
	Name        string       `xml:"name,attr"`
	CIdentifier string       `xml:"http://www.gtk.org/introspection/c/1.0 identifier,attr"`
	ShadowedBy  string       `xml:"shadowed-by,attr"`
	Shadows     string       `xml:"shadows,attr"`
	Throws      bool         `xml:"throws,attr"`
	MovedTo     string       `xml:"moved-to,attr"`
	Parameters  *Parameters  `xml:"http://www.gtk.org/introspection/core/1.0 parameters"`
	ReturnValue *ReturnValue `xml:"http://www.gtk.org/introspection/core/1.0 return-value"`
	InfoAttrs
	InfoElements
}

type Callback struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 callback"`
	CType   string   `xml:"http://www.gtk.org/introspection/c/1.0 type,attr"`
	CallableAttrs
}

type Class struct {
	XMLName  xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 class"`
	Name     string   `xml:"name,attr"`
	Parent   string   `xml:"parent,attr"`
	Abstract bool     `xml:"abstract,attr"`

	CType         string `xml:"http://www.gtk.org/introspection/c/1.0 type,attr"`
	CSymbolPrefix string `xml:"http://www.gtk.org/introspection/c/1.0 symbol-prefix,attr"`

	GLibTypeName   string `xml:"http://www.gtk.org/introspection/glib/1.0 type-name,attr"`
	GLibGetType    string `xml:"http://www.gtk.org/introspection/glib/1.0 get-type,attr"`
	GLibTypeStruct string `xml:"http://www.gtk.org/introspection/glib/1.0 type-struct,attr"`

	InfoAttrs
	InfoElements

	Functions      []Function      `xml:"http://www.gtk.org/introspection/core/1.0 function"`
	Implements     []Implements    `xml:"http://www.gtk.org/introspection/core/1.0 implements"`
	Constructors   []Constructor   `xml:"http://www.gtk.org/introspection/core/1.0 constructor"`
	Methods        []Method        `xml:"http://www.gtk.org/introspection/core/1.0 method"`
	VirtualMethods []VirtualMethod `xml:"http://www.gtk.org/introspection/core/1.0 virtual-method"`
	Fields         []Field         `xml:"http://www.gtk.org/introspection/core/1.0 field"`
	Signals        []Signal        `xml:"http://www.gtk.org/introspection/glib/1.0 signal"`
}

type Constant struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 constant"`
	Name    string   `xml:"name,attr"`
	Value   string   `xml:"value,attr"`
	CType   string   `xml:"http://www.gtk.org/introspection/c/1.0 type,attr"`

	Type Type `xml:"http://www.gtk.org/introspection/core/1.0 type"`

	InfoAttrs
	InfoElements
}

type Constructor struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 constructor"`
	CallableAttrs
}

type Doc struct {
	XMLName  xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 doc"`
	Filename string   `xml:"filename,attr"`
	String   string   `xml:",innerxml"`
	Line     int      `xml:"line,attr"`
}

type DocDeprecated struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 doc-deprecated"`
	String  string   `xml:",innerxml"`
}

type DocElements struct {
	Doc            *Doc
	DocDeprecated  *DocDeprecated
	SourcePosition *SourcePosition
}

type Enum struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 enumeration"`

	Name            string `xml:"name,attr"` // Go case
	CType           string `xml:"http://www.gtk.org/introspection/c/1.0 type,attr"`
	GLibTypeName    string `xml:"http://www.gtk.org/introspection/glib/1.0 type-name,attr"`
	GLibGetType     string `xml:"http://www.gtk.org/introspection/glib/1.0 get-type,attr"`
	GLibErrorDomain string `xml:"http://www.gtk.org/introspection/glib/1.0 error-domain,attr"`

	Members   []Member   `xml:"http://www.gtk.org/introspection/core/1.0 member"`
	Functions []Function `xml:"http://www.gtk.org/introspection/core/1.0 function"`

	InfoAttrs
	InfoElements
}

type Field struct {
	XMLName  xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 field"`
	Name     string   `xml:"name,attr"`
	Bits     int      `xml:"bits,attr"`
	Private  bool     `xml:"private,attr"`
	Writable bool     `xml:"writable,attr"` // default false
	Readable *bool    `xml:"readable,attr"` // default true
	AnyType
	Callback *Callback
	Doc      *Doc
}

// IsReadable returns true if the field is readable.
func (f Field) IsReadable() bool {
	return f.Readable == nil || *f.Readable
}

type Function struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 function"`
	CallableAttrs
}

type Implements struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 implements"`
	Name    string   `xml:"name,attr"`
}

type Include struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 include"`
	Name    string   `xml:"name,attr"`
	Version Version  `xml:"version,attr"`
}

type InfoAttrs struct {
	Introspectable    *bool   `xml:"introspectable,attr"` // default true
	Deprecated        bool    `xml:"deprecated,attr"`
	DeprecatedVersion Version `xml:"deprecated-version,attr"`
	Version           Version `xml:"version,attr"`
	Stability         string  `xml:"stability,attr"`
}

func (inf InfoAttrs) GetInfoAttrs() InfoAttrs {
	return inf
}

func ParseVersion(str string) (Version, error) {
	var v Version

	err := v.UnmarshalText([]byte(str))

	if err != nil {
		return v, err
	}

	return v, nil
}

type Version struct {
	Major int
	Minor int
	Patch int
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// LessEqual implements the less than or equal relation
func (v Version) LessEqual(other Version) bool {
	if v.Major < other.Major {
		return true
	}
	if v.Major > other.Major {
		return false
	}
	if v.Minor < other.Minor {
		return true
	}
	if v.Minor > other.Minor {
		return false
	}

	return v.Patch <= other.Patch
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (v *Version) UnmarshalText(in []byte) error {
	text := string(in)

	if len(text) == 0 {
		return nil
	}

	parts := strings.Split(text, ".")

	if len(parts) > 3 {
		return fmt.Errorf("invalid version")
	}

	var major int
	var minor int
	var patch int
	var err error

	if len(parts) > 0 {
		major, err = strconv.Atoi(parts[0])
		if err != nil {
			return fmt.Errorf(`error while parsing major version "%s" of "%s": %w`, parts[0], text, err)
		}
	}
	if len(parts) > 1 {
		minor, err = strconv.Atoi(parts[1])
		if err != nil {
			return fmt.Errorf(`error while parsing minor version "%s" of "%s": %w`, parts[1], text, err)
		}
	}
	if len(parts) > 2 && parts[2] != "" {
		patch, err = strconv.Atoi(parts[2])
		if err != nil {
			return fmt.Errorf(`error while parsing patch version "%s" of "%s": %w`, parts[2], text, err)
		}
	}

	*v = Version{
		Major: major,
		Minor: minor,
		Patch: patch,
	}

	return nil
}

var _ encoding.TextUnmarshaler = &Version{}

// IsIntrospectable returns true if the InfoAttrs indicates that the type is
// introspectable.
func (inf InfoAttrs) IsIntrospectable() bool {
	return inf.Introspectable == nil || *inf.Introspectable
}

type InfoElements struct {
	DocElements
	Annotations []Annotation `xml:"http://www.gtk.org/introspection/core/1.0 attribute"`
}

func (inf InfoElements) GetInfoElements() InfoElements {
	return inf
}

type InstanceParameter struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 instance-parameter"`
	ParameterAttrs
}

type Interface struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 interface"`
	Name    string   `xml:"name,attr"`

	CType         string `xml:"http://www.gtk.org/introspection/c/1.0 type,attr"`
	CSymbolPrefix string `xml:"http://www.gtk.org/introspection/c/1.0 symbol-prefix,attr"`

	GLibTypeName   string `xml:"http://www.gtk.org/introspection/glib/1.0 type-name,attr"`
	GLibGetType    string `xml:"http://www.gtk.org/introspection/glib/1.0 get-type,attr"`
	GLibTypeStruct string `xml:"http://www.gtk.org/introspection/glib/1.0 type-struct,attr"`

	Functions      []Function      `xml:"http://www.gtk.org/introspection/core/1.0 function"`
	Methods        []Method        `xml:"http://www.gtk.org/introspection/core/1.0 method"`
	VirtualMethods []VirtualMethod `xml:"http://www.gtk.org/introspection/core/1.0 virtual-method"`
	Prerequisites  []Prerequisite  `xml:"http://www.gtk.org/introspection/core/1.0 prerequisite"`
	Signals        []Signal        `xml:"http://www.gtk.org/introspection/glib/1.0 signal"`

	InfoAttrs
	InfoElements
}

type Member struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 member"`

	Names       []xml.Attr `xml:"name,attr"`
	Value       string     `xml:"value,attr"`
	CIdentifier string     `xml:"http://www.gtk.org/introspection/c/1.0 identifier,attr"`
	GLibNick    string     `xml:"http://www.gtk.org/introspection/glib/1.0 nick,attr"`

	InfoAttrs
	InfoElements
}

func (m Member) Name() string {
	return m.nameAttr(xml.Name{Local: "name"})
}

func (m Member) GLibName() string {
	return m.nameAttr(xml.Name{Space: "http://www.gtk.org/introspection/glib/1.0", Local: "name"})
}

func (m Member) nameAttr(name xml.Name) string {
	for _, attr := range m.Names {
		if attr.Name == name {
			return attr.Value
		}
	}
	return ""
}

type Method struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 method"`
	CallableAttrs
}

type Namespace struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 namespace"`

	Name                string  `xml:"name,attr"`
	Version             Version `xml:"version,attr"`
	CIdentifierPrefixes string  `xml:"http://www.gtk.org/introspection/c/1.0 identifier-prefixes,attr"`
	CSymbolPrefixes     string  `xml:"http://www.gtk.org/introspection/c/1.0 symbol-prefixes,attr"`
	Prefix              string  `xml:"http://www.gtk.org/introspection/c/1.0 prefix,attr"`
	SharedLibrary       string  `xml:"shared-library,attr"`

	Aliases     []Alias      `xml:"http://www.gtk.org/introspection/core/1.0 alias"`
	Classes     []Class      `xml:"http://www.gtk.org/introspection/core/1.0 class"`
	Interfaces  []Interface  `xml:"http://www.gtk.org/introspection/core/1.0 interface"`
	Records     []Record     `xml:"http://www.gtk.org/introspection/core/1.0 record"`
	Enums       []Enum       `xml:"http://www.gtk.org/introspection/core/1.0 enumeration"`
	Functions   []Function   `xml:"http://www.gtk.org/introspection/core/1.0 function"`
	Unions      []Union      `xml:"http://www.gtk.org/introspection/core/1.0 union"`
	Bitfields   []Bitfield   `xml:"http://www.gtk.org/introspection/core/1.0 bitfield"`
	Callbacks   []Callback   `xml:"http://www.gtk.org/introspection/core/1.0 callback"`
	Constants   []Constant   `xml:"http://www.gtk.org/introspection/core/1.0 constant"`
	Annotations []Annotation `xml:"http://www.gtk.org/introspection/core/1.0 attribute"`
	Boxeds      []Boxed      `xml:"http://www.gtk.org/introspection/core/1.0 boxed"`
}

type Package struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 package"`
	Name    string   `xml:"name,attr"`
}

type Parameter struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 parameter"`
	ParameterAttrs
}

type ParameterAttrs struct {
	Name            string `xml:"name,attr"`
	Direction       string `xml:"direction,attr"`
	Scope           string `xml:"scope,attr"`
	Closure         *int   `xml:"closure,attr"`
	Destroy         *int   `xml:"destroy,attr"`
	CallerAllocates bool   `xml:"caller-allocates,attr"`
	Skip            bool   `xml:"skip,attr"`
	Optional        bool   `xml:"optional,attr"`
	Nullable        bool   `xml:"nullable,attr"`

	TransferOwnership
	AnyType
	Doc *Doc
}

type Parameters struct {
	XMLName           xml.Name           `xml:"http://www.gtk.org/introspection/core/1.0 parameters"`
	InstanceParameter *InstanceParameter `xml:"http://www.gtk.org/introspection/core/1.0 instance-parameter"`
	Parameters        []Parameter        `xml:"http://www.gtk.org/introspection/core/1.0 parameter"`
}

type Prerequisite struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 prerequisite"`
	Name    string   `xml:"name,attr"`
}

type Property struct{}

type Record struct {
	XMLName              xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 record"`
	Name                 string   `xml:"name,attr"`
	CType                string   `xml:"http://www.gtk.org/introspection/c/1.0 type,attr"`
	GLibTypeName         string   `xml:"http://www.gtk.org/introspection/glib/1.0 type-name,attr"`
	GLibGetType          string   `xml:"http://www.gtk.org/introspection/glib/1.0 get-type,attr"`
	CSymbolPrefix        string   `xml:"http://www.gtk.org/introspection/c/1.0 symbol-prefix,attr"`
	GLibIsGTypeStructFor string   `xml:"http://www.gtk.org/introspection/glib/1.0 is-gtype-struct-for,attr"`
	Disguised            bool     `xml:"disguised,attr"`
	Foreign              bool     `xml:"foreign,attr"`

	Fields       []Field       `xml:"http://www.gtk.org/introspection/core/1.0 field"`
	Functions    []Function    `xml:"http://www.gtk.org/introspection/core/1.0 function"`
	Unions       []Union       `xml:"http://www.gtk.org/introspection/core/1.0 union"`
	Methods      []Method      `xml:"http://www.gtk.org/introspection/core/1.0 method"`
	Constructors []Constructor `xml:"http://www.gtk.org/introspection/core/1.0 constructor"`
	Properties   []Property    `xml:"http://www.gtk.org/introspection/core/1.0 property"`

	InfoAttrs
	InfoElements
}

type ReturnValue struct {
	XMLName        xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 return-value"`
	Scope          string   `xml:"scope,attr"`
	Closure        *int     `xml:"closure,attr"`
	Destroy        *int     `xml:"destroy,attr"`
	Introspectable bool     `xml:"introspectable,attr"`
	Nullable       bool     `xml:"nullable,attr"`
	Skip           bool     `xml:"skip,attr"`
	AllowNone      bool     `xml:"allow-none,attr"`
	TransferOwnership
	DocElements
	AnyType
}

type Signal struct {
	XMLName   xml.Name   `xml:"http://www.gtk.org/introspection/glib/1.0 signal"`
	Name      string     `xml:"name,attr"`
	Detailed  bool       `xml:"detailed,attr"`
	When      SignalWhen `xml:"when,attr"`
	Action    bool       `xml:"action,attr"`
	NoHooks   bool       `xml:"no-hooks,attr"`
	NoRecurse bool       `xml:"no-recurse,attr"`
	InfoElements
	Parameters  *Parameters  `xml:"http://www.gtk.org/introspection/core/1.0 parameters"`
	ReturnValue *ReturnValue `xml:"http://www.gtk.org/introspection/core/1.0 return-value"`
}

type SignalWhen string

const (
	SignalWhenFirst   = "first"
	SignalWhenLast    = "last"
	SignalWhenCleanup = "cleanup"
)

type SourcePosition struct {
	XMLName  xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 source-position"`
	Filename string   `xml:"filename,attr"`
	Line     int      `xml:"line,attr"`
	Column   int      `xml:"column,attr"`
}

type TransferOwnership struct {
	TransferOwnership string `xml:"transfer-ownership,attr"`
}

type Type struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 type"`

	Name           string `xml:"name,attr"`
	CType          string `xml:"http://www.gtk.org/introspection/c/1.0 type,attr"`
	Introspectable *bool  `xml:"introspectable,attr"`

	DocElements
	// Types is the type's inner types. e.g. generic types of container types
	Types []Type `xml:"http://www.gtk.org/introspection/core/1.0 type"`
}

func (typ Type) IsIntrospectable() bool {
	return typ.Introspectable == nil || *typ.Introspectable
}

type Union struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 union"`

	Name          string `xml:"name,attr"` // Go case
	CType         string `xml:"http://www.gtk.org/introspection/c/1.0 type,attr"`
	CSymbolPrefix string `xml:"http://www.gtk.org/introspection/c/1.0 symbol-prefix,attr"`
	GLibTypeName  string `xml:"http://www.gtk.org/introspection/glib/1.0 type-name,attr"`
	GLibGetType   string `xml:"http://www.gtk.org/introspection/glib/1.0 get-type,attr"`

	InfoAttrs
	InfoElements

	Fields       []Field       `xml:"http://www.gtk.org/introspection/core/1.0 field"`
	Constructors []Constructor `xml:"http://www.gtk.org/introspection/core/1.0 constructor"`
	Methods      []Method      `xml:"http://www.gtk.org/introspection/core/1.0 method"`
	Functions    []Function    `xml:"http://www.gtk.org/introspection/core/1.0 function"`
	Records      []Record      `xml:"http://www.gtk.org/introspection/core/1.0 record"`
}

type VarArgs struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 varargs"`
}

type VirtualMethod struct {
	XMLName xml.Name `xml:"http://www.gtk.org/introspection/core/1.0 virtual-method"`

	Invoker string `xml:"invoker,attr"`
	CallableAttrs
}
