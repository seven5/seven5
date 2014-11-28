package client

import ()

type builder interface {
	build(NarrowDom)
}

type getDomAttrFunc func(NarrowDom) Attribute

type builderBase struct {
	cons Constraint
	get  getDomAttrFunc
}

type anyBuilder struct {
	builderBase
	a Attribute
}

type reverseBuilder struct {
	builderBase
	a Attribute
}

type booleanBuilder struct {
	builderBase
	b BooleanAttribute
}

func (self *anyBuilder) build(n NarrowDom) {
	dest := self.builderBase.get(n)
	if self.builderBase.cons == nil {
		self.builderBase.cons = eqConstraint{self.a}
	}
	dest.Attach(self.builderBase.cons)
}

func (self *reverseBuilder) build(n NarrowDom) {
	dest := self.a
	if self.builderBase.cons == nil {
		self.builderBase.cons = eqConstraint{self.builderBase.get(n)}
	}
	dest.Attach(self.builderBase.cons)
}

func (self *booleanBuilder) build(n NarrowDom) {
	dest := self.builderBase.get(n)
	if self.builderBase.cons == nil {
		self.builderBase.cons = &BoolEq{self.b}
	}
	dest.Attach(self.builderBase.cons)
}

func textAttrBuilder(attr Attribute, cons Constraint) builder {
	return &anyBuilder{
		a: attr,
		builderBase: builderBase{cons, func(n NarrowDom) Attribute {
			return NewTextAttr(n)
		}}}
}

func valueAttrBuilder(attr Attribute, cons Constraint) builder {
	return &reverseBuilder{
		a: attr,
		builderBase: builderBase{cons, func(n NarrowDom) Attribute {
			raw := attr.Demand()
			if s, ok := raw.(StringEqualer); ok {
				n.SetVal(s.S)
			}
			return NewValueAttr(n)
		}}}
}

func htmlAttrBuilder(h htmlAttrName, attr Attribute, cons Constraint) builder {
	return &anyBuilder{
		a: attr,
		builderBase: builderBase{cons, func(n NarrowDom) Attribute {
			return NewHtmlAttrAttr(n, h)
		}}}
}

func propBuilder(p propName, attr BooleanAttribute, cons Constraint) builder {
	return &booleanBuilder{
		b: attr,
		builderBase: builderBase{cons, func(n NarrowDom) Attribute {
			return NewPropAttr(n, p)
		}}}
}

func cssExistenceBuilder(c CssClass, b BooleanAttribute) builder {
	return &booleanBuilder{
		b: b,
		builderBase: builderBase{nil, func(n NarrowDom) Attribute {
			return NewCssExistenceAttr(n, c)
		}}}
}

func displayBuilder(b BooleanAttribute) builder {
	return &booleanBuilder{
		b: b,
		builderBase: builderBase{nil, func(n NarrowDom) Attribute {
			return NewDisplayAttr(n)
		}}}
}
