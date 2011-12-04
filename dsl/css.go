package dsl

import (
	//	"bytes"
	"fmt"
	"reflect"
	"strings"
	//"os"
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
	Em     float32
	Px_bad int
	Pt_bad int
	//px and pt are bad, use em
}

func (self FontSizeValue) String() string {
	switch {
	case self.Em != float32(0):
		return fmt.Sprintf("%.2fem", self.Em)
	case self.Px_bad != 0:
		return fmt.Sprintf("%dpx", self.Px_bad)
	case self.Pt_bad != 0:
		return fmt.Sprintf("%dpt", self.Pt_bad)
	}
	panic("no font size set!")
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
	return TrueFieldPrinter(self)
}

type TextAlignValue struct {
	left    bool
	right   bool
	center  bool
	justify bool
	inherit bool
}

func (self TextAlignValue) String() string {
	return TrueFieldPrinter(self)
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
	return TrueFieldPrinter(self)
}

type SizeValue struct {
	Px      int
	Em      float32
	Mm      int
	Cm      int
	Percent int
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
	case self.Percent != 0:
		return fmt.Sprintf("%d%%", self.Percent)
	}
	panic("no size value set!")
}

type FloatValue struct {
	left    bool
	right   bool
	none    bool
	inherit bool
}

func (self FloatValue) String() string {
	return TrueFieldPrinter(self)
}

/////////////////////  BORDER is a complex composite

//has to be separate from other sizes because there are multiple ways to set it
type BorderWidthValue struct {
	V []SizeValue
}

func (self BorderWidthValue) String() string {
	return SpacePrinter(self.V)
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
	return BorderWidthValue{V: []SizeValue{top, right, bot, left}}
}

type BorderStyleValueRaw struct {
	none    bool
	hidden  bool
	dotted  bool
	dashed  bool
	solid   bool
	groove  bool
	dbl     bool
	ridge   bool
	inset   bool
	outset  bool
	inherit bool
}

func (self BorderStyleValueRaw) String() string {
	return TrueFieldPrinter(self)
}

//has to be separate from other sizes because there are multiple ways to set it
type BorderStyleValue struct {
	V []BorderStyleValueRaw
}

func (self BorderStyleValue) String() string {
	return SpacePrinter(self.V)
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
	return SpacePrinter(self.V)
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

type OverflowValue struct {
	visible    bool
	hidden     bool
	scroll     bool
	auto       bool
	no_display bool
	no_content bool
}

func (self OverflowValue) String() string {
	return TrueFieldPrinter(self)
}

//Margin
type MarginValue struct {
	V []SizeValue
}

func (self MarginValue) String() string {
	return SpacePrinter(self.V)
}

//Padding
type PaddingValue struct {
	V []SizeValue
}

func (self PaddingValue) String() string {
	return SpacePrinter(self.V)
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
var Gray = ColorValue{0x808080}

var Gray99 = ColorValue{0x999999}
var Gray33 = ColorValue{0x333333}
var Gray44 = ColorValue{0x444444}
var Gray80 = ColorValue{0x808080}

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

var DoubleSize = FontSizeValue{Em: 2.0}
var NormalSize = FontSizeValue{Em: 1.0}
var OneAndHalfSize = FontSizeValue{Em: 1.5}
var OneAndQuarterSize = FontSizeValue{Em: 1.25}
var HalfSize = FontSizeValue{Em: 0.5}
var TwoAndHalfSize = FontSizeValue{Em: 2.5}

var OnePix = SizeValue{Px: 1}
var TwoPix = SizeValue{Px: 2}
var ThreePix = SizeValue{Px: 3}
var FourPix = SizeValue{Px: 4}
var FivePix = SizeValue{Px: 5}

var NoBorderStyle = BorderStyleValueRaw{none: true}
var HiddenBorderStyle = BorderStyleValueRaw{hidden: true}
var DottedBorderStyle = BorderStyleValueRaw{dotted: true}
var DashedBorderStyle = BorderStyleValueRaw{dashed: true}
var SolidBorderStyle = BorderStyleValueRaw{solid: true}
var GrooveBorderStyle = BorderStyleValueRaw{groove: true}
var DoubleBorderStyle = BorderStyleValueRaw{dbl: true}
var RidgeBorderStyle = BorderStyleValueRaw{ridge: true}
var InsetBorderStyle = BorderStyleValueRaw{inset: true}
var OutsetBorderStyle = BorderStyleValueRaw{outset: true}
var InheritBorderStyle = BorderStyleValueRaw{inherit: true}

//
// ATTRIBUTES
//

type Attr interface {
	attrTag()
}

//COLOR
type Color struct {
	Value ColorValue
}

func (self Color) String() string {
	return PropPrinter("color", self.Value)
}
func (self Color) attrTag() {
}

//FONT FAMILY
type FontFamily struct {
	Value FontFamilyValue
}

func (self FontFamily) String() string {
	return PropPrinter("font-family", self.Value)
}
func (self FontFamily) attrTag() {
}

//FONT SIZE
type FontSize struct {
	Value FontSizeValue
}

func (self FontSize) String() string {
	return PropPrinter("font-size", self.Value)
}
func (self FontSize) attrTag() {
}

//DISPLAY
type Display struct {
	Value DisplayValue
}

func (self Display) String() string {
	return PropPrinter("display", self.Value)
}
func (self Display) attrTag() {
}

//TEXT ALIGN
type TextAlign struct {
	Value TextAlignValue
}

func (self TextAlign) String() string {
	return PropPrinter("text-align", self.Value)
}
func (self TextAlign) attrTag() {
}

//TEXT DECORATION
type TextDecoration struct {
	Value TextDecorationValue
}

func (self TextDecoration) String() string {
	return PropPrinter("text-decoration", self.Value)
}
func (self TextDecoration) attrTag() {
}

//
// BORDER
//
type Border struct {
	Width BorderWidthValue
	Style BorderStyleValue
	Color BorderColorValue
}

func (self Border) String() string {
	if len(self.Width.V) == 1 && len(self.Style.V) == 1 && len(self.Color.V) == 1 {
		return fmt.Sprintf("border: %s %s %s", self.Width.V[0], self.Style.V[0], self.Color.V[0])
	}
	result := ""
	if len(self.Width.V) > 0 {
		result += PropPrinter("border-width", self.Width) + ";"
	}
	if len(self.Style.V) > 0 {
		result += PropPrinter("border-style", self.Style) + ";"
	}
	if len(self.Color.V) > 0 {
		result += PropPrinter("border-color", self.Style) + ";"
	}
	if len(self.Width.V) == 0 && len(self.Style.V) == 0 && len(self.Color.V) == 0 {
		panic("no values set in the border!")
	}
	return strings.TrimRight(result, ";")
}
func (self Border) attrTag() {
}

// Border: BorderWidth
type BorderWidth struct {
	Width BorderWidthValue
}

func (self BorderWidth) String() string {
	return PropPrinter("border-width", self.Width)
}
func (self BorderWidth) attrTag() {
}

// Border: BorderStyle
type BorderStyle struct {
	Style BorderStyleValue
}

func (self BorderStyle) String() string {
	return PropPrinter("border-style", self.Style)
}
func (self BorderStyle) attrTag() {
}

//Border: BorderColor
type BorderColor struct {
	Color BorderColorValue
}

func (self BorderColor) String() string {
	return PropPrinter("border-color", self.Color)
}
func (self BorderColor) attrTag() {
}

//Set all three parts to a single value
func AllBorders(width SizeValue, style BorderStyleValueRaw, color ColorValue) Border {
	var result Border
	result.Width = AllBorderWidth(width)
	result.Style = AllBorderStyle(style)
	result.Color = AllBorderColor(color)
	return result
}

//Width
type Width struct {
	Value SizeValue
}

func (self Width) String() string {
	return PropPrinter("width", self.Value)
}
func (self Width) attrTag() {
}
//Height
type Height struct {
	Value SizeValue
}

func (self Height) String() string {
	return PropPrinter("height", self.Value)
}
func (self Height) attrTag() {
}

//Overflow-Y
type OverflowY struct {
	Value OverflowValue
}

func (self OverflowY) String() string {
	return PropPrinter("overflow-y", self.Value)
}

func (self OverflowY) attrTag() {
}

//Overflow-X
type OverflowX struct {
	Value OverflowValue
}

func (self OverflowX) String() string {
	return PropPrinter("overflow-x", self.Value)
}
func (self OverflowX) attrTag() {
}

//Float
type Float struct {
	Value FloatValue
}

func (self Float) String() string {
	return PropPrinter("float", self.Value)
}
func (self Float) attrTag() {
}

//Margin
type Margin struct {
	Value MarginValue
}

func (self Margin) String() string {
	return PropPrinter("margin", self.Value)
}
func (self Margin) attrTag() {
}

func AllMargins(v SizeValue) Margin {
	return Margin{MarginValue{V: []SizeValue{v}}}
}
func TopBottomAndLeftRightMargin(tb SizeValue, lr SizeValue) Margin {
	return Margin{MarginValue{V: []SizeValue{tb, lr}}}
}
func TopAndBottomWithLeftRightSameMargin(top SizeValue, leftright SizeValue, bot SizeValue) Margin {
	return Margin{MarginValue{V: []SizeValue{top, leftright, bot}}}
}
func Margins(top SizeValue, right SizeValue, bot SizeValue, left SizeValue) Margin {
	return Margin{MarginValue{V: []SizeValue{top, right, bot, left}}}
}

//Padding
type Padding struct {
	V PaddingValue
}

func (self Padding) String() string {
	return PropPrinter("padding", self.V)
}
func (self Padding) attrTag() {
}

func AllPadding(v SizeValue) Padding {
	return Padding{PaddingValue{V: []SizeValue{v}}}
}
func TopBottomAndLeftRightPadding(tb SizeValue, lr SizeValue) Padding {
	return Padding{PaddingValue{V: []SizeValue{tb, lr}}}
}
func TopAndBottomWithLeftRightSamePadding(top SizeValue, leftright SizeValue, bot SizeValue) Padding {
	return Padding{PaddingValue{V: []SizeValue{top, leftright, bot}}}
}
func Paddings(top SizeValue, right SizeValue, bot SizeValue, left SizeValue) Padding {
	return Padding{PaddingValue{V: []SizeValue{top, right, bot, left}}}
}
//Margin-top
type MarginTop struct {
	Value SizeValue
}

func (self MarginTop) String() string {
	return PropPrinter("margin-top", self.Value)
}
func (self MarginTop) attrTag() {
}

//Margin-right
type MarginRight struct {
	Value SizeValue
}

func (self MarginRight) String() string {
	return PropPrinter("margin-right", self.Value)
}
func (self MarginRight) attrTag() {
}

//Margin-bottom
type MarginBottom struct {
	Value SizeValue
}

func (self MarginBottom) String() string {
	return PropPrinter("margin-bottom", self.Value)
}
func (self MarginBottom) attrTag() {
}

//Margin-left
type MarginLeft struct {
	Value SizeValue
}

func (self MarginLeft) String() string {
	return PropPrinter("margin-left", self.Value)
}
func (self MarginLeft) attrTag() {
}

//Padding-top
type PaddingTop struct {
	Value SizeValue
}

func (self PaddingTop) String() string {
	return PropPrinter("padding-top", self.Value)
}
func (self PaddingTop) attrTag() {
}

//Padding-right
type PaddingRight struct {
	Value SizeValue
}

func (self PaddingRight) String() string {
	return PropPrinter("padding-right", self.Value)
}
func (self PaddingRight) attrTag() {
}

//Padding-bottom
type PaddingBottom struct {
	Value SizeValue
}

func (self PaddingBottom) String() string {
	return PropPrinter("padding-bottom", self.Value)
}
func (self PaddingBottom) attrTag() {
}

//Padding-left
type PaddingLeft struct {
	Value SizeValue
}

func (self PaddingLeft) String() string {
	return PropPrinter("padding-left", self.Value)
}
func (self PaddingLeft) attrTag() {
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

var OverflowXVisible = OverflowX{OverflowValue{visible: true}}
var OverflowYVisible = OverflowY{OverflowValue{visible: true}}
var OverflowXHidden = OverflowX{OverflowValue{hidden: true}}
var OverflowYHidden = OverflowY{OverflowValue{hidden: true}}
var OverflowXScroll = OverflowX{OverflowValue{scroll: true}}
var OverflowYScroll = OverflowY{OverflowValue{scroll: true}}
var OverflowXAuto = OverflowX{OverflowValue{scroll: true}}
var OverflowYAuto = OverflowY{OverflowValue{scroll: true}}
var OverflowXNoDisplay = OverflowX{OverflowValue{no_display: true}}
var OverflowYNoDisplay = OverflowY{OverflowValue{no_display: true}}
var OverflowXNoContent = OverflowX{OverflowValue{no_content: true}}
var OverflowYNoContent = OverflowY{OverflowValue{no_content: true}}

var FloatLeft = Float{FloatValue{left:true}}
var FloatRight = Float{FloatValue{right:true}}
var FloatNone = Float{FloatValue{none:true}}
var FloatInherit= Float{FloatValue{inherit:true}}

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

type StyleSheet struct {
	Name string  /* for coordination with HTML, not for display*/
	S StmtSeq
}
func (self StyleSheet) String() string {
	return fmt.Sprintf("%s",self.S)
}

func (self StmtSeq) String() string {
	result := ""
	for _, s := range self {
		result += fmt.Sprintf("%v", s)
	}
	return result
}

//
// Utilities
//
func SpacePrinter(raw interface{}) string {
	v := reflect.ValueOf(raw)
	targ := make([]interface{}, v.Len(), v.Cap())
	for i := 0; i < v.Len(); i++ {
		targ[i] = v.Index(i).Interface()
	}
	result := ""
	for _, b := range targ {
		result += " " + fmt.Sprintf("%s", b)
	}
	return fmt.Sprintf("%s", strings.TrimLeft(result, " "))
}

func TrueFieldPrinter(raw interface{}) string {
	t := reflect.TypeOf(raw)
	v := reflect.ValueOf(raw)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Type.Kind() != reflect.Bool {
			continue
		}
		name := f.Name
		if v.Field(i).Bool() {
			return strings.Replace(name, "_", "-", -1 /*NOLIMIT*/ )
		}
	}
	panic("no true field set in struct!")
}

func PropPrinter(prop string, v interface{}) string {
	return fmt.Sprintf("%s: %s", prop, v)
}
