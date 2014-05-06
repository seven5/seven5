package concorde

import (
	"fmt"
	"github.com/gopherjs/jquery"
	//"honnef.co/go/js/console"
)

const (
	CONCORDE_DATA     = "concorde_%s"
	CONSTRAINT_MARKER = "constraint"
)

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
	j    jquery.JQuery
	id   string
	set  SideEffectFunc
	get  ValueFunc
}

type styleAttr struct {
	*domAttr
	name string
}

func newDomAttr(j jquery.JQuery, id string, g ValueFunc, s SideEffectFunc) *domAttr {
	result := &domAttr{
		j:  j,
		id: id,
	}
	result.attr = NewAttribute(EAGER, g, s)
	return result
}

func (self *domAttr) getData() string {
	d := self.j.Data(fmt.Sprintf(CONCORDE_DATA, self.id))
	if d == nil {
		return ""
	}
	return d.(string)
}

func (self *domAttr) removeData() {
	self.j.SetData(fmt.Sprintf(CONCORDE_DATA, self.id), nil)
}

func (self *domAttr) setData(v string) {
	self.j.SetData(fmt.Sprintf(CONCORDE_DATA, self.id), v)
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

func (self *domAttr) Id() string {
	return self.id
}

//NewStyleAttribute returns a dom attribute connected to the css
//style value named by n and selected by the selector j.  This results
//in a call to SetCSS and uses fmt.Sprintf() to format it's value,
//so the constraint result may be any type.  This the lower level
//interface, most users will probably prefer to use the
//StyleAttrBuilder interface.
func NewStyleAttr(n string, j jquery.JQuery) DomAttribute {
	result := &styleAttr{name: n}
	result.domAttr = newDomAttr(j, fmt.Sprintf("style:%s", n), result.get, result.set)
	return result
}

func (self *styleAttr) get() Equaler {
	if self.j.Css(self.name) == "undefined" {
		return StringEqualer{""}
	}
	return StringEqualer{self.j.Css(self.name)}
}

func (self *styleAttr) set(e Equaler) {
	self.j.SetCss(self.name, fmt.Sprintf("%s", e))
}

type textAttr struct {
	*domAttr
}

//NewTextAttr returns a dom attribute connected to the text property
//of the elements matched by j.   Most users will probably prefer to
//use the TextBuilder API, this is the lower level access to the raw
//DomAttribute.  This attribute uses fmt.Sprintf() to compute the final
//text value written, so the constraint result can be any type.
func NewTextAttr(j jquery.JQuery) DomAttribute {
	result := &textAttr{}
	result.domAttr = newDomAttr(j, "text", result.get, result.set)
	return result
}

func (self *textAttr) get() Equaler {
	if self.j.Text() == "undefined" {
		return StringEqualer{""}
	}
	return StringEqualer{self.j.Text()}
}

func (self *textAttr) set(e Equaler) {
	self.j.SetText(fmt.Sprintf("%v", e))
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
func NewHtmlAttrAttr(j jquery.JQuery, a htmlAttrName) DomAttribute {
	result := &htmlAttrAttr{name: a}
	result.domAttr = newDomAttr(j, "attr:"+string(a), result.get, result.set)
	return result
}

func (self *htmlAttrAttr) get() Equaler {
	if self.j.Attr(string(self.name)) == "undefined" {
		return StringEqualer{""}
	}
	return StringEqualer{self.j.Attr(string(self.name))}
}

func (self *htmlAttrAttr) set(e Equaler) {
	/*	b, ok := e.(BoolEqualer)
		if ok {
			if b.B {
				self.j.SetProp(string(self.name), true)
			} else {
				self.j.SetProp(string(self.name), false)
			}
			return
		}
	*/
	self.j.SetAttr(string(self.name), fmt.Sprintf("%v", e))
}

type propAttr struct {
	*domAttr
	name propName
}

//NewPropAttr provides an interface to the dom attribute for the property
//named, for the elements that are matched by j.  This is primarily
//useful for things that have true/false state (such checked, selected,
//or disabled state) so it expects a boolean attribute.
func NewPropAttr(j jquery.JQuery, n propName) DomAttribute {
	result := &propAttr{name: n}
	result.domAttr = newDomAttr(j, "prop:"+string(n), result.get, result.set)
	return result
}

func (self *propAttr) get() Equaler {
	b := self.j.Prop(string(self.name))
	return BoolEqualer{b.(bool)}
}

func (self *propAttr) set(e Equaler) {
	self.j.SetProp(string(self.name), e.(BoolEqualer).B)
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
	VALUE       = newAttrName("value")

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
func NewCssExistenceAttr(j jquery.JQuery, clazz CssClass) DomAttribute {
	result := &cssExistenceAttr{clazz: clazz}
	result.domAttr = newDomAttr(j, "cssclass:"+clazz.ClassName(), result.get, result.set)
	return result
}

func (self *cssExistenceAttr) get() Equaler {
	return BoolEqualer{self.j.HasClass(self.clazz.ClassName())}
}

func (self *cssExistenceAttr) set(e Equaler) {
	if e.(BoolEqualer).B {
		self.j.AddClass(self.clazz.ClassName())
	} else {
		self.j.RemoveClass(self.clazz.ClassName())
	}
}

//NewDisplayAttr returns a dom element that is connected to the css
//"display" attribute.  This is a special case of NewStyleAttribute
//that understands that a boolean can be used to display (true) or
//hide a given dom element.  This is the lower level interface and
//most users will prefer the DisplayAttrBuilder or DisplayAttribute
//calls.
func NewDisplayAttr(j jquery.JQuery) DomAttribute {
	result := &styleAttr{name: "display"}
	result.domAttr = newDomAttr(j, "style:display", result.getDisplay,
		result.setDisplay)
	return result
}

func (self *styleAttr) getDisplay() Equaler {
	if self.j.Css("display") == "undefined" {
		return BoolEqualer{true}
	}
	return BoolEqualer{self.j.Css("display") != "none"}
}

func (self *styleAttr) setDisplay(e Equaler) {
	if e.(BoolEqualer).B {
		self.j.SetCss("display", "")
	} else {
		self.j.SetCss("display", "none")
	}
}
