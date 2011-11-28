package css

import (
	//	"bytes"
	"fmt"
	"strings"
)

//
// SELECTORS
//
type Id string
type Class string
type Elem string
type Composite string

type Stor interface {
	storTag()
}

func (self Class) String() string {
	return "." + string(self)
}
func (self Class) storTag() {
}

func (self Id) String() string {
	return "#" + string(self)
}
func (self Id) storTag() {
}

func (self Elem) String() string {
	return strings.ToUpper(string(self))
}
func (self Elem) storTag() {
}

type ParentChildSelector struct {
	Parent, Child Stor
}

func (self *ParentChildSelector) String() string {
	return fmt.Sprintf("%v > %v", self.Parent, self.Child)
}

func (self ParentChildSelector) storTag() {
}

type AllElementSelector struct {

}

func (self AllElementSelector) String() string {
	return "*"
}
func (self AllElementSelector) storTag() {

}

//composites are computed by some other function
func (self Composite) String() string {
	return string(self)
}

func (self Composite) storTag() {
}

func Hover(e Elem) Stor {
	return Composite(fmt.Sprintf("%v:hover", e))
}
func Selector(s ...Stor) Stor {
	result := ""
	for _, x := range s {
		result += " " + fmt.Sprintf("%v", x)
	}
	return Composite(strings.TrimLeft(result, " "))
}

func Inherit(c ...Class) Stor {
	result := ""
	for _, x := range c {
		result += fmt.Sprintf("%v", x)
	}
	return Composite(result)
}

//
// SELECTOR CONSTANTS
//

var TH = Elem("TH")
var TD = Elem("TD")
var BODY = Elem("BODY")
var A = Elem("A")
var H1 = Elem("H1")
var H2 = Elem("H2")
var H3 = Elem("H3")
var H4 = Elem("H4")
var H5 = Elem("H5")
var DT = Elem("DT")
var EM = Elem("EM")
var STAR AllElementSelector

//
// VALUES
//

type ColorValue struct {
	RGB int
}

func (self ColorValue) String() string {
	return fmt.Sprintf("#%06x", self.RGB)
}

type FontFamilyValue struct {
	Name []string
}

func (self FontFamilyValue) String() string {
	result := ""
	for _, f := range self.Name {
		if strings.Contains(f, " ") {
			result += fmt.Sprintf(`"%s",`, f)
		} else {
			result += fmt.Sprintf("%s,", f)
		}
	}
	return strings.TrimRight(result, ",")
}

//we separate FontSizeValue just to discourage use of specific font sizes
type FontSizeValue struct {
	em float32
	//px and pt are bad, use em
}

func (self FontSizeValue) String() string {
	return fmt.Sprintf("%.2fem", self.em)
}

type DisplayValue struct {
	none               bool
	inline             bool
	table              bool
	run_in             bool
	table_caption      bool
	table_cell         bool
	table_column       bool
	table_column_group bool
	table_footer_group bool
	table_hader_group  bool
	table_row          bool
	table_row_group    bool
	inherit            bool
}

func (self DisplayValue) String() string {
	switch {
	case self.none:
		return "none"
	case self.inline:
		return "none"
	case self.table:
		return "table"
	case self.table_caption:
		return "table-caption"
	case self.table_cell:
		return "table-cell"
	case self.table_column:
		return "table-column"
	case self.table_row:
		return "table-row"
	case self.inherit:
		return "inherit"
	}
	panic("no table display value set!")
}

type TextAlignValue struct {
	left    bool
	right   bool
	center  bool
	justify bool
	inherit bool
}

func (self TextAlignValue) String() string {
	switch {
	case self.left:
		return "left"
	case self.right:
		return "right"
	case self.center:
		return "center"
	case self.justify:
		return "justify"
	case self.inherit:
		return "inherit"
	}
	panic("no text align value set!")

}

type TextDecorationValue struct {
	none         bool
	underline    bool
	overline     bool
	line_through bool
	blink        bool
	inherit      bool
}

func (self TextDecorationValue) String() string {
	switch {
	case self.none:
		return "none"
	case self.underline:
		return "underline"
	case self.overline:
		return "overline"
	case self.line_through:
		return "line-through"
	case self.inherit:
		return "inherit"
	}
	panic("no text decoration value set!")

}

type SizeValue struct {
	Px int
	Em float32
	Mm int
	Cm int
}

func (self SizeValue) String() string {
	switch {
	case self.Px != 0:
		return fmt.Sprintf("%dpx", self.Px)
	case self.Em != float32(0):
		return fmt.Sprintf("%.2fem", self.Em)
	case self.Mm != 0:
		return fmt.Sprintf("%dmm", self.Mm)
	case self.Cm != 0:
		return fmt.Sprintf("%dcm", self.Cm)
	}
	panic("no size value set!")
}

//has to be separate from other sizes because there are multiple ways to set it
type BorderWidthValue struct {
	V []SizeValue
}

func (self BorderWidthValue) String() string {
	result := ""
	for _, i := range self.V {
		result += " " + fmt.Sprintf("%v",i)
	}
	return strings.TrimLeft(result, " ")
}

func AllBorderWidth(v SizeValue) BorderWidthValue {
	return BorderWidthValue{V: []SizeValue{v}}
}
func TopBottomAndLeftRightBorderWidth(tb SizeValue, lr SizeValue) BorderWidthValue {
	return BorderWidthValue{V: []SizeValue{tb, lr}}
}
func TopAndBottomWithLeftRightSameBorderWidth(top SizeValue, leftright SizeValue, bot SizeValue) BorderWidthValue {
	return BorderWidthValue{V: []SizeValue{top, leftright, bot}}
}

func BorderWidths(top SizeValue, right SizeValue, bot SizeValue, left SizeValue) BorderWidthValue {
	return BorderWidthValue{V: []SizeValue{top,right,bot,left}}
}

type BorderStyleValueRaw struct {
	none    bool
	hidden  bool
	dotted  bool
	dashed  bool
	solid   bool
	groove  bool
	dbl  bool
	ridge   bool
	inset   bool
	outset  bool
	inherit bool
}

func (self BorderStyleValueRaw) String() string {
	switch {
	case self.none:
		return "none"
	case self.hidden:
		return "hidden"
	case self.dotted:
		return "dotted"
	case self.dashed:
		return "dashed"
	case self.solid:
		return "solid"
	case self.groove:
		return "groove"
	case self.dbl:
		return "double"
	case self.ridge:
		return "ridge"
	case self.inset:
		return "inset"
	case self.inset:
		return "outsite"
	case self.inherit:
		return "inherit"
	}
	panic("no border style set!")
}

//has to be separate from other sizes because there are multiple ways to set it
type BorderStyleValue struct {
	V []BorderStyleValueRaw
}

func (self BorderStyleValue) String() string {
	result := ""
	for _, i := range self.V {
		result += " " + fmt.Sprintf("%v",i)
	}
	return strings.TrimLeft(result, " ")
}

func AllBorderStyle(v BorderStyleValueRaw) BorderStyleValue {
	return BorderStyleValue{V: []BorderStyleValueRaw{v}}
}
func TopBottomAndLeftRightBorderStyle(tb BorderStyleValueRaw, lr BorderStyleValueRaw) BorderStyleValue {
	return BorderStyleValue{V: []BorderStyleValueRaw{tb, lr}}
}
func TopAndBottomWithLeftAndRightSameBorderStyle(top BorderStyleValueRaw, leftright BorderStyleValueRaw, bot BorderStyleValueRaw) BorderStyleValue {
	return BorderStyleValue{V: []BorderStyleValueRaw{top, leftright, bot}}
}
func BorderStyles(top BorderStyleValueRaw, right BorderStyleValueRaw, bot BorderStyleValueRaw, left BorderStyleValueRaw) BorderStyleValue {
	return BorderStyleValue{V: []BorderStyleValueRaw{top, right, bot, left}}
}

type BorderColorValue struct {
	V []ColorValue
}

func (self BorderColorValue) String() string {
	result := ""
	for _, i := range self.V {
		result += " " + fmt.Sprintf("%v",i)
	}
	return strings.TrimLeft(result, " ")
}

func AllBorderColor(v ColorValue) BorderColorValue {
	return BorderColorValue{V: []ColorValue{v}}
}
func TopBottomAndLeftRightBorderColor(tb ColorValue, lr ColorValue) BorderColorValue {
	return BorderColorValue{V: []ColorValue{tb, lr}}
}
func TopAndBottomWithLeftAndRightSameBorderColor(top ColorValue, leftright ColorValue, bot ColorValue) BorderColorValue {
	return BorderColorValue{V: []ColorValue{top, leftright, bot}}
}
func BorderColors(top ColorValue, right ColorValue, bot ColorValue, left ColorValue) BorderColorValue {
	return BorderColorValue{V: []ColorValue{top, right, bot, left}}
}


//
// VALUE CONSTANTS
//

//colors http://www.w3schools.com/cssref/css_colornames.asp
var AliceBlue = ColorValue{0xF0F8FF}
var AntiqueWhite = ColorValue{0xFAEBD7}
var Aqua = ColorValue{0x00FFFF}
var Red = ColorValue{0xFF0000}
var Black = ColorValue{0x000000}

var Gray99 = ColorValue{0x999999}
var Gray33 = ColorValue{0x333333}

//fonts http://www.w3.org/Style/Examples/007/fonts.en.html
var Helvetica = FontFamilyValue{[]string{"Helvetica", "sans-serif"}}
var Verdana = FontFamilyValue{[]string{"Verdana", "sans-serif"}}
var GillSans = FontFamilyValue{[]string{"Gill Sans", "sans-serif"}}
var Avantgarde = FontFamilyValue{[]string{"Avantgarde", "sans-serif"}}
var HelveticaNarrow = FontFamilyValue{[]string{"Helvetica Narrow", "sans-serif"}}
var Times = FontFamilyValue{[]string{"Times", "serif"}}
var TimesNewRoman = FontFamilyValue{[]string{"Times New Roman", "serif"}}
var Palatino = FontFamilyValue{[]string{"Palatino", "serif"}}
var Bookman = FontFamilyValue{[]string{"Bookman", "serif"}}
var NewCenturySchoolbook = FontFamilyValue{[]string{"New Century Schoolbook", "serif"}}
var AndaleMono = FontFamilyValue{[]string{"Andale Mono", "monospace"}}
var CourierNew = FontFamilyValue{[]string{"Courier New", "monospace"}}
var Courier = FontFamilyValue{[]string{"Courier", "monospace"}}
var LucidaTypewriter = FontFamilyValue{[]string{"Lucidatypewriter", "monospace"}}
var ComicSans = FontFamilyValue{[]string{"Comic Sans", "Comic Sans MS", "cursive"}}
var ZapfChancery = FontFamilyValue{[]string{"Zapf Chancery", "cursive"}}
var Coronetscript = FontFamilyValue{[]string{"Coronetscript", "cursive"}}
var Florence = FontFamilyValue{[]string{"Florence", "cursive"}}
var Impact = FontFamilyValue{[]string{"Impact", "fantasy"}}
var Arnoldboecklin = FontFamilyValue{[]string{"Arnoldboecklin", "fantasy"}}
var Oldtown = FontFamilyValue{[]string{"Oldtown", "fantasy"}}
var Blippo = FontFamilyValue{[]string{"Blippo", "fantasy"}}
var Brushstroke = FontFamilyValue{[]string{"Brushstroke", "fantasy"}}

var Monospace = FontFamilyValue{[]string{"monospace"}}
var Serif = FontFamilyValue{[]string{"serif"}}
var SansSerif = FontFamilyValue{[]string{"sans serif"}}
var Fantasy = FontFamilyValue{[]string{"fantasy"}}
var Cursive = FontFamilyValue{[]string{"cursive"}}

var DoubleSize = FontSizeValue{2.0}
var NormalSize = FontSizeValue{1.0}
var OneAndHalfSize = FontSizeValue{1.5}
var OneAndQuarterSize = FontSizeValue{1.25}
var HalfSize = FontSizeValue{0.5}
var TwoAndHalfSize = FontSizeValue{2.5}

var OnePix = SizeValue{Px: 1}
var TwoPix = SizeValue{Px: 2}
var ThreePix = SizeValue{Px: 3}
var FourPix = SizeValue{Px: 4}
var FivePix = SizeValue{Px: 5}

var NoBorderStyle = BorderStyleValueRaw{none:true}
var HiddenBorderStyle = BorderStyleValueRaw{hidden:true}
var DottedBorderStyle = BorderStyleValueRaw{dotted:true}
var DashedBorderStyle = BorderStyleValueRaw{dashed:true}
var SolidBorderStyle = BorderStyleValueRaw{solid:true}
var GrooveBorderStyle = BorderStyleValueRaw{groove:true}
var DoubleBorderStyle = BorderStyleValueRaw{dbl:true}
var RidgeBorderStyle = BorderStyleValueRaw{ridge:true}
var InsetBorderStyle = BorderStyleValueRaw{inset:true}
var OutsetBorderStyle = BorderStyleValueRaw{outset:true}
var InheritBorderStyle = BorderStyleValueRaw{inherit:true}

//
// ATTRIBUTES
//

type Attr interface {
	attrTag()
}

type Color struct {
	Value ColorValue
}

func (self Color) String() string {
	return fmt.Sprintf("color: %v", self.Value)
}

func (self Color) attrTag() {
}

type FontFamily struct {
	Value FontFamilyValue
}

func (self FontFamily) String() string {
	return fmt.Sprintf("font-family: %v", self.Value)
}

func (self FontFamily) attrTag() {
}

type FontSize struct {
	Value FontSizeValue
}

func (self FontSize) String() string {
	return fmt.Sprintf("font-size: %v", self.Value)
}

func (self FontSize) attrTag() {
}

type Display struct {
	Value DisplayValue
}

func (self Display) String() string {
	return fmt.Sprintf("display: %v", self.Value)
}

func (self Display) attrTag() {
}

type TextAlign struct {
	Value TextAlignValue
}

func (self TextAlign) String() string {
	return fmt.Sprintf("text-align: %v", self.Value)
}

func (self TextAlign) attrTag() {
}

type TextDecoration struct {
	Value TextDecorationValue
}

func (self TextDecoration) String() string {
	return fmt.Sprintf("text-decoration: %v", self.Value)
}

func (self TextDecoration) attrTag() {
}

type Border struct {
	Width BorderWidthValue
	Style BorderStyleValue
	Color BorderColorValue
}

func (self Border) String() string {
	if len(self.Width.V)==1 && len(self.Style.V)==1 && len(self.Color.V)==1 {
		return fmt.Sprintf("border: %v %v %v", self.Width, self.Style, self.Color)
	}
	panic("not sure how to handle more complex border decls")
}

func (self Border) attrTag() {
}

func AllBorders(width SizeValue, style BorderStyleValueRaw, color ColorValue) Border {
	var result Border
	result.Width = AllBorderWidth(width)
	result.Style = AllBorderStyle(style)
	result.Color = AllBorderColor(color)
	return result
}

//
// COMPOSITE CONSTANTS
//

var DisplayNone = Display{DisplayValue{none: true}}

var TextAlignRight = TextAlign{TextAlignValue{right: true}}
var TextAlignCenter = TextAlign{TextAlignValue{center: true}}
var TextAlignLeft = TextAlign{TextAlignValue{left: true}}
var TextAlignJustify = TextAlign{TextAlignValue{justify: true}}
var TextAlignInherit = TextAlign{TextAlignValue{inherit: true}}

var TextDecorationNone = TextDecoration{TextDecorationValue{none: true}}
var TextDecorationUnderline = TextDecoration{TextDecorationValue{underline: true}}
var TextDecorationOverline = TextDecoration{TextDecorationValue{overline: true}}
var TextDecorationBlink = TextDecoration{TextDecorationValue{blink: true}}
var TextDecorationLineThrough = TextDecoration{TextDecorationValue{line_through: true}}
var TextDecorationInherit = TextDecoration{TextDecorationValue{inherit: true}}

//
// Statement
//

type Stmt_ interface {
	stmtTag()
}

//simple case has 1 of each, so this is
//#header {
//    font-family:monospace;
//}
type Stmt struct {
	S Stor
	A Attr
}

func (self Stmt) String() string {
	result := fmt.Sprintf("%v {\n", self.S)
	result += fmt.Sprintf("\t%v;\n", self.A)
	return result + "}\n"
}
func (self Stmt) stmtTag() {
}

//one selector, more than one attr
//a {
//    color:#999999;
//    text-decoration:none;
//}
type StmtN struct {
	S Stor
	A []Attr
}

func (self StmtN) String() string {
	result := fmt.Sprintf("%v {\n", self.S)
	for _, a := range self.A {
		result += fmt.Sprintf("\t%v;\n", a)
	}
	return result + "}\n"
}
func (self StmtN) stmtTag() {
}

//more than one selector, more than one attr
//h1, h2 {
//    color:#999999;
//    text-decoration:none;
//}

type NStmtN struct {
	S []Stor
	A []Attr
}

func (self NStmtN) String() string {
	result := ""
	for _, s := range self.S {
		result += fmt.Sprintf("%v,", s)
	}
	result = strings.TrimRight(result, ",")

	result = fmt.Sprintf("%v {", result)
	for _, a := range self.A {
		result += fmt.Sprintf("\t%v;\n", a)
	}
	return result + "}\n"
}
func (self NStmtN) stmtTag() {
}

type StmtSeq []Stmt_

func (self StmtSeq) String() string {
	result := ""
	for _, s := range self {
		result += fmt.Sprintf("%v", s)
	}
	return result
}
