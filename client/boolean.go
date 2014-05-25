package client

import (
	"fmt"
	//"honnef.co/go/js/console"
)

//Boolean attributes are Attributes that have boolean values.
type BooleanAttribute interface {
	Attribute
	Value() bool
	Set(b bool)
}

//BooleanSimple is a BooleanAttribute that holds a simple value. Attempts
//to Attach() a constraint to this value will panic.
type BooleanSimple struct {
	*AttributeImpl
}

//BoolEqualer is a wrapper around bool types for use with
//attributes.  It has the Equal() method for comparison.
type BoolEqualer struct {
	B bool
}

//Equal compares two values, must be BoolEqualers, to see if they
//are the same.
func (self BoolEqualer) Equal(e Equaler) bool {
	return self.B == e.(BoolEqualer).B
}

//String returns this value as a string (via fmt.Sprintf)
func (self BoolEqualer) String() string {
	return fmt.Sprintf("%v", self.B)
}

func (self *BooleanSimple) Value() bool {
	return self.Demand().(BoolEqualer).B
}

func (self *BooleanSimple) Set(b bool) {
	self.SetEqualer(BoolEqualer{b})
}

//NewBooleanSimple creates a new BooleanAttribute with a simple
//boolean initial value.
func NewBooleanSimple(b bool) BooleanAttribute {
	result := &BooleanSimple{NewAttribute(NORMAL, nil, nil)}
	result.Set(b)
	return result
}

//BooleanInverter is an attribute that can invert the value
//of the value it depends on.
type BooleanInverter struct {
	dep Attribute
}

func (self *BooleanInverter) Inputs() []Attribute {
	return []Attribute{self.dep}
}

func (self *BooleanInverter) Fn(in []Equaler) Equaler {
	if len(in) != 1 {
		panic("unexpected number of parameters to boolean inverter!")
	}
	return BoolEqualer{!in[0].(BoolEqualer).B}
}

//NewBooleanInverter returns constraint that inverts a boolean input.
func NewBooleanInverter(attr BooleanAttribute) Constraint {
	result := &BooleanInverter{attr}
	return result
}

//Booleq is a boolean equality constraint.
type BoolEq struct {
	dep BooleanAttribute
}

func (self *BoolEq) Inputs() []Attribute {
	return []Attribute{self.dep}
}

func (self *BoolEq) Fn(in []Equaler) Equaler {
	return BoolEqualer{self.dep.Value()}
}
