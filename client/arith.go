package client

//ArithmeticConstraint allows simple arithmetic operations on integers.
type ArithmeticConstraint interface {
	Constraint
}

//Arithmetic op represents one of the binary operators of arithmetic
//plus the mod (%) operator.  It also has a special case, FUNC_OP
//meaning the user will supply a function to compute the operation.
type ArithmeticOp int

const (
	ADD_OP = iota
	SUB_OP
	MULT_OP
	DIV_OP
	MOD_OP
	FUNC_OP
)

func (self ArithmeticOp) String() string {
	switch self {
	case ADD_OP:
		return "+"
	case SUB_OP:
		return "-"
	case MULT_OP:
		return "*"
	case DIV_OP:
		return "/"
	case MOD_OP:
		return "%"
	case FUNC_OP:
		return "func"
	}
	panic("unknown ArithmeticOp")
}

type arithImpl struct {
	op ArithmeticOp
	in []Attribute
	fn func([]int) int
}

func (self *arithImpl) Inputs() []Attribute {
	return self.in
}

func (self *arithImpl) Fn(in []Equaler) Equaler {
	result := in[0].(IntEqualer).I
	if self.op != FUNC_OP {
		for _, raw := range in[1:] {
			arg := raw.(IntEqualer).I
			switch self.op {
			case ADD_OP:
				result += arg
			case SUB_OP:
				result -= arg
			case MULT_OP:
				result *= arg
			case DIV_OP:
				result /= arg
			case MOD_OP:
				result %= arg
			default:
				panic("unknown integer operation type")
			}
		}
		return IntEqualer{result}
	}
	//func op case
	params := make([]int, len(in))
	for index, i := range in {
		params[index] = i.(IntEqualer).I
	}
	return IntEqualer{self.fn(params)}
}

//NewArithmeticConstraint creates a constraint object with the default
//operation given.  If the fn is set to nil, the arithmetic operation
//provided is applied to all the arguments.  For subtraction, modulus, and division,
//this may not be what you want since this operations are more commonly
//binary (two parameters).  If you have a complex function of
//integers, it is useful to call this method directly but most users
//will want to use AdditionConstraint(), SubtractionConstraint() or similar since
//these have types and semantics "baked in".
func NewArithmeticConstraint(op ArithmeticOp, attr []Attribute, fn func([]int) int) ArithmeticConstraint {
	if len(attr) == 0 {
		panic("doesn't make any sense to create a constraint on no attributes")
	}
	if fn != nil && op != FUNC_OP {
		panic("if you are supplying the func, you should not pass an operation type")
	}
	return &arithImpl{
		op: op,
		in: attr,
		fn: fn,
	}
}

//AdditionConstraint creates a constraint that adds two values. The only reason
//to pass the function argument is to create a result that differs from the
//sum by a constant.  Note that the function, if provided, must return the
//"sum" as well as adding any constants.
func AdditionConstraint(a1 IntegerAttribute, a2 IntegerAttribute, fn func(int, int) int) ArithmeticConstraint {
	if fn != nil {
		return NewArithmeticConstraint(FUNC_OP, []Attribute{a1, a2},
			func(x []int) int {
				return fn(x[0], x[1])
			})
	} else {
		return NewArithmeticConstraint(ADD_OP, []Attribute{a1, a2}, nil)
	}
}

//SubtractionConstraint creates a constraint that subtractions two values. The only reason
//to pass the function argument is to create a result that differs from the
//sum by a constant.  Note that the function, if provided, must return the
//"difference" as well as subtracting anyok, constants.
func SubtractionConstraint(a1 IntegerAttribute, a2 IntegerAttribute, fn func(int, int) int) ArithmeticConstraint {
	if fn != nil {
		return NewArithmeticConstraint(FUNC_OP, []Attribute{a1, a2},
			func(x []int) int {
				return fn(x[0], x[1])
			})
	} else {
		return NewArithmeticConstraint(SUB_OP, []Attribute{a1, a2}, nil)
	}
}

//SumConstraint creates a constraint that add any number of attributes.
//The attribute values are copied so the caller can discard the slice
//after this use.
func SumConstraint(attr ...IntegerAttribute) ArithmeticConstraint {
	attribute := make([]Attribute, len(attr))
	for i, a := range attr {
		attribute[i] = Attribute(a)
	}
	return NewArithmeticConstraint(ADD_OP, attribute, nil)
}

//ProductConstraint creates a constraint that multiplies any number of attributes.
//The attribute vaules are copied so the slice can be discarded.
func ProductConstraint(attr ...IntegerAttribute) ArithmeticConstraint {
	attribute := make([]Attribute, len(attr))
	for i, a := range attr {
		attribute[i] = Attribute(a)
	}
	return NewArithmeticConstraint(MULT_OP, attribute, nil)
}
