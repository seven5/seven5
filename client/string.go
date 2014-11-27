package client

import (
	"fmt"
)

//StringAttributes are attributes are Attributes that have string values.
type StringAttribute interface {
	Attribute
	Value() string
	Set(s string)
}

//StringSimple is an StringAttribute which cannot have any
//incoming edges, it is just a source of a simple string value.
type StringSimple struct {
	*AttributeImpl
}

//StringEqualer is a wrapper around string types for use with
//attributes.  It has the Equal() method for comparison.
type StringEqualer struct {
	S string
}

//Equal compares two values, must be StringEqualers, to see if they
//are the same.
func (self StringEqualer) Equal(e Equaler) bool {
	return self.S == e.(StringEqualer).S
}

//String returns this value as a string (via fmt.Sprintf)
func (self StringEqualer) String() string {
	return fmt.Sprintf("%s", self.S)
}

func (self *StringSimple) Value() string {
	return self.Demand().(StringEqualer).S
}

func (self *StringSimple) Set(s string) {
	self.SetEqualer(StringEqualer{s})
}

//NewStringSimple creates a new StringAttribute with a simple
//int initial value.
func NewStringSimple(s string) StringAttribute {
	result := &StringSimple{
		NewAttribute(NORMAL, nil, nil),
	}
	result.Set(s)
	return result
}

type eqConstraint struct {
	dep Attribute
}

func (self eqConstraint) Inputs() []Attribute {
	return []Attribute{self.dep}
}

func (self eqConstraint) Fn(in []Equaler) Equaler {
	if len(in) != 1 {
		panic("wrong number of inputs to equality constraint")
	}
	return self.dep.Demand()
}

func StringEquality(src Attribute) Constraint {
	return eqConstraint{src}
}

func Equality(dest, src Attribute) {
	dest.Attach(eqConstraint{src})
}
