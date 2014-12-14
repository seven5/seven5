package client

import (
	"fmt"
	"github.com/gopherjs/jquery"
	//"honnef.co/go/js/console"
)

const (
	SEVEN5_DATA       = "seven5_%s"
	CONSTRAINT_MARKER = "constraint"
)

//DomAttribute is a constrainable entity that references a portion of the dom.
//The referenced dom element may be either a fixed portion of the dom or
//part of a dynamically created (sub)tree.
type DomAttribute interface {
	Attribute
	Id() string
}

//domAttr is thin layer over attribute that has two extra features.
//First it uses the jquery storage to store if a value is constrained
//or not, so as to prevent programming errors if the same object is
//referenced in two ways. Second, it knows how to read from the dom
//values and write to the dom values on update.  This latter read/write
//functionality is delegated to specific subtypes.
type domAttr struct {
	attr *AttributeImpl
	t    NarrowDom
	id   string
	set  SideEffectFunc
	get  ValueFunc
}

type styleAttr struct {
	*domAttr
	name string
}

func newDomAttr(t NarrowDom, id string, g ValueFunc, s SideEffectFunc) *domAttr {
	result := &domAttr{
		t:  t,
		id: id,
	}
	result.attr = NewAttribute(EAGER, g, s)
	return result
}

func (self *domAttr) getData() string {
	return self.t.Data(fmt.Sprintf(SEVEN5_DATA, self.id))
}

func (self *domAttr) removeData() {
	self.t.RemoveData(fmt.Sprintf(SEVEN5_DATA, self.id))
}

func (self *domAttr) setData(v string) {
	self.t.SetData(fmt.Sprintf(SEVEN5_DATA, self.id), v)
}

//verifyConstraint tests to see if there is already a constraint or not
//on this dom object. Pass false to verify that there is not a current
//constraint.
func (self *domAttr) verifyConstraint(b bool) {
	s := self.getData()
	if b {
		if s == "" {
			panic(fmt.Sprintf("expected to have constraint on %s but did not!", self.id))
		}
		if s != CONSTRAINT_MARKER {
			panic(fmt.Sprintf("expected to find constraint marker but found %s", s))
		}
	} else {
		if s != "" {
			panic(fmt.Sprintf("did not expect to have constraint on %s but found one: %s", self.id, s))
		}
	}
}

func (self *domAttr) Attach(c Constraint) {
	self.verifyConstraint(false)
	self.attr.Attach(c)
	self.setData(CONSTRAINT_MARKER)
}

func (self *domAttr) Detach() {
	self.verifyConstraint(true)
	self.attr.Detach()
	self.removeData()
}
func (self *domAttr) Demand() Equaler {
	return self.attr.Demand()
}
func (self *domAttr) SetEqualer(e Equaler) {
	self.attr.SetEqualer(e)
}
func (self *domAttr) SetDebugName(n string) {
	self.attr.SetDebugName(n)
}
func (self *domAttr) Id() string {
	return self.id
}

//NewStyleAttribute returns a dom attribute connected to the css
//style value named by n and selected by the selector j.  This results
//in a call to SetCSS and uses fmt.Sprintf() to format it's value,
//so the constraint result may be any type.  This the lower level
//interface, most users will probably prefer to use the
//StyleAttrBuilder interface.
func NewStyleAttr(n string, t NarrowDom) DomAttribute {
	result := &styleAttr{name: n}
	result.domAttr = newDomAttr(t, fmt.Sprintf("style:%s", n), result.get, result.set)
	return result
}

func (self *styleAttr) get() Equaler {
	if self.t.Css(self.name) == "undefined" {
		return StringEqualer{""}
	}
	return StringEqualer{self.t.Css(self.name)}
}

func (self *styleAttr) set(e Equaler) {
	self.t.SetCss(self.name, fmt.Sprintf("%s", e))
}

type textAttr struct {
	*domAttr
}

//NewTextAttr returns a dom attribute connected to the text property
//of the elements matched by j.   Most users will probably prefer to
//use the TextBuilder API, this is the lower level access to the raw
//DomAttribute.  This attribute uses fmt.Sprintf() to compute the final
//text value written, so the constraint result can be any type.
func NewTextAttr(t NarrowDom) DomAttribute {
	result := &textAttr{}
	result.domAttr = newDomAttr(t, "text", result.get, result.set)
	return result
}

func (self *textAttr) get() Equaler {
	return StringEqualer{self.t.Text()}
}

func (self *textAttr) set(e Equaler) {
	self.t.SetText(fmt.Sprintf("%v", e))
}

type htmlAttrAttr struct {
	*domAttr
	name htmlAttrName
}

//NewHtmlAttrAttr provides an interface to the dom "attribute" (in
//the constraint sense) for the given html attribute name, on the
//elements that are matched by j.  Most users will probably prefer
//to use the HtmlAttrBuilder interface, this is the lower level access
//to the raw DomAttribute().  This attribute uses fmt.Sprintf()
//to compute the final value assigned to the dom element, so the
//constraint result can be any type.
func NewHtmlAttrAttr(t NarrowDom, a htmlAttrName) DomAttribute {
	result := &htmlAttrAttr{name: a}
	result.domAttr = newDomAttr(t, "attr:"+string(a), result.get, result.set)
	return result
}

func (self *htmlAttrAttr) get() Equaler {
	return StringEqualer{self.t.Attr(string(self.name))}
}

func (self *htmlAttrAttr) set(e Equaler) {
	self.t.SetAttr(string(self.name), fmt.Sprintf("%v", e))
}

type propAttr struct {
	*domAttr
	name propName
}

//NewPropAttr provides an interface to the dom attribute for the property
//named, for the elements that are matched by j.  This is primarily
//useful for things that have true/false state (such checked, selected,
//or disabled state) so it expects a boolean attribute.
func NewPropAttr(t NarrowDom, n propName) DomAttribute {
	result := &propAttr{name: n}
	result.domAttr = newDomAttr(t, "prop:"+string(n), result.get, result.set)
	return result
}

func (self *propAttr) get() Equaler {
	b := self.t.Prop(string(self.name))
	return BoolEqualer{b}
}

func (self *propAttr) set(e Equaler) {
	self.t.SetProp(string(self.name), e.(BoolEqualer).B)
}

type htmlAttrName string
type propName string

func newAttrName(s string) htmlAttrName {
	return htmlAttrName(s)
}

func newPropName(s string) propName {
	return propName(s)
}

var (
	REL         = newAttrName("rel")
	LINK        = newAttrName("link")
	TYPE        = newAttrName("type")
	PLACEHOLDER = newAttrName("placeholder")
	HREF        = newAttrName("href")
	SRC         = newAttrName("src")
	WIDTH       = newAttrName("width")
	HEIGHT      = newAttrName("height")
	VALUE       = newAttrName("value")
	SIZE        = newAttrName("size")
	MAXLENGTH   = newAttrName("maxlength")

	CHECKED  = newPropName("checked")
	SELECTED = newPropName("selected")
	DISABLED = newPropName("disabled")
)

type cssExistenceAttr struct {
	*domAttr
	clazz CssClass
}

//NewCssExistenceAttr returns a dom attribute that should be computed
//via a constraint yielding a boolean (BoolEqualer).  If the boolean
//is true, the css class provided is attached to the elements that match
//j. If the boolean value is provided, the css class is removed. Most
//users will probably prefer to use the CssExistenceBuilder interface,
//this is the lower level access to the dom attribute.
func NewCssExistenceAttr(t NarrowDom, clazz CssClass) DomAttribute {
	result := &cssExistenceAttr{clazz: clazz}
	result.domAttr = newDomAttr(t, "cssclass:"+clazz.ClassName(), result.get, result.set)
	return result
}

func (self *cssExistenceAttr) get() Equaler {
	//print("get existenc attribute", self.clazz)
	return BoolEqualer{self.t.HasClass(self.clazz.ClassName())}
}

func (self *cssExistenceAttr) set(e Equaler) {
	//print("set existence attribute", self.clazz.ClassName(), e.(BoolEqualer).B)
	if e.(BoolEqualer).B {
		self.t.AddClass(self.clazz.ClassName())
	} else {
		self.t.RemoveClass(self.clazz.ClassName())
	}
}

//NewDisplayAttr returns a dom element that is connected to the css
//"display" attribute.  This is a special case of NewStyleAttribute
//that understands that a boolean can be used to display (true) or
//hide a given dom element.  This is the lower level interface and
//most users will prefer the DisplayAttrBuilder or DisplayAttribute
//calls.
func NewDisplayAttr(t NarrowDom) DomAttribute {
	result := &styleAttr{name: "display"}
	result.domAttr = newDomAttr(t, "style:display", result.getDisplay,
		result.setDisplay)
	return result
}

func (self *styleAttr) getDisplay() Equaler {
	if self.t.Css("display") == "undefined" {
		return BoolEqualer{true}
	}
	return BoolEqualer{self.t.Css("display") != "none"}
}

func (self *styleAttr) setDisplay(e Equaler) {
	if e.(BoolEqualer).B {
		self.t.SetCss("display", "")
	} else {
		self.t.SetCss("display", "none")
	}
}

//NewValueAttr creates a _source_ attribute from a value in the DOM.  In other
//words you use this as a constraint INPUT not as a value output (such as
//with NewDisplayAttr or NewCssExistenceAttr).  This is useful for using the
//value of input fields as a source in constraint graph.  This should probably
//be used only with tags of type INPUT although this is not enforced.
func NewValueAttr(t NarrowDom) Attribute {
	result := NewAttribute(VALUE_ONLY, func() Equaler {
		return StringEqualer{S: t.Val()}
	}, nil)

	t.On(INPUT_EVENT, func(e jquery.Event) {
		result.markDirty()
	})
	return result
}
