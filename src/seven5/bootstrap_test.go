package seven5

import (
	. "launchpad.net/gocheck"
	"testing"
	"seven5/util"
	"os"
)


// Hook up gocheck into the gotest runner.
func Test(t *testing.T) { TestingT(t) }

//suite of tests
type S struct{
	Logger util.SimpleLogger 
}
var _ = Suite(&S{})

func (self *S) SetUpSuite(c *C) {
	self.Logger = util.NewTerminalLogger(os.Stdout, util.DEBUG, true)
}

func (self *S) TestValidator(c *C) {
	/*
	c.Check(validator.Validate(filepath.Join(testdata,"testproj1"),self.Logger),
		Equals,true)		
	c.Check(validator.Validate(filepath.Join(testdata,"testproj2"),self.Logger),
		Equals,false)
		
	c.Check(validator.Validate(filepath.Join(testdata,"testproj3"),self.Logger),
		Equals,false)
	*/
}

func (self *S) TestGroupieConfig(c *C) {
	t1:=util.ReadTestData(self.Logger, "groupieconfig","test1.json")
	ur:=util.ReadTestData(self.Logger, "groupieconfig","test-unknown-role.json")

	roleName := "ProjectValidator"
	
	conf, err:=parseGroupieConfig(t1)
	c.Assert(err, IsNil)
	c.Assert(len(conf), Not(Equals), 0)
	c.Assert(conf[roleName], Not(IsNil))
	c.Check(conf[roleName].TypeName, Equals, "plugin.DefaultProjectValidator")
	c.Check(len(conf[roleName].ImportsNeeded), Not(Equals), 0)

	_, err=parseGroupieConfig(ur)
	c.Assert(err, Not(IsNil))

}

func (self *S) TestBootstrapPill(c *C) {
	t1:=util.ReadTestData(self.Logger, "groupieconfig","test1.json")
	
}

