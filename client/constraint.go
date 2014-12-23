package client

import (
	"fmt"
)

var (
	TestMode = false
)

//Equaler is an interface that forces objects to be comparable
//for equality.  The constraint algorithm needs this to know when
//a value changes.
type Equaler interface {
	Equal(Equaler) bool
}

//Attribute represents a value that can be manipulated by a constraint
//function.  Note that this object has no public set or get methods because
//typically consumers will use a typed version of this such as
//BooleanAttribute or IntegerAttribute.   An attribute can
//have a constraint attached to indicate that the value of the
//Attribute is computed from other attributes.  Note that setting
//an attribute which has a Constraint attached will result in
//a panic.
type Attribute interface {
	Attach(Constraint)
	Detach()
	Demand() Equaler
	SetEqualer(Equaler)
	SetDebugName(n string)
}

//Constraint is a function of Attributes that results in an attribute.
//Note that the implementation of a constraint.  The implementation
//of the Fn() will be passed the current values of the Attributes
//returned by Inputs().  The implementation of Fn() must be careful
//to not use other values that affect the result since these are
//not known to the system and unexpected/wrong results will occur.
//Most users of this interface will find it easier to use some
//specialized version of the interface that offers typed values to
//the computing function.
type Constraint interface {
	Inputs() []Attribute
	Fn([]Equaler) Equaler
}

type ConstraintFunc func([]Equaler) Equaler

type simpleConstraint struct {
	attr []Attribute
	fn   ConstraintFunc
}

//NewSimpleConstraint returns a new constraint that can derive its value
//from a fixed set of inputs.  The function fn is called to compute the vaule
//of the constraints and the parameters to fn are the vaules (in order)
//of the attributes at the current time.
func NewSimpleConstraint(fn ConstraintFunc, attr ...Attribute) Constraint {
	return simpleConstraint{attr, fn}
}

func (self simpleConstraint) Inputs() []Attribute {
	return self.attr
}

func (self simpleConstraint) Fn(v []Equaler) Equaler {
	return self.fn(v)
}

// SelectableAttribute provides an abstraction which allows an attribute
// to be turned on an off by another boolean attribute, called the chooser.
// When the chooser is false, this attribute's value is a
// constant (typically provided at construction time).  If the chooser is true
// then this attribute is a shim over another attribute, also provided at
// construction-time.  Note that assignemnts to this attribute are logically
// passed through to the attribute, but are ignored if the chooser is false.
// Calls to SetEqualer will be honored, and passed through, if the chooser is true.
type SelectableAttribute struct {
	*AttributeImpl
	chooser         BooleanAttribute
	unselectedValue Equaler
	attr            Attribute
	cons            Constraint
}

//NewSelectableAttribute creates a new SelectableAttribute with a given chooser
//a boolean attribute, a constant to be used when the chooser is false, and
//attribute to used when the chooser is true.  If you pass nil as the chooser
//this call will create a new boolean attribute to use as the chooser and
//it's initial value will be false.  The attribute provided must not be nil.
func NewSelectableAttribute(chooser BooleanAttribute, unselectedValue Equaler,
	attr Attribute) *SelectableAttribute {

	cattr := chooser
	if cattr == nil {
		cattr = NewBooleanSimple(false)
	}
	result := &SelectableAttribute{
		chooser:         cattr,
		unselectedValue: unselectedValue,
		attr:            attr,
	}
	result.AttributeImpl = NewAttribute(NORMAL, nil, nil)
	result.cons = NewSimpleConstraint(result.value, cattr, attr)
	result.AttributeImpl.Attach(result.cons)
	return result
}

//Chooser return  the chooser that is in use for this
//selectable attribute.  This is useful when you want feedabck about the chooser
//itself, rather than the composite output provided by SelectableAttribute.
func (s *SelectableAttribute) Chooser() BooleanAttribute {
	return s.chooser
}

//Selectable returns the attribute that is turned on and off inside this
//atribute.
func (s *SelectableAttribute) Selectable() Attribute {
	return s.attr
}

//value is the function that takes in the values of the two dependent attributes
//and figures out the value of this attribute, embodied in wrapped.
func (s *SelectableAttribute) value(v []Equaler) Equaler {
	chooserVal := v[0].(BoolEqualer).B
	attrVal := v[1]

	if chooserVal {
		return attrVal
	}
	return s.unselectedValue
}

//Attach attaches a constraint to this selectable attribute.  Note that this
//constraint will be called on any changes to the SelectableAttribute's boolean
//chooser or its underlying attribute.
func (s *SelectableAttribute) Attach(cons Constraint) {
	s.AttributeImpl.Attach(cons)
}

//Detach removes a constraint attached to this SelectableAttribute.
func (s *SelectableAttribute) Detach() {
	s.AttributeImpl.Detach()
}

//SetEqualer assigns the value provided to the underlying attribute of this
//type, only if the chooser of this SelectableAttribute is true.  If the chooser
//is false, this call is ignored.
func (s *SelectableAttribute) SetEqualer(e Equaler) {
	if s.chooser.Demand().(BoolEqualer).B {
		s.attr.SetEqualer(e)
	}
}

//SetDebugName is useful when examining debug messages from seven5 itself.
func (s *SelectableAttribute) SetDebugName(n string) {
	s.AttributeImpl.SetDebugName(n)
	s.chooser.SetDebugName(fmt.Sprintf("%s-chooser", n))
	s.attr.SetDebugName(fmt.Sprintf("%s-attr", n))
}

//SetSelected sets the selected state of this attribute to the value provided.
func (s *SelectableAttribute) SetSelected(b bool) {
	s.chooser.SetEqualer(BoolEqualer{B: b})
}

//Selected returns the current selected state.
func (s *SelectableAttribute) Selected() bool {
	return s.chooser.Demand().(BoolEqualer).B
}

//Demand returns the value of the attribute that is selectable if the chooser
//is true, otherwise the constant value provided at construction time.
func (s *SelectableAttribute) Demand() Equaler {
	return s.AttributeImpl.Demand()
}
