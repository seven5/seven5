package dsl

import (
	"fmt"
	"launchpad.net/gocheck"
	"testing"
)

// Hook up gocheck into the default gotest runner.
func Test(t *testing.T) { gocheck.TestingT(t) }

// This is the "suite" structure for objects that need to survive the whole of the tests
// in this file.
type CSSSuite struct {

}

const NO_LIMIT = -39

// hook up suite to gocheck
var _ = gocheck.Suite(&CSSSuite{})

func (self *CSSSuite) SetUpSuite(c *gocheck.C) {
}

func (self *CSSSuite) TearDownSuite(c *gocheck.C) {
}

func (self *CSSSuite) TestBasic(c *gocheck.C) {
	urgent := Stmt{Class("urgent"), Color{Red}}
	c.Assert(fmt.Sprintf("%v",urgent), gocheck.Equals, ".urgent {\n\tcolor: #ff0000;\n}\n")
	
	body:=Stmt{BODY,FontFamily{GillSans}}
	c.Check(fmt.Sprintf("%v",body), gocheck.Equals, "BODY {\n\tfont-family: \"Gill Sans\",sans-serif;\n}\n")
	
	header:=Stmt{Id("header"),FontFamily{Monospace}}
	c.Check(fmt.Sprintf("%v",header), gocheck.Equals, "#header {\n\tfont-family: monospace;\n}\n")

	headerh1:=Stmt{Selector(Id("header"),H1),FontSize{DoubleSize}}
	c.Check(fmt.Sprintf("%s",headerh1), gocheck.Equals, "#header H1 {\n\tfont-size: 2.00em;\n}\n")
	
}

func (self *CSSSuite) TestMultipleAttrs(c *gocheck.C) {
	
	sidebar:=StmtN{Id("sidebar"),[]Attr{DisplayNone,TextAlignRight}}
	c.Check(fmt.Sprintf("%s",sidebar), gocheck.Equals, "#sidebar {\n\tdisplay: none;\n\ttext-align: right;\n}\n")

	sidebardt:=StmtN{Selector(Id("sidebar"),DT),[]Attr{FontFamily{Monospace},FontSize{OneAndQuarterSize}}}
	c.Check(fmt.Sprintf("%s",sidebardt), gocheck.Equals, "#sidebar DT {\n\tfont-family: monospace;\n\tfont-size: 1.25em;\n}\n")

	a:=StmtN{A,[]Attr{Color{ColorValue{0x999999}},TextDecorationNone}}
	c.Check(fmt.Sprintf("%s",a), gocheck.Equals, "A {\n\tcolor: #999999;\n\ttext-decoration: none;\n}\n")

}

func (self *CSSSuite) TestBorder(c *gocheck.C) {
	foo:=Stmt{Id("foo"),AllBorders(OnePix,SolidBorderStyle,Black)}
	c.Check(fmt.Sprintf("%s",foo), gocheck.Equals, "#foo {\n\tborder: 1px solid #000000;\n}\n")

	bar1:=Stmt{Id("bar"),Border{Style:TopBottomAndLeftRightBorderStyle(DashedBorderStyle, DottedBorderStyle)}}
	bar2:=Stmt{Id("bar"),BorderStyle{TopBottomAndLeftRightBorderStyle(DashedBorderStyle, DottedBorderStyle)}}
	expect:="#bar {\n\tborder-style: dashed dotted;\n}\n"
	
	c.Check(fmt.Sprintf("%s",bar1), gocheck.Equals, expect)
	c.Check(fmt.Sprintf("%s",bar2), gocheck.Equals, expect)

}

func (self *CSSSuite) TestPrintStructFieldHandlesUnderscores(c *gocheck.C) {
	baz:=Stmt{Id("baz"),OverflowYNoContent}
	c.Check(fmt.Sprintf("%s",baz), gocheck.Equals, "#baz {\n\toverflow-y: no-content;\n}\n")
	
}
