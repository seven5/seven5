package client

import (
	"fmt"
	"github.com/gopherjs/jquery"
	//"honnef.co/go/js/console"
)

//CssClass represents a single CSS class defined elsewhere. This interface
//is useful in conjunction with Concorde creation of HTML.
type CssClass interface {
	ClassName() string
}

//clssClassImpl is an implementation of CssClass that has a simple string
//name.  This is useful like this:
//    var Example = CssClassImpl{"example"}
type cssClassImpl struct {
	name string
}

//className just returns the stored value of the class name, usually set
//at creation-time.
func (self cssClassImpl) ClassName() string {
	return self.name
}

//NewCssClass returns a selctor that can find .name
func NewCssClass(name string) CssClass {
	return cssClassImpl{
		name: name,
	}
}

//HtmlId represents a particular item in the DOM tree.
type HtmlId interface {
	Dom() NarrowDom
	TagName() string
	Id() string
	StyleAttribute(string) DomAttribute
	TextAttribute() DomAttribute
	DisplayAttribute() DomAttribute
	CssExistenceAttribute(clazz CssClass) DomAttribute
}

//htmlIdImpl is an implementation of HtmlId that has a fixed tag and
//identifier.  This is useful to create a selector like input#foo:
//    var SomeField = NewHtmlId("input","foo")
type htmlIdImpl struct {
	tag   string
	id    string
	t     NarrowDom
	cache map[string]DomAttribute
}

//newHtmlIdNoCheck is the same as NewHtmlId except that no check is performed
//to see if the named element exists in the DOM.  This is useful in a few cases
//where dynamically constructed HTML but should not be used by client code
//because it is error prone.
func newHtmlIdNoCheck(tag, id string) HtmlId {
	if TestMode {
		return htmlIdImpl{
			tag:   tag,
			id:    id,
			t:     newTestOps(),
			cache: make(map[string]DomAttribute),
		}
	}
	jq := jquery.NewJQuery(tag + "#" + id)

	return htmlIdImpl{
		tag:   tag,
		id:    id,
		t:     wrap(jq),
		cache: make(map[string]DomAttribute),
	}
}

//NewHtmlId returns a selctor that can find tag#id in the dom. Note that
//in production mode (with jquery) this panics if this not "hit" exactly
//one html entity.
func NewHtmlId(tag, id string) HtmlId {
	if TestMode {
		return htmlIdImpl{
			tag:   tag,
			id:    id,
			t:     newTestOps(),
			cache: make(map[string]DomAttribute),
		}
	}
	jq := jquery.NewJQuery(tag + "#" + id)
	if jq.Length != 1 {
		panic(fmt.Sprintf("probably your HTML and your code are out of sync: %s", tag+"#"+id))
	}
	return htmlIdImpl{
		tag:   tag,
		id:    id,
		t:     wrap(jq),
		cache: make(map[string]DomAttribute),
	}
}

//Dom returns a handle to the dom accessor, which varies between test and
//browser modes.
func (self htmlIdImpl) Dom() NarrowDom {
	return self.t
}

func (self htmlIdImpl) Selector() string {
	return self.TagName() + "#" + self.Id()
}

//TagName returns the stored tagname.
func (self htmlIdImpl) TagName() string {
	return self.tag
}

//Id returns the stored id.
func (self htmlIdImpl) Id() string {
	return self.id
}

func (self htmlIdImpl) cachedAttribute(name string) DomAttribute {
	return self.cache[name]
}

func (self htmlIdImpl) setCachedAttribute(name string, attr DomAttribute) {
	self.cache[name] = attr
}

func (self htmlIdImpl) getOrCreateAttribute(name string, fn func() DomAttribute) DomAttribute {
	attr := self.cachedAttribute(name)
	if attr != nil {
		return attr
	}
	attr = fn()
	self.setCachedAttribute(name, attr)
	return attr
}

//StyleAttribute returns the dom style attribute for the given name
//on the dom element selected by this object.
func (self htmlIdImpl) StyleAttribute(name string) DomAttribute {
	return self.getOrCreateAttribute("style:"+name, func() DomAttribute {
		return NewStyleAttr(name, self.t)
	})
}

//DisplayAttribute returns the dom style attribute "display" and
//expects this to be connected to a constraint returning boolean.
func (self htmlIdImpl) DisplayAttribute() DomAttribute {
	return self.getOrCreateAttribute("display", func() DomAttribute {
		return NewDisplayAttr(self.t)
	})
}

//CssExistenceAttribute returns an attribute that is hooked to
//particular CSS closs.
func (self htmlIdImpl) CssExistenceAttribute(clazz CssClass) DomAttribute {
	return self.getOrCreateAttribute("cssexist:"+clazz.ClassName(), func() DomAttribute {
		return NewCssExistenceAttr(self.t, clazz)
	})
}

//TextAttribute returns the dom text attribute for the dom element
//selected by this object.
func (self htmlIdImpl) TextAttribute() DomAttribute {
	return self.getOrCreateAttribute("text", func() DomAttribute {
		return NewTextAttr(self.t)
	})
}
