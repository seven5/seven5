package client

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
