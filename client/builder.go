package client

import ()

type builder interface {
	build(NarrowDom)
}

type getDomAttrFunc func(NarrowDom) DomAttribute

type builderBase struct {
	cons Constraint
	get  getDomAttrFunc
}

type anyBuilder struct {
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
		builderBase: builderBase{cons, func(n NarrowDom) DomAttribute {
			return NewTextAttr(n)
		}}}
}

func htmlAttrBuilder(h htmlAttrName, attr Attribute, cons Constraint) builder {
	return &anyBuilder{
		a: attr,
		builderBase: builderBase{cons, func(n NarrowDom) DomAttribute {
			return NewHtmlAttrAttr(n, h)
		}}}
}

func propBuilder(p propName, attr BooleanAttribute, cons Constraint) builder {
	return &booleanBuilder{
		b: attr,
		builderBase: builderBase{cons, func(n NarrowDom) DomAttribute {
			return NewPropAttr(n, p)
		}}}
}

func cssExistenceBuilder(c CssClass, b BooleanAttribute) builder {
	return &booleanBuilder{
		b: b,
		builderBase: builderBase{nil, func(n NarrowDom) DomAttribute {
			return NewCssExistenceAttr(n, c)
		}}}
}

func displayBuilder(b BooleanAttribute) builder {
	return &booleanBuilder{
		b: b,
		builderBase: builderBase{nil, func(n NarrowDom) DomAttribute {
			return NewDisplayAttr(n)
		}}}
}
