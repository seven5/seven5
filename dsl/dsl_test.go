package dsl

import (
	"fmt"
	"launchpad.net/gocheck"
	"strings"
	"testing"
)

// Hook up gocheck into the default gotest runner.
func Test(t *testing.T) { gocheck.TestingT(t) }

// This is the "suite" structure for objects that need to survive the whole of the tests
// in this file.
type DSLSuite struct {

}

const NO_LIMIT = -39

// hook up suite to gocheck
var _ = gocheck.Suite(&DSLSuite{})

func (self *DSLSuite) SetUpSuite(c *gocheck.C) {
}

func (self *DSLSuite) TearDownSuite(c *gocheck.C) {
}

func (self *DSLSuite) TestBasic(c *gocheck.C) {
	urgent := Stmt{Class("urgent"), Color{Red}}
	c.Assert(fmt.Sprintf("%v", urgent), gocheck.Equals, ".urgent {\n\tcolor: #ff0000;\n}\n")

	body := Stmt{BODY, FontFamily{GillSans}}
	c.Check(fmt.Sprintf("%v", body), gocheck.Equals, "BODY {\n\tfont-family: \"Gill Sans\",sans-serif;\n}\n")

	header := Stmt{Id("header"), FontFamily{Monospace}}
	c.Check(fmt.Sprintf("%v", header), gocheck.Equals, "#header {\n\tfont-family: monospace;\n}\n")

	headerh1 := Stmt{Selector(Id("header"), H1), FontSize{DoubleSize}}
	c.Check(fmt.Sprintf("%s", headerh1), gocheck.Equals, "#header H1 {\n\tfont-size: 2.00em;\n}\n")

}

func (self *DSLSuite) TestMultipleAttrs(c *gocheck.C) {

	sidebar := StmtN{Id("sidebar"), []Attr{DisplayNone, TextAlignRight}}
	c.Check(fmt.Sprintf("%s", sidebar), gocheck.Equals, "#sidebar {\n\tdisplay: none;\n\ttext-align: right;\n}\n")

	sidebardt := StmtN{Selector(Id("sidebar"), DT), []Attr{FontFamily{Monospace}, FontSize{OneAndQuarterSize}}}
	c.Check(fmt.Sprintf("%s", sidebardt), gocheck.Equals, "#sidebar DT {\n\tfont-family: monospace;\n\tfont-size: 1.25em;\n}\n")

	a := StmtN{A, []Attr{Color{ColorValue{0x999999}}, TextDecorationNone}}
	c.Check(fmt.Sprintf("%s", a), gocheck.Equals, "A {\n\tcolor: #999999;\n\ttext-decoration: none;\n}\n")

}

func (self *DSLSuite) TestBorder(c *gocheck.C) {
	foo := Stmt{Id("foo"), AllBorders(OnePix, SolidBorderStyle, Black)}
	c.Check(fmt.Sprintf("%s", foo), gocheck.Equals, "#foo {\n\tborder: 1px solid #000000;\n}\n")

	bar1 := Stmt{Id("bar"), Border{Style: TopBottomAndLeftRightBorderStyle(DashedBorderStyle, DottedBorderStyle)}}
	bar2 := Stmt{Id("bar"), BorderStyle{TopBottomAndLeftRightBorderStyle(DashedBorderStyle, DottedBorderStyle)}}
	expect := "#bar {\n\tborder-style: dashed dotted;\n}\n"

	c.Check(fmt.Sprintf("%s", bar1), gocheck.Equals, expect)
	c.Check(fmt.Sprintf("%s", bar2), gocheck.Equals, expect)

}

func (self *DSLSuite) TestPrintStructFieldHandlesUnderscores(c *gocheck.C) {
	baz := Stmt{Id("baz"), OverflowYNoContent}
	c.Check(fmt.Sprintf("%s", baz), gocheck.Equals, "#baz {\n\toverflow-y: no-content;\n}\n")

}

func (self *DSLSuite) TestClearAfterRow(c *gocheck.C) {

	row := Row{}
	c.Check(1, gocheck.Equals, len(row.Info(true)))
	c.Check(1, gocheck.Equals, len(row.Info(false)))
	c.Check("clear", gocheck.Equals, row.Info(true)[0].Class[0])

	row = Row{[]Kids{
		Column{[]Kids{Box{Width: 2, Suffix: 1}, Box{Width: 2, Suffix: 1}}},
		Column{[]Kids{Box{Width: 5, Suffix: 1}, Box{Width: 5, Suffix: 1}}},
		Column{[]Kids{Box{Width: 2, Prefix: 1}}},
	}}

	c.Check("clear", gocheck.Equals, row.Info(true)[len(row.Info(true))-1].Class[0])
	for i, r := range row.Info(true) {
		if i == len(row.Info(true))-1 { //skip last one, it SHOULD be a clear
			continue
		}
		doesNotHaveClass(r,"clear",c)
	}
}

func doesNotHaveClass(info *ClassIdInfo, target string, c *gocheck.C) {
	for _, j := range info.Class {
		c.Check(strings.Index(j, target), gocheck.Equals, -1)
	}

}
func hasClass(info *ClassIdInfo, target string, c *gocheck.C) {
	for _, j := range info.Class {
		if j==target {
			return
		}
	}
	c.Fail()
}

func nestedSizeOf(info []*ClassIdInfo, target int, c *gocheck.C) {
	c.Assert(len(info), gocheck.Equals, 1)
	c.Assert(len(info[0].Nested), gocheck.Equals, target)
	c.Check(len(info[0].Nested), gocheck.Equals, target)
}

func (self *DSLSuite) TestColsEmitAlphaOmega(c *gocheck.C) {

	//with only one child, alpha and omega should never be used
	oneItemCol := Column{[]Kids{Box{Width: 2}}}
	info := oneItemCol.Info(true)
	nestedSizeOf(info,1,c)
	doesNotHaveClass(info[0].Nested[0],"alpha",c)
	doesNotHaveClass(info[0].Nested[0],"omega",c)
	
	info = oneItemCol.Info(false)
	nestedSizeOf(info,1,c)
	doesNotHaveClass(info[0].Nested[0],"alpha",c)
	doesNotHaveClass(info[0].Nested[0],"omega",c)

	twoItemCol := Column{[]Kids{Box{Width: 2},Box{Width:1}}}
	info = twoItemCol.Info(false)
	nestedSizeOf(info,2,c)
	hasClass(info[0].Nested[0],"alpha",c)
	hasClass(info[0].Nested[0],"grid_2",c)
	
	hasClass(info[0].Nested[1],"omega",c)
	hasClass(info[0].Nested[1],"grid_1",c)
	
	threeItemCol := Column{[]Kids{Box{Width: 2}, Box{Width: 1}, Box{Width: 3}}}
	info = threeItemCol.Info(false)
	
	nestedSizeOf(info,3,c)

}
