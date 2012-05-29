package seven5

import (
	"fmt"
	. "launchpad.net/gocheck"
	"os"
	"seven5/groupie"
	"seven5/util"
	"strings"
	"testing"
)

// Hook up gocheck into the gotest runner.
func Test(t *testing.T) { TestingT(t) }

//suite of tests
type S struct {
	logger util.SimpleLogger
}

var _ = Suite(&S{})

func (self *S) SetUpSuite(c *C) {
	self.logger = util.NewTerminalLogger(os.Stdout, util.DEBUG, true)
}

func (self *S) TestGroupieConfig(c *C) {
	t1 := util.ReadTestData(c, "groupieconfig", "test1.json")
	ur := util.ReadTestData(c, "groupieconfig", "test-unknown-role.json")

	roleName := groupie.VALIDATEPROJECT

	conf, err := parseGroupieConfig(t1)
	c.Assert(err, IsNil)
	c.Assert(len(conf), Not(Equals), 0)
	c.Assert(conf[roleName], Not(IsNil))
	c.Check(strings.Index(conf[roleName].TypeName,"DefaultValidateProject"),
		 Not(Equals), -1)
	c.Check(len(conf[roleName].ImportsNeeded), Not(Equals), 0)

	_, err = parseGroupieConfig(ur)
	c.Assert(err, Not(IsNil))

}

func (self *S) TestBootstrapPill(c *C) {
	b := &bootstrap{logger: self.logger, request: nil}
	checkNil := []bool{false, true, true}

	for i, check := range checkNil {
		dir := util.FindTestDataPath(c, "groupieconfig",
			fmt.Sprintf("bootstraptest%d", i+1))
		conf := b.configureSeven5(dir)
		if check {
			c.Check(conf, IsNil)
		} else {
			c.Check(conf, Not(IsNil))
			c.Check(b.takeSeven5Pill(conf), Not(Equals), "")
		}
	}
}
