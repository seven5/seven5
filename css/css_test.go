package css

import (
	"launchpad.net/gocheck"
	"testing"
	"strings"
)

// Hook up gocheck into the default gotest runner.
func Test(t *testing.T) { gocheck.TestingT(t) }

// This is the "suite" structure for objects that need to survive the whole of the tests
// in this file.
type CSSSuite struct {

}

// hook up suite to gocheck
var _ = gocheck.Suite(&CSSSuite{})

func (self *CSSSuite) SetUpSuite(c *gocheck.C) {
}

func (self *CSSSuite) TearDownSuite(c *gocheck.C) {
}

var basic = ClassStyle{
	"content-navigation",
	Style{
		//shorthand 
		BorderColor{0x3bbfce},
		//more explicit
		Color{Rgb: 0x2b9eabe},
	},
	nil, //no inheritance... this nil is annoying but it can be avoided, see next example
}

var basicExpected = ".content-navigation {\n\tborder-color: #3bbfce;\n\tcolor: #2b9eabe;\n}\n"

const blue = 0x3bbfce

var basic2 = ClassStyle{Class: CSSClass("border"),
	Style: Style{
		Padding{All: Size{Px: 16}},
		Margin{All: Size{Px: 16}},
		BorderColor{blue},
	},
}

var basic2Expected = ".border {\n\tpadding: 16px;\n\tmargin: 16px;\n\tborder-color: #3bbfce;\n}\n"

func (s *CSSSuite) TestBasicTypes(c *gocheck.C) {
	c.Check(basic.String(), gocheck.Equals, basicExpected)
	c.Check(basic2.String(), gocheck.Equals, basic2Expected)
}

var noSize = ClassStyle{Class: CSSClass("border"),
	Style: Style{
		Padding {},
	},
}


var twoSize = ClassStyle{Class: CSSClass("border"),
	Style: Style{Padding{All: Size{Px:1, Pt:1}}},
}

var allTop = ClassStyle{Class: CSSClass("border"),
			Style: Style{
			Padding{All: Size{Px:1},
					TopBottom: Size{Px:1}}},
}

var justTop = ClassStyle{Class: CSSClass("border"),
			Style: Style{
			Padding{TopBottom: Size{Px:1}}},
}

func (s *CSSSuite) TestBadBoxProps(c *gocheck.C) {
	c.Check(strings.Contains(noSize.String(),"PANIC"), gocheck.Not(gocheck.Equals), -1)
	c.Check(strings.Contains(twoSize.String(),"PANIC"), gocheck.Not(gocheck.Equals), -1)
	c.Check(strings.Contains(allTop.String(),"PANIC"), gocheck.Not(gocheck.Equals), -1)
	c.Check(strings.Contains(justTop.String(),"PANIC"), gocheck.Not(gocheck.Equals), -1)
}
