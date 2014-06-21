package client

import (
	"github.com/gopherjs/jquery"
	//"honnef.co/go/js/console"
)

type FormElement interface {
	Selector() string
	ContentAttribute() Attribute
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
	Val() string
	Event(EventName, EventFunc)
}

type radioGroupImpl struct {
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
	result.htmlIdImpl.Event(INPUT_EVENT, func(jquery.JQuery, jquery.Event) {
		result.attr.markDirty()
	})
	return result
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
	result := radioGroupImpl{
		selector: "input[name=\"" + name + "\"][type=\"radio\"]",
	}
	result.attr = NewAttribute(VALUE_ONLY, result.value, nil)
	result.Event(CLICK, func(jquery.JQuery, jquery.Event) {
		result.attr.markDirty()
	})
	return result
}

func (self radioGroupImpl) value() Equaler {
	return StringEqualer{S: self.Val()}
}

func (self radioGroupImpl) Event(name EventName, fn EventFunc) {
	h := &eventHandler{name, jquery.NewJQuery(self.selector), fn}
	h.register()
}

func (self radioGroupImpl) Val() string {
	jq := jquery.NewJQuery(self.selector + ":checked")
	val := jq.Val()
	if val == "undefined" {
		return ""
	}
	return val
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
