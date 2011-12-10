package dsl

import (
//	"fmt"
	"launchpad.net/gocheck"
	"testing"
)

// Hook up gocheck into the default gotest runner.
func Test2(t *testing.T) { gocheck.TestingT(t) }

// This is the "suite" structure for objects that need to survive the whole of the tests
// in this file.
type HtmlSuite struct {

}

// hook up suite to gocheck
var _ = gocheck.Suite(&HtmlSuite{})

func (self *HtmlSuite) SetUpSuite(c *gocheck.C) {
}

func (self *HtmlSuite) TearDownSuite(c *gocheck.C) {
}

func (self *HtmlSuite) TestRecursiveGridSize(c *gocheck.C) {
}