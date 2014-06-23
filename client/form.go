package client

import (
	"github.com/gopherjs/jquery"
	//"honnef.co/go/js/console"
)

type FormElement interface {
	Selector() string
	ContentAttribute() Attribute
	Val() string
	SetVal(string)
}

//InputTextId is a special case of HtmlId that is the exported text
//of a type in field.
type InputTextId interface {
	HtmlId
	FormElement
}

type inputTextIdImpl struct {
	htmlIdImpl
	attr *AttributeImpl
}

type RadioGroup interface {
	FormElement
	Dom() NarrowDom
}

type radioGroupImpl struct {
	dom      NarrowDom
	selector string
	attr     *AttributeImpl
}

//NewInputTextId returns a reference to a static part of the DOM that is the
//an input tag with the given id.
func NewInputTextId(id string) InputTextId {
	result := inputTextIdImpl{
		htmlIdImpl: NewHtmlId("input", id).(htmlIdImpl),
	}
	result.attr = NewAttribute(VALUE_ONLY, result.value, nil)
	result.htmlIdImpl.Dom().On(INPUT_EVENT, func(jquery.Event) {
		result.attr.markDirty()
	})
	return result
}

//SetText puts the text provided into the value (such as with an input field).
func (self inputTextIdImpl) SetVal(s string) {
	self.htmlIdImpl.t.SetVal(s)
}

//Val returns the value of an input field.
func (self inputTextIdImpl) Val() string {
	return self.htmlIdImpl.t.Val()
}

func (self inputTextIdImpl) value() Equaler {
	return StringEqualer{S: self.Val()}
}

func (self inputTextIdImpl) ContentAttribute() Attribute {
	return self.attr
}

//NewRadioGroup selects a named set of radio button elements
//with a given name.
func NewRadioGroup(name string) RadioGroup {
	selector := "input[name=\"" + name + "\"][type=\"radio\"]"
	result := radioGroupImpl{selector: selector}
	if TestMode {
		result.dom = newTestOps()
	} else {
		wrap(jquery.NewJQuery(selector))
	}
	result.attr = NewAttribute(VALUE_ONLY, result.value, nil)
	result.dom.On(CLICK, func(jquery.Event) {
		result.attr.markDirty()
	})
	return result
}

func (self radioGroupImpl) Dom() NarrowDom {
	return jqueryWrapper{jq: jquery.NewJQuery()}
}

func (self radioGroupImpl) value() Equaler {
	return StringEqualer{S: self.dom.Val()}
}

func (self radioGroupImpl) Selector() string {
	return self.selector
}

func (self radioGroupImpl) ContentAttribute() Attribute {
	return self.attr
}

type formValidFunc func(map[string]Equaler) Equaler

type formValidConstraint struct {
	id []FormElement
	fn formValidFunc
}

func (self formValidConstraint) Inputs() []Attribute {
	result := make([]Attribute, len(self.id))
	for i, id := range self.id {
		result[i] = id.ContentAttribute()
	}
	return result
}

func (self formValidConstraint) Fn(v []Equaler) Equaler {
	args := make(map[string]Equaler)
	for i, id := range self.id {
		args[id.Selector()] = v[i]
	}
	return self.fn(args)
}

//NewFormValidConstraint creates a constraint that uses the function
//supplied to compute a value and has dependencies on all the remaining
//parameters.
func NewFormValidConstraint(fn formValidFunc, id ...FormElement) formValidConstraint {
	return formValidConstraint{fn: fn, id: id}
}
