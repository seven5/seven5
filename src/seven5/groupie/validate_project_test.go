package groupie

import (
	. "launchpad.net/gocheck"
	"testing"
	"seven5/util"
	"os"
	"fmt"
)

// Hook up gocheck into the gotest runner.
func Test(t *testing.T) { TestingT(t) }

//suite of tests
type S struct{
	logger util.SimpleLogger 
}
var _ = Suite(&S{})

func (self *S) SetUpSuite(c *C) {
	self.logger = util.NewTerminalLogger(os.Stdout, util.DEBUG, true)
}

func (self *S) TestValidator(c *C) {
	checkList := []bool{ true, false, false, false}
	vp := &DefaultValidateProject{}
	
	for i, check := range(checkList) {
		raw := vp.Exec("ignored", util.FindTestDataPath(c,"projectvalidator",
				fmt.Sprintf("testproj%d",i+1)),
				nil, nil, self.logger)
		result := raw.(*ValidateProjectResult)	
		//we reverse the logic because we are checking for the *error* flag
		if check {
			c.Check(result.Error, Equals, false)
		} else {
			c.Check(result.Error, Equals, true)
		}
	}
}

