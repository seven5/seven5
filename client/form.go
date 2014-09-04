package client

import (
	"github.com/gopherjs/jquery"
	//"honnef.co/go/js/console"
	"fmt"
)

type FormElement interface {
	Dom() NarrowDom
	Selector() string
	ContentAttribute() Attribute
	Val() string
	SetVal(string)
}

//InputTextId is a special case of HtmlId that is the exported text
//of a type in field.
type InputTextId interface {
	FormElement
}

type inputTextIdImpl struct {
	htmlIdImpl
	attr *AttributeImpl
}

type RadioGroup interface {
	FormElement
}

type SelectGroup interface {
	FormElement
}

type radioGroupImpl struct {
	dom      NarrowDom
	selector string
	attr     *AttributeImpl
}

type selectGroupImpl struct {
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
	self.htmlIdImpl.Dom().Trigger(INPUT_EVENT)
}

//Val returns the value of an input field.
func (self inputTextIdImpl) Val() string {
	v := self.htmlIdImpl.t.Val()
	if v == "undefined" {
		return ""
	}
	return v
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
	selector := "input:radio[name=\"" + name + "\"]"
	result := radioGroupImpl{selector: selector}
	if TestMode {
		result.dom = newTestOps()
	} else {
		result.dom = wrap(jquery.NewJQuery(selector))
	}
	result.attr = NewAttribute(VALUE_ONLY, result.value, nil)
	result.dom.On(CLICK, func(jquery.Event) {
		result.attr.markDirty()
	})
	return result
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

//Val returns the currently selected item or "" if no item is selected. It
//cannot deal with values that are "" or are exactly the string "undefined".
func (self radioGroupImpl) Val() string {
	var v string
	switch d := self.dom.(type) {
	case jqueryWrapper:
		v = d.jq.Filter(":checked").Val()
	case *testOpsImpl:
		v = d.Val()
	default:
		panic("unknown dom type")
	}
	if v == "undefined" {
		return ""
	}
	return v
}

func (self radioGroupImpl) Dom() NarrowDom {
	return self.dom
}

//SetVal sets the current value of the radio buttons defined in
//in the radioGroup.
func (self radioGroupImpl) SetVal(s string) {
	switch d := self.dom.(type) {
	case jqueryWrapper:
		child := d.jq.Filter(fmt.Sprintf("[value=\"%s\"]", s))
		child.SetProp("checked", true)
	case *testOpsImpl:
		d.SetVal(s)
		d.Trigger(CLICK)
	default:
		panic("unknown type of dom pointer!")
	}
}

//NewSelectGroup selects one of a named set of elements, usually
//rendered as a drop down list.
func NewSelectGroupId(id string) SelectGroup {
	selector := "select#" + id
	result := selectGroupImpl{selector: selector}
	if TestMode {
		result.dom = newTestOps()
	} else {
		result.dom = wrap(jquery.NewJQuery(selector))
	}
	result.attr = NewAttribute(VALUE_ONLY, result.value, nil)
	result.dom.On(CLICK, func(jquery.Event) {
		result.attr.markDirty()
	})
	return result
}

func (self selectGroupImpl) value() Equaler {
	return StringEqualer{S: self.dom.Val()}
}

func (self selectGroupImpl) Selector() string {
	return self.selector
}

func (self selectGroupImpl) ContentAttribute() Attribute {
	return self.attr
}

func (self selectGroupImpl) Dom() NarrowDom {
	return self.dom
}

//Val returns the currently selected item or "" if no item is selected. It
//cannot deal with values that are "" or are exactly the string "undefined".
func (self selectGroupImpl) Val() string {
	var v string
	switch d := self.dom.(type) {
	case jqueryWrapper:
		v = d.jq.Find(":selected").Val()
	case *testOpsImpl:
		v = d.Val()
	default:
		panic("unknown dom type")
	}
	if v == "undefined" {
		return ""
	}
	return v
}

//SetVal sets the current value of the group defined in
//in the selectGroup.
func (self selectGroupImpl) SetVal(s string) {
	switch d := self.dom.(type) {
	case jqueryWrapper:
		child := d.jq.Filter(fmt.Sprintf("[value=\"%s\"]", s))
		child.SetProp("selected", true)
	case *testOpsImpl:
		d.SetVal(s)
		d.Trigger(CLICK)
	default:
		panic("unknown type of dom pointer!")
	}
}

//
// FORM VALIDATION CONSTRAINT
//
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
