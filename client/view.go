package client

import (
	"fmt"
	//"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	//"honnef.co/go/js/console"
	//"reflect"
	"strings"
)

type ViewImpl struct {
	tag      string
	id       string
	classes  []string
	style    string
	children []*ViewImpl
	text     string
	builders []builder
	event    []*eventHandler
	ndom     NarrowDom
}

type option func(*ViewImpl) *ViewImpl

func Event(name EventName, fn EventFunc) option {
	return func(self *ViewImpl) *ViewImpl {
		self.event = append(self.event, &eventHandler{
			name: name,
			fn:   fn,
		})
		return self
	}
}

func Style(s string) option {
	return func(self *ViewImpl) *ViewImpl {
		self.style = self.style + s
		return self
	}
}

func Class(cl CssClass) option {
	return func(self *ViewImpl) *ViewImpl {
		self.classes = append(self.classes, cl.ClassName())
		return self
	}
}

func Text(str string) option {
	return func(self *ViewImpl) *ViewImpl {
		self.text = str
		return self
	}
}

func IdConstant(id string) option {
	return func(self *ViewImpl) *ViewImpl {
		if id == "" {
			panic("should not be calling IdConstant() with an empty id")
		}
		self.id = id
		return self
	}
}

func Id(id HtmlId) option {
	return func(self *ViewImpl) *ViewImpl {
		if id == nil {
			panic("should not be calling Id() on a nil HtmlId")
		}
		self.id = id.Id()
		return self
	}
}

func ModelId(m ModelName) option {
	return func(self *ViewImpl) *ViewImpl {
		id := fmt.Sprintf("%s-%s", strings.ToLower(self.tag), m.Id())
		self.id = newHtmlIdNoCheck(self.tag, id).Id()
		return self
	}
}

func addBuilder(b builder) option {
	return func(self *ViewImpl) *ViewImpl {
		self.builders = append(self.builders, b)
		return self
	}
}

func CssExistence(c CssClass, b BooleanAttribute) option {
	return addBuilder(cssExistenceBuilder(c, b))
}

func PropEqual(n propName, b BooleanAttribute) option {
	return addBuilder(propBuilder(n, b, nil))
}

func HtmlAttrEqual(h htmlAttrName, attr Attribute) option {
	return addBuilder(htmlAttrBuilder(h, attr, nil))
}

func HtmlAttrConstant(h htmlAttrName, str string) option {
	attr := NewStringSimple(str)
	return addBuilder(htmlAttrBuilder(h, attr, nil))
}

func TextEqual(attr Attribute) option {
	return addBuilder(textAttrBuilder(attr, nil))
}

//BindEqual constrains the attribute provided to be the same as the value of
//the tag this is located in.  Typically, this is an INPUT tag.  Data flows
//_from_ the input text that the user types to the attribute, not the other
//way around. There is a strange, but useful, edge case
//in the initialization of the INPUT tag: ff the attr provided returns a string
//value, that value is used to initialize INPUT field.
func BindEqual(attr Attribute) option {
	return addBuilder(valueAttrBuilder(attr, nil))
}

//Bind constrains the attribute provided to be a function of the value of
//the tag this call is located in.  Typically, this is an INPUT tag.  Data flows
//_from_ the input text that the user types to the attribute via this constraint
//given, not the other way around.  There is a strange, but useful, edge case
//in the initialization of the INPUT tag: ff the attr provided returns a string
//value, that value is used to initialize INPUT field.
func Bind(attr Attribute, cons Constraint) option {
	return addBuilder(valueAttrBuilder(attr, cons))
}

//HtmlIdFromModel returns an HtmlId object from the given modelname and
//tagname.  Resulting id is unique to the modelname and tag, but not
//between tags with the same name.
func HtmlIdFromModel(tag string, m ModelName) HtmlId {
	id := fmt.Sprintf("%s-%s", strings.ToLower(tag), m.Id())
	return NewHtmlId(tag, id)
}

//ParseHtml returns a NarrowDom that points at the fragment
//of HTML provided in t.  No attempt is made to validate that
//the HTML is sensible, much less syntatically correct.
func ParseHtml(t string) NarrowDom {
	parsed := jquery.ParseHTML(t)
	var nDom NarrowDom
	if TestMode {
		nDom = newTestOps()
	} else {
		if len(parsed) > 1 {
			div := jquery.NewJQuery(jquery.ParseHTML("<div/>"))
			div.Append(parsed...)
			return wrap(div)
		} else {
			nDom = wrap(jquery.NewJQuery(parsed[0]))
		}
	}
	return nDom
}

// Build converts a tree of *ViewImpls to a NarrowDom tree.
func (p *ViewImpl) Build() NarrowDom {
	id := ""
	classes := ""
	styles := ""

	//short circuit all this building if the node is already built
	if p.ndom != nil {
		return p.ndom
	}

	if p.id != "" {
		id = fmt.Sprintf(" id='%s'", p.id)
	}

	if p.classes != nil {
		classes = fmt.Sprintf(" class='%s'", strings.Join(p.classes, " "))
	}
	if p.style != "" {
		styles = fmt.Sprintf(" style='%s'", p.style)
	}

	var t string
	if p.text == "" {
		t = fmt.Sprintf("<%s%s%s%s/>", p.tag, id, classes, styles)
	} else {
		t = fmt.Sprintf("<%s%s%s%s>%s</%s>", p.tag, id, classes, styles, p.text, p.tag)
	}
	nDom := ParseHtml(t)
	for _, child := range p.children {
		built := child.Build()
		nDom.Append(built)
	}

	if p.builders != nil {
		for _, b := range p.builders {
			if b == nil {
				panic("found a nil builder in tree construction!")
			}
			b.build(nDom)
		}
	}

	if p.event != nil {
		for _, h := range p.event {
			//we have the object now, assign to j
			h.t = nDom
			h.register()
		}
	}
	return nDom
}

//HtmlConstant is a wrapper for creating an entire subtree from HTML text.  It makes
//no attempt to guarantee that the provided text is safe or sensible or
//anything.  This takes a constant not an attribute because it's not clear that
//we can meaningfully do anything other than just add the HTML provided as text
//to the DOM.
func HtmlConstant(s string) *ViewImpl {
	return &ViewImpl{
		ndom: ParseHtml(s),
	}
}

func IMG(obj ...interface{}) *ViewImpl {
	return tag("img", obj...)
}

func FORM(obj ...interface{}) *ViewImpl {
	return tag("form", obj...)
}

func DIV(obj ...interface{}) *ViewImpl {
	return tag("div", obj...)
}

func INPUT(obj ...interface{}) *ViewImpl {
	return tag("input", obj...)
}

func TEXTAREA(obj ...interface{}) *ViewImpl {
	return tag("textarea", obj...)
}

func LABEL(obj ...interface{}) *ViewImpl {
	return tag("label", obj...)
}

func A(obj ...interface{}) *ViewImpl {
	return tag("a", obj...)
}

func SPAN(obj ...interface{}) *ViewImpl {
	return tag("span", obj...)
}

func STRONG(obj ...interface{}) *ViewImpl {
	return tag("strong", obj...)
}
func P(obj ...interface{}) *ViewImpl {
	return tag("p", obj...)
}

func EM(obj ...interface{}) *ViewImpl {
	return tag("em", obj...)
}

func H1(obj ...interface{}) *ViewImpl {
	return tag("h1", obj...)
}

func H2(obj ...interface{}) *ViewImpl {
	return tag("h2", obj...)
}

func H3(obj ...interface{}) *ViewImpl {
	return tag("h3", obj...)
}

func H4(obj ...interface{}) *ViewImpl {
	return tag("h4", obj...)
}

func H5(obj ...interface{}) *ViewImpl {
	return tag("h5", obj...)
}

func H6(obj ...interface{}) *ViewImpl {
	return tag("h6", obj...)
}

func HR(obj ...interface{}) *ViewImpl {
	return tag("hr", obj...)
}

func LI(obj ...interface{}) *ViewImpl {
	return tag("li", obj...)
}

func UL(obj ...interface{}) *ViewImpl {
	return tag("ul", obj...)
}
func OL(obj ...interface{}) *ViewImpl {
	return tag("ol", obj...)
}
func BUTTON(obj ...interface{}) *ViewImpl {
	return tag("button", obj...)
}
func PRE(obj ...interface{}) *ViewImpl {
	return tag("pre", obj...)
}
func OPTION(obj ...interface{}) *ViewImpl {
	return tag("option", obj...)
}
func TABLE(obj ...interface{}) *ViewImpl {
	return tag("table", obj...)
}
func TR(obj ...interface{}) *ViewImpl {
	return tag("tr", obj...)
}
func TD(obj ...interface{}) *ViewImpl {
	return tag("td", obj...)
}
func TH(obj ...interface{}) *ViewImpl {
	return tag("th", obj...)
}
func tag(tagName string, obj ...interface{}) *ViewImpl {
	p := &ViewImpl{tag: tagName}

	for i := 0; i < len(obj); i++ {
		if obj[i] == nil {
			panic("nil value in view construction")
		}
		opt, ok := obj[i].(option)
		if ok {
			opt(p)
			continue
		}
		v, ok := obj[i].(*ViewImpl)
		if v == nil && ok {
			continue
		}
		if ok {
			p.children = append(p.children, v)
			continue
		}
		varr, isArray := obj[i].([]*ViewImpl)
		if isArray {
			for _, v := range varr {
				if v != nil {
					p.children = append(p.children, v)
				}
			}
			continue
		}
		panic(fmt.Sprintf("unable to understand type of parameter: %v (%T %d) to %s", obj[i], obj[i], i, tagName))
	}
	return p
}
