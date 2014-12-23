package client

import (
	"github.com/gopherjs/jquery"
)

//NarrowDom is a narrow interface to accessing the DOM that can be
//simulated for test purposes.
type NarrowDom interface {
	SetData(string, string)
	Data(string) string
	RemoveData(string)
	Css(string) string
	SetCss(string, string)
	Text() string
	SetText(string)
	Attr(string) string
	SetAttr(string, string)
	Prop(string) bool
	SetProp(string, bool)
	On(EventName, EventFunc)
	Trigger(EventName)
	HasClass(string) bool
	AddClass(string)
	RemoveClass(string)
	Val() string
	SetVal(string)
	Clear()
	Append(...NarrowDom)
	Prepend(...NarrowDom)
	Before(NarrowDom)
	Remove()
}

type jqueryWrapper struct {
	jq jquery.JQuery
}

func wrap(j jquery.JQuery) NarrowDom {
	return jqueryWrapper{j}
}

type testOpsImpl struct {
	data     map[string]string
	css      map[string]string
	text     string
	val      string
	attr     map[string]string
	prop     map[string]bool
	radio    map[string]string
	classes  map[string]int
	event    map[EventName]EventFunc
	parent   *testOpsImpl //need the nil
	children []*testOpsImpl
}

func newTestOps() NarrowDom {
	result := testOpsImpl{
		data:    make(map[string]string),
		css:     make(map[string]string),
		attr:    make(map[string]string),
		prop:    make(map[string]bool),
		classes: make(map[string]int),
		event:   make(map[EventName]EventFunc),
	}
	return &result
}

func (self *testOpsImpl) SetData(k string, v string) {
	self.data[k] = v
}

func (self jqueryWrapper) SetData(k string, v string) {
	self.jq.SetData(k, v)
}

func (self *testOpsImpl) RemoveData(k string) {
	delete(self.data, k)
}

func (self jqueryWrapper) RemoveData(k string) {
	self.jq.RemoveData(k)
}

func (self *testOpsImpl) Data(k string) string {
	return self.data[k]
}

func (self jqueryWrapper) Data(k string) string {
	i := self.jq.Data(k)
	if i == nil {
		return ""
	}
	return i.(string)
}

func (self *testOpsImpl) Css(k string) string {
	return self.css[k]
}
func (self jqueryWrapper) Css(k string) string {
	return self.jq.Css(k)
}

func (self *testOpsImpl) SetCss(k string, v string) {
	self.css[k] = v
}
func (self jqueryWrapper) SetCss(k string, v string) {
	self.jq.SetCss(k, v)
}

func (self *testOpsImpl) Text() string {
	return self.text
}

func (self jqueryWrapper) Text() string {
	return self.jq.Text()
}

func (self *testOpsImpl) SetText(v string) {
	self.text = v
}

func (self jqueryWrapper) SetText(v string) {
	self.jq.SetText(v)
}

func (self *testOpsImpl) Attr(k string) string {
	return self.attr[k]
}
func (self jqueryWrapper) Attr(k string) string {
	return self.jq.Attr(k)
}

func (self *testOpsImpl) SetAttr(k string, v string) {
	self.attr[k] = v
}
func (self jqueryWrapper) SetAttr(k string, v string) {
	self.jq.SetAttr(k, v)
}

func (self *testOpsImpl) Prop(k string) bool {
	return self.prop[k]
}
func (self jqueryWrapper) Prop(k string) bool {
	return self.jq.Prop(k).(bool)
}

func (self *testOpsImpl) SetProp(k string, v bool) {
	self.prop[k] = v
}
func (self jqueryWrapper) SetProp(k string, v bool) {
	self.jq.SetProp(k, v)
}

func (self *testOpsImpl) On(name EventName, fn EventFunc) {
	self.event[name] = fn
}

func (self jqueryWrapper) On(n EventName, fn EventFunc) {
	handler := eventHandler{
		name: n,
		fn:   fn,
		t:    self,
	}
	self.jq.On(n.String(), handler.handle)
}

func (self *testOpsImpl) Trigger(name EventName) {
	fn, ok := self.event[name]
	if ok {
		handler := eventHandler{
			fn:   fn,
			name: name,
			t:    self,
		}
		handler.handle(jquery.Event{Type: name.String()})
	}
}

func (self jqueryWrapper) Trigger(n EventName) {
	self.jq.Trigger(n.String())
}

func (self *testOpsImpl) HasClass(k string) bool {
	_, ok := self.classes[k]
	return ok
}

func (self jqueryWrapper) HasClass(k string) bool {
	return self.jq.HasClass(k)
}

func (self *testOpsImpl) Val() string {
	return self.val
}

func (self jqueryWrapper) Val() string {
	v := self.jq.Val()
	return v
}

func (self *testOpsImpl) SetVal(s string) {
	self.val = s
}

func (self jqueryWrapper) SetVal(s string) {
	self.jq.SetVal(s)
	self.jq.Underlying().Call("trigger", "input")
}

func (self *testOpsImpl) AddClass(s string) {
	self.classes[s] = 0
}

func (self jqueryWrapper) AddClass(s string) {
	self.jq.AddClass(s)
}

func (self *testOpsImpl) RemoveClass(s string) {
	delete(self.classes, s)
}

func (self jqueryWrapper) RemoveClass(s string) {
	self.jq.RemoveClass(s)
}
func (self *testOpsImpl) Clear() {
	self.children = make([]*testOpsImpl, 10)
}
func (self jqueryWrapper) Clear() {
	self.jq.Empty()
}
func (self *testOpsImpl) Remove() {
	siblings := self.parent.children
	for i, c := range siblings {
		if c == self {
			if i == len(siblings)-1 {
				self.parent.children = siblings[:i]
			} else {
				self.parent.children = append(siblings[:i], siblings[i+1:]...)
			}
		}
	}
}
func (self jqueryWrapper) Remove() {
	self.jq.Remove()
}

func (self *testOpsImpl) Append(childrennd ...NarrowDom) {
	for _, nd := range childrennd {
		child := nd.(*testOpsImpl)
		if child.parent != nil {
			panic("can't add a child, it already has a parent!")
		}
		child.parent = self
		self.children = append(self.children, child)
	}
}

func (self jqueryWrapper) Append(childrennd ...NarrowDom) {
	for _, nd := range childrennd {
		wrapper := nd.(jqueryWrapper)
		self.jq.Append(wrapper.jq)
	}
}

func (self *testOpsImpl) Prepend(childrennd ...NarrowDom) {
	for _, nd := range childrennd {
		child := nd.(*testOpsImpl)
		if child.parent != nil {
			panic("can't add a child, it already has a parent!")
		}
		child.parent = self
		self.children = append([]*testOpsImpl{child}, self.children...)
	}
}

//we have to walk the children provided in reverse order so their order is
//preserved on repeated calls to prepend
func (self jqueryWrapper) Prepend(childrennd ...NarrowDom) {
	for i := len(childrennd) - 1; i >= 0; i-- {
		nd := childrennd[i]
		wrapper := nd.(jqueryWrapper)
		self.jq.Prepend(wrapper.jq)
	}
}
func (self *testOpsImpl) Before(nd NarrowDom) {
	child := nd.(*testOpsImpl)
	parent := child.parent
	done := false
	for i, cand := range parent.children {
		if cand == child {
			rest := parent.children[i:]
			parent.children = append(parent.children[0:i], child)
			parent.children = append(parent.children, rest...)
			done = true
			break
		}
	}
	if !done {
		panic("unable to find child to insert before!")
	}
}

func (self jqueryWrapper) Before(nd NarrowDom) {
	wrapper := nd.(jqueryWrapper)
	self.jq.Before(wrapper.jq)
}

func (self *testOpsImpl) SetRadioButton(groupName string, value string) {
	self.radio[groupName] = value
}

func (self *testOpsImpl) RadioButton(groupName string) string {
	return self.radio[groupName]
}

func (self jqueryWrapper) SetRadioButton(groupName string, value string) {
	selector := "input[name=\"" + groupName + "\"][type=\"radio\"]"
	print("selector is ", selector)
	jq := jquery.NewJQuery(selector)
	jq.SetVal(value)
}

func (self jqueryWrapper) RadioButton(groupName string) string {
	selector := "input[name=\"" + groupName + "\"][type=\"radio\"]" + ":checked"
	print("selector is ", selector)
	jq := jquery.NewJQuery(selector)
	return jq.Val()
}
