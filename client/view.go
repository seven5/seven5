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
	children []*ViewImpl
	text     string
	builders []builder
	event    []*eventHandler
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

func Class(cl CssClass) option {
	return func(self *ViewImpl) *ViewImpl {
		self.classes = append(self.classes, cl.ClassName())
		return self
	}
}

func Id(id HtmlId) option {
	return func(self *ViewImpl) *ViewImpl {
		self.id = id.Id()
		return self
	}
}

func ModelId(m ModelName) option {
	return func(self *ViewImpl) *ViewImpl {
		self.id = HtmlIdFromModel(self.tag, m).Id()
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

func TextEqual(attr Attribute) option {
	return addBuilder(textAttrBuilder(attr, nil))
}

//HtmlIdFromModel returns an HtmlId object from the given modelname and
//tagname.  Resulting id is unique to the modelname and tag, but not
//between tags with the same name.
func HtmlIdFromModel(tag string, m ModelName) HtmlId {
	id := fmt.Sprintf("%s-%s", strings.ToLower(tag), m.Id())
	return NewHtmlId(tag, id)
}

func (p *ViewImpl) Build() jquery.JQuery {
	id := ""
	classes := ""
	if p.id != "" {
		id = fmt.Sprintf(" id='%s'", p.id)
	}

	if p.classes != nil {
		classes = fmt.Sprintf(" class='%s'", strings.Join(p.classes, " "))
	}
	var t string
	if p.text == "" {
		t = fmt.Sprintf("<%s%s%s/>", p.tag, id, classes)
	} else {
		t = fmt.Sprintf("<%s%s%s>%s</%s>", p.tag, id, classes, p.text, p.tag)
	}
	parsed := jquery.ParseHTML(t)
	j := jquery.NewJQuery(parsed[0])
	for _, child := range p.children {
		j.Append(child.Build())
	}

	if p.builders != nil {
		for _, b := range p.builders {
			b.build(j)
		}
	}

	if p.event != nil {
		for _, h := range p.event {
			//we have the object now, assign to j
			h.j = j
			h.register()
		}
	}

	return j
}

func IMG(obj ...interface{}) *ViewImpl {
	return tag("img", obj...)
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

func tag(tagName string, obj ...interface{}) *ViewImpl {
	p := &ViewImpl{tag: tagName}

	for i := 0; i < len(obj); i++ {
		if obj[i] == nil {
			panic("cant use a nil value in view construction")
		}
		opt, ok := obj[i].(option)
		if ok {
			opt(p)
			continue
		}
		v, ok := obj[i].(*ViewImpl)
		if ok {
			p.children = append(p.children, v)
			continue
		}
		panic(fmt.Sprintf("unable to understand type of parameter: %v (%T %d) to %s", obj[i], obj[i], i, tagName))
	}
	return p
}
