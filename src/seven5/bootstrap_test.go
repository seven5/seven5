package seven5

import (
	. "launchpad.net/gocheck"
	"testing"
	"fmt"
)


// Hook up gocheck into the gotest runner.
func Test(t *testing.T) { TestingT(t) }

//suite of tests
type S struct{
	Logger SimpleLogger 
}
var _ = Suite(&S{})

func (self *S) SetUpSuite(c *C) {
	self.Logger = NewTerminalLogger(DEBUG, true)
}

const test1 =
`
{
	'projectValidator': 'seven5.plugins.ProjectValidator'
}
`

func (self *S) TestBuildWithValidator(c *C) {
	self.Logger.Debug("starting a test...")
}