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
	Select() jquery.JQuery
	StyleAttribute(string) DomAttribute
	TextAttribute() DomAttribute
	DisplayAttribute() DomAttribute
	Event(EventName, EventFunc)
	CssExistenceAttribute(clazz CssClass)
	Val() string
	SetText(string)
}

//htmlIdImpl is an implementation of HtmlId that has a fixed tag and
//identifier.  This is useful to create a selector like input#foo:
//    var SomeField = NewHtmlId("input","foo")
type htmlIdImpl struct {
	tag string
	id  string
}

//NewHtmlId returns a selctor that can find tag#id in the dom.
func NewHtmlId(tag, id string) HtmlId {
	return &htmlIdImpl{
		tag: tag,
		id:  id,
	}
}

//TagName returns the stored tagname.
func (self *htmlIdImpl) TagName() string {
	return self.tag
}

//Id returns the stored id.
func (self *htmlIdImpl) Id() string {
	return self.id
}

func (self *htmlIdImpl) Select() jquery.JQuery {
	return jquery.NewJQuery(self.TagName() + "#" + self.Id())
}

//Val returns the value of an input field.  Note that this probably
//will not do what you want if the object in question is not an
//input or textarea.
func (self *htmlIdImpl) Val() string {
	return self.Select().Val()
}

func (self *htmlIdImpl) SetText(s string) {
	self.Select().SetText(s)
}

//StyleAttribute returns the dom style attribute for the given name
//on the dom element selected by this object.
func (self *htmlIdImpl) StyleAttribute(name string) DomAttribute {
	return NewStyleAttr(name, self.Select())
}

//DisplayAttribute returns the dom style attribute "display" and
//expects this to be connected to a constraint returning boolean.
func (self *htmlIdImpl) DisplayAttribute() DomAttribute {
	return NewDisplayAttr(self.Select())
}

//TextAttribute returns the dom text attribute for the dom element
//selected by this object.
func (self *htmlIdImpl) TextAttribute() DomAttribute {
	return NewTextAttr(self.Select())
}

//Event hooks an event func to the event named.
//XXX should probably define the constants for event names as their own type
func (self *htmlIdImpl) Event(n EventName, f EventFunc) {
	h := &eventHandler{n, self.Select(), f}
	h.register()
}

//Selector is a way to pick 0 or more dom objects from the document.
type Selector interface {
	Select() jquery.JQuery
}
