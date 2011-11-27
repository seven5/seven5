package css

import (
	"bytes"
	"fmt"
)

//
// Note: There are a lot of things wrong with this file from a CSS point of view.
// Note: such as Borders and Fonts not being fully specified.  I was trying to get
// Note: notation of the DSL right first.
//

//for colors
type ColorValue struct {
	Rgb int
}
//for sizes... just pick one
type Size struct {
	Px      int
	Pt      int
	Percent int
	Em      int
}

//one of N choices need to a sequence of bools of which one will be true
type TextAlign struct {
	Left   bool
	Center bool
	Right  bool
}

type FontWeight struct {
	Bold   bool
	Normal bool
}

type Float struct {
	Left  bool
	Right bool
}

type FontSize struct {
	Em float32
	Pt float32
}

type FontFamily struct {
	Name string
}
//illustrate the use of constants exported from this package in a nice way
var SERIF = FontFamily{"serif"}
var SANS_SERIF = FontFamily{"sans-serif"}
var TIMES = FontFamily{"times"}
var COURIER = FontFamily{"courier"}
var ARIAL = FontFamily{"arial"}
var BOLD = FontWeight{Bold: true}
var NORMAL_WEIGHT = FontWeight{Normal: true}

//Font should only be used when you are sure that you want to set ALL the font properties.
//The printout should check that everything has been set.
type Font struct {
	Family FontFamily
	Weight FontWeight
	Size   FontSize
}

//Border should only be used when you are sure that you want to set ALL the border properties.
//The printout should check that everything has been set.
type Border struct {
	Width Size
	Style BorderStyle
	Color Color
}

type BorderWidth struct {
	Width Size
}
type BorderStyle struct {
	Style string
}

var DOTTED = BorderStyle{"dotted"}
var SOLID = BorderStyle{"solid"}
var DASHED = BorderStyle{"dashed"}
var NO_BORDER = BorderStyle{"none"}

type BorderColor ColorValue
type Color ColorValue
type Background ColorValue

type Padding struct {
	All       Size
	TopBottom Size
	LeftRight Size
}

//prevent errors by making it explicit... would need to cross-check that both topBottom
//and leftRight were set when using that notation... no way to set all 4 because if you
//do that you could just the individual ones like MarginLeft
type Margin struct {
	All       Size
	TopBottom Size
	LeftRight Size
}
type MarginLeft Size
type MarginRight Size
type MarginTop Size
type MarginBottom Size

//this is a hack to allow you to basically put anything on the style of a "TextBox"
type TextBox Style
type TH TextBox
type TD TextBox

type Style []interface{}

type DomId string
type CSSClass string

type IdStyle struct {
	Name  DomId
	Style Style
}

type ClassStyle struct {
	Class  CSSClass
	Style  Style
	Parent *ClassStyle
}

type CrossStyle struct {
	Class     CSSClass
	Style     Style
	CrossWith *ClassStyle
}

func Reg(s interface{}) {
	//really, three types of parameters are allowed
	// IdStyle, ClassStyle, CrossStyle but trying to make the init function
	// easy to machine generate... have to use reflection to figure it out

}

func CSS(stylesheetName string) {
	//we are done now, code can be generated
	//probably can be a combo of recursion and reflection...
}

//
// IMPLEMENTATION
//
func (self TextAlign) String() string {
	switch {
	case self.Left:
		return "text-align: left;\n"
	case self.Center:
		return "text-align: center;\n"
	case self.Right:
		return "text-align: right;\n"
	}
	panic("TextAlign: no option chosen")
}

func (self BorderColor) String() string {
	return fmt.Sprintf("\tborder-color: #%x;\n", self.Rgb)
}
func (self Color) String() string {
	return fmt.Sprintf("\tcolor: #%x;\n", self.Rgb)
}

func (self Size) Zero() bool {
	return self.Pt == 0 && self.Px == 0 && self.Percent == 0
}

func (self Size) String() string {
	if self.Zero() {
		panic("empty size!")
	}
	count := 0
	if self.Pt != 0 {
		count++
	}
	if self.Px != 0 {
		count++
	}
	if self.Percent != 0 {
		count++
	}
	if self.Em != 0 {
		count++
	}
	if count > 1 {
		panic("too many fields set in size!")
	}
	if self.Px != 0 {
		return fmt.Sprintf("%dpx", self.Px)
	}
	if self.Pt != 0 {
		return fmt.Sprintf("%dpt", self.Pt)
	}
	if self.Pt != 0 {
		return fmt.Sprintf("%dem", self.Em)
	}
	return fmt.Sprintf("%d%%", self.Pt)

}

func (self Padding) String() string {
	return fourSides("padding", self.All, self.TopBottom, self.LeftRight)
}

func (self Margin) String() string {
	return fourSides("margin", self.All, self.TopBottom, self.LeftRight)
}

func fourSides(name string, all Size, tb Size, lr Size) string {
	if all.Zero() && lr.Zero() && tb.Zero() {
		panic(fmt.Sprintf("no size given for %s!", name))
	}
	if !all.Zero() && (!lr.Zero() || !tb.Zero()) {
		panic(fmt.Sprintf("too many sizes given for %s!", name))
	}
	if !all.Zero() {
		return fmt.Sprintf("\t%s: %v;\n", name, all)
	}
	if !lr.Zero() || !tb.Zero() {
		panic(fmt.Sprintf("cannot specify just one of LeftRight and TopBottom for %s", name))
	}
	return fmt.Sprintf("\t%s: %v %v\n", name, tb, lr)
}

func (self Style) String() string {
	buff := new(bytes.Buffer)
	self.print(buff, 0)
	return buff.String()
}

func (self Style) print(buff *bytes.Buffer, curr int) {
	buff.WriteString(fmt.Sprintf("%v", (self)[curr]))
	if len(self)-1 > curr {
		self.print(buff, curr+1)
	}
}

func (self *ClassStyle) String() string {
	return fmt.Sprintf(".%s {\n%v}\n", self.Class, self.Style)
}
