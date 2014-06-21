package client

import (
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
	TagName() string
	Id() string
	StyleAttribute(string) DomAttribute
	TextAttribute() DomAttribute
	DisplayAttribute() DomAttribute
	Event(EventName, EventFunc)
	CssExistenceAttribute(clazz CssClass) DomAttribute
	Val() string
	SetText(string)
}

//htmlIdImpl is an implementation of HtmlId that has a fixed tag and
//identifier.  This is useful to create a selector like input#foo:
//    var SomeField = NewHtmlId("input","foo")
type htmlIdImpl struct {
	tag   string
	id    string
	jq    jquery.JQuery
	cache map[string]DomAttribute
}

//NewHtmlId returns a selctor that can find tag#id in the dom.
func NewHtmlId(tag, id string) HtmlId {
	return htmlIdImpl{
		tag:   tag,
		id:    id,
		jq:    jquery.NewJQuery(tag + "#" + id),
		cache: make(map[string]DomAttribute),
	}
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

//Val returns the value of an input field.  Note that this probably
//will not do what you want if the object in question is not an
//input or textarea.
func (self htmlIdImpl) Val() string {
	return self.jq.Val()
}

//SetText puts the text provided into the tag.
func (self htmlIdImpl) SetText(s string) {
	self.jq.SetText(s)
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
		return NewStyleAttr(name, self.jq)
	})
}

//DisplayAttribute returns the dom style attribute "display" and
//expects this to be connected to a constraint returning boolean.
func (self htmlIdImpl) DisplayAttribute() DomAttribute {
	return self.getOrCreateAttribute("display", func() DomAttribute {
		return NewDisplayAttr(self.jq)
	})
}

//CssExistenceAttribute returns an attribute that is hooked to
//particular CSS closs.
func (self htmlIdImpl) CssExistenceAttribute(clazz CssClass) DomAttribute {
	return self.getOrCreateAttribute("cssexist:"+clazz.ClassName(), func() DomAttribute {
		return NewCssExistenceAttr(self.jq, clazz)
	})
}

//TextAttribute returns the dom text attribute for the dom element
//selected by this object.
func (self htmlIdImpl) TextAttribute() DomAttribute {
	return self.getOrCreateAttribute("text", func() DomAttribute {
		return NewTextAttr(self.jq)
	})
}

//Event hooks an event func to the event named.
//XXX should probably define the constants for event names as their own type
func (self htmlIdImpl) Event(n EventName, f EventFunc) {
	h := &eventHandler{n, self.jq, f}
	h.register()
}
