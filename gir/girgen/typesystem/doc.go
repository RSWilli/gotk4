package typesystem

import "github.com/diamondburned/gotk4/gir"

type Documented interface {
	Documentation() Doc
}

type Doc struct {
	Doc           string
	DocDeprecated string
	Deprecated    bool
}

func (d Doc) Documentation() Doc {
	return d
}

func NewSimpleDoc(girdoc *gir.Doc) Doc {
	var doc string

	if girdoc != nil {
		doc = girdoc.String
	}

	return Doc{
		Doc: doc,
	}
}

func NewDoc(attrs *gir.InfoAttrs, elements *gir.InfoElements) Doc {
	var doc string
	var docDeprecated string
	var deprecated bool

	if attrs != nil {
		deprecated = attrs.Deprecated
	}

	if elements != nil && elements.Doc != nil {
		doc = elements.Doc.String
	}

	if elements != nil && elements.DocDeprecated != nil {
		doc = elements.DocDeprecated.String
	}

	return Doc{
		Doc:           doc,
		DocDeprecated: docDeprecated,
		Deprecated:    deprecated,
	}
}

type ParamDoc struct {
	Doc      string
	Name     string
	Optional bool
	Nullable bool
}

func NewParamDoc(attrs gir.ParameterAttrs) ParamDoc {
	return ParamDoc{}
}

func NewReturnDoc(attrs *gir.ReturnValue) ParamDoc {
	return ParamDoc{}
}
