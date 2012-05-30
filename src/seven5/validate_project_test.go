package seven5

import (
	. "launchpad.net/gocheck"
	"seven5/util"
	"os"
	"fmt"
)

// Hook up gocheck into the gotest runner.
//func Test(t *testing.T) { TestingT(t) }

//suite of tests
type vpSuite struct{
	logger util.SimpleLogger 
}
var _ = Suite(&vpSuite{})

func (self *vpSuite) SetUpSuite(c *C) {
	self.logger = util.NewTerminalLogger(os.Stdout, util.DEBUG, true)
}

func (self *vpSuite) TestValidator(c *C) {
	checkList := []bool{ true, false, false, false, false}
	vp := &DefaultValidateProject{}
	
	for i, check := range(checkList) {
		dir := util.FindTestDataPath(c,"projectvalidator",fmt.Sprintf("testproj%d",i+1))
		appConfig, err:= decodeAppConfig(dir)
		if err!=nil{
			if i!=1 /*expected on test2*/{
				c.Errorf("Could not read application config in %s: %s",dir, err)
			}
			continue
		}
		raw := vp.Exec("ignored", dir, appConfig, nil /*ignored*/, self.logger)
		result := raw.(*ValidateProjectResult)	
		//we reverse the logic because we are checking for the *error* flag
		if check {
			c.Check(result.Error, Equals, false)
		} else {
			c.Check(result.Error, Equals, true)
		}
	}
}

