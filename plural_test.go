package seven5

import (
	"launchpad.net/gocheck"
	"testing"
)
// Hook up gocheck into the default gotest runner.
func Test(t *testing.T) { gocheck.TestingT(t) }

// This is the "suite" structure for objects that need to survive the whole of the tests
// in this file.
type PluralSuite struct {
}

var p = &PluralSuite{}

// hook up suite to gocheck
var _ = gocheck.Suite(p)

//TestPlural from http://code.activestate.com/recipes/577781-pluralize-word-convert-singular-word-to-its-plural/
func (self *PluralSuite) TestPlural(c *gocheck.C) {
	c.Check(Pluralize(""),gocheck.Equals,"")
	c.Check(Pluralize("goose"),gocheck.Equals,"geese")
	c.Check(Pluralize("dolly"),gocheck.Equals,"dollies")
	c.Check(Pluralize("genius"),gocheck.Equals,"genii")
	c.Check(Pluralize("jones"),gocheck.Equals,"joneses")
	c.Check(Pluralize("pass"),gocheck.Equals,"passes")
	c.Check(Pluralize("zero"),gocheck.Equals,"zeros")
	c.Check(Pluralize("casino"),gocheck.Equals,"casinos")
	c.Check(Pluralize("hero"),gocheck.Equals,"heroes")
	c.Check(Pluralize("church"),gocheck.Equals,"churches")
	c.Check(Pluralize("x"),gocheck.Equals,"xs")
	c.Check(Pluralize("car"),gocheck.Equals,"cars")
}