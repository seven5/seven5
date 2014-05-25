package client

import (
	"fmt"
	//	"honnef.co/go/js/console"
)

//IntegerAttributes are attributes are Attributes that have int values.
type IntegerAttribute interface {
	Attribute
	Value() int
	Set(i int)
}

//IntegerSimple is an IntegerAttribute which cannot have any
//incoming edges, it is just a source of a simple int value.
type IntegerSimple struct {
	*AttributeImpl
}

//IntEqualer is a wrapper around int types for use with
//attributes.  It has the Equal() method for comparison.
type IntEqualer struct {
	I int
}

//Equal compares two values, must be IntEqualers, to see if they
//are the same.
func (self IntEqualer) Equal(e Equaler) bool {
	return self.I == e.(IntEqualer).I
}

//String returns this integer as a string (via fmt.Sprintf)
func (self IntEqualer) String() string {
	return fmt.Sprintf("%d", self.I)
}

func (self *IntegerSimple) Value() int {
	return self.Demand().(IntEqualer).I
}

func (self *IntegerSimple) Set(i int) {
	self.SetEqualer(IntEqualer{i})
}

//NewIntegerSimple creates a new IntegerAttribute with a simple
//int initial value.
func NewIntegerSimple(i int) IntegerAttribute {
	result := &IntegerSimple{
		NewAttribute(NORMAL, nil, nil),
	}
	result.Set(i)
	return result
}

//IntegerEq is an integer equality constraint.
type IntegerEq struct {
	dep IntegerAttribute
}

func (self *IntegerEq) Inputs() []Attribute {
	return []Attribute{self.dep}
}

func (self *IntegerEq) Fn(in []Equaler) Equaler {
	return IntEqualer{self.dep.Value()}
}
