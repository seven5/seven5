package client

import (
	"fmt"
)

//FloatAttributes are attributes are Attributes that have float64 values.
type FloatAttribute interface {
	Attribute
	Value() float64
	Set(f float64)
}

//FloatSimple is an FloatAttribute which cannot have any
//incoming edges, it is just a source of a simple float64 value.
type FloatSimple struct {
	*AttributeImpl
}

//FloatEqualer is a wrapper around float64 types for use with
//attributes.  It has the Equal() method for comparison.
type FloatEqualer struct {
	F float64
}

//Equal compares two values, must be FloatEqualers, to see if they
//are the same.
func (self FloatEqualer) Equal(e Equaler) bool {
	return self.F == e.(FloatEqualer).F
}

//Float returns this Float as a string (via fmt.Sprintf)
func (self FloatEqualer) String() string {
	return fmt.Sprintf("%f", self.F)
}

func (self *FloatSimple) Value() float64 {
	return self.Demand().(FloatEqualer).F
}

func (self *FloatSimple) Set(f float64) {
	self.SetEqualer(FloatEqualer{f})
}

//NewFloatSimple creates a new FloatAttribute with a simple
//float64 initial value.
func NewFloatSimple(f float64) FloatAttribute {
	result := &FloatSimple{
		NewAttribute(NORMAL, nil, nil),
	}
	result.Set(f)
	return result
}

//FloatEq is an float equality constraint.
type FloatEq struct {
	dep FloatAttribute
}

func (self *FloatEq) Inputs() []Attribute {
	return []Attribute{self.dep}
}

func (self *FloatEq) Fn(in []Equaler) Equaler {
	return FloatEqualer{self.dep.Value()}
}
