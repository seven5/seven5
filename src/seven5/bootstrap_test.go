package seven5

import (
	. "launchpad.net/gocheck"
	"testing"
	"seven5/util"
)


// Hook up gocheck into the gotest runner.
func Test(t *testing.T) { TestingT(t) }

//suite of tests
type S struct{
	Logger util.SimpleLogger 
}
var _ = Suite(&S{})

func (self *S) SetUpSuite(c *C) {
	self.Logger = util.NewTerminalLogger(util.DEBUG, true)
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
	t1:=util.ReadTestData("groupieconfig","test1.json", self.Logger)
	_=util.ReadTestData("groupieconfig","test1.json", self.Logger)
	_=util.ReadTestData("groupieconfig","test1.json", self.Logger)

	roleName := "ProjectValidator"
	
	conf, err:=ParseGroupieConfig(t1)
	c.Assert(err, IsNil)
	c.Assert(len(conf), Not(Equals), 0)
	c.Assert(conf[roleName], Not(IsNil))
	c.Check(conf[roleName].TypeName, Equals, "plugins.DefaultProjectValidator")
	c.Check(len(conf[roleName].ImportsNeeded), Not(Equals), 0)
}

func (self *S) TestBootstrapPill(c *C) {
/*
	config := bootstrapConfiguration(basicConfig, self.Logger)
	c.Check(bootstrapSeven5(config, self.Logger),Not(Equals),"")
*/
}

