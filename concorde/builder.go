package concorde

import (
	"github.com/gopherjs/jquery"
)

type builder interface {
	build(jquery.JQuery)
}

type getDomAttrFunc func(jquery.JQuery) DomAttribute

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

func (self *anyBuilder) build(j jquery.JQuery) {
	dest := self.builderBase.get(j)
	if self.builderBase.cons == nil {
		self.builderBase.cons = eqConstraint{self.a}
	}
	dest.Attach(self.builderBase.cons)
}

func (self *booleanBuilder) build(j jquery.JQuery) {
	dest := self.builderBase.get(j)
	if self.builderBase.cons == nil {
		self.builderBase.cons = &BoolEq{self.b}
	}
	dest.Attach(self.builderBase.cons)
}

func textAttrBuilder(attr Attribute, cons Constraint) builder {
	return &anyBuilder{
		a: attr,
		builderBase: builderBase{cons, func(j jquery.JQuery) DomAttribute {
			return NewTextAttr(j)
		}}}
}

func htmlAttrBuilder(h htmlAttrName, attr Attribute, cons Constraint) builder {
	return &anyBuilder{
		a: attr,
		builderBase: builderBase{cons, func(j jquery.JQuery) DomAttribute {
			return NewHtmlAttrAttr(j, h)
		}}}
}

func propBuilder(p propName, attr BooleanAttribute, cons Constraint) builder {
	return &booleanBuilder{
		b: attr,
		builderBase: builderBase{cons, func(j jquery.JQuery) DomAttribute {
			return NewPropAttr(j, p)
		}}}
}

func cssExistenceBuilder(c CssClass, b BooleanAttribute) builder {
	return &booleanBuilder{
		b: b,
		builderBase: builderBase{nil, func(j jquery.JQuery) DomAttribute {
			return NewCssExistenceAttr(j, c)
		}}}
}

func displayBuilder(b BooleanAttribute) builder {
	return &booleanBuilder{
		b: b,
		builderBase: builderBase{nil, func(j jquery.JQuery) DomAttribute {
			return NewDisplayAttr(j)
		}}}
}
