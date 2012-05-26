package seven5

import (
	. "launchpad.net/gocheck"
	"testing"
	"encoding/json"
	"strings"
	"seven5/util"
	"seven5/plugins"
	"path/filepath"
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
	self.Logger = util.NewTerminalLogger(util.DEBUG, true)
}

const basicConfig =
`
{
	"Plugins" : {
		"ProjectValidator" : "plugins.DefaultProjectValidator"
	}
}
`

//but the VALUES must be case sensitive!
const caseInsensitiveConfig =
`
{
	"plugins" : {
		"projectvalidator" : "plugins.DefaultProjectValidator"
	}
}
`

const badConfig =
`
{
	"Plugins" : {
		"projectValidater" : "plugins.DefaultProjectValidator"
	}
}
`

func (self *S) TestReadJsonConfigWithValidator(c *C) {
	decoder := json.NewDecoder(strings.NewReader(basicConfig))
	var project ProjectConfig
	decoder.Decode(&project)
	c.Check(project.Plugins.ProjectValidator, Equals, "plugins.DefaultProjectValidator")
}

func (self *S) TestDefaultValidator(c *C) {
	var validator plugins.ProjectValidator = &plugins.DefaultProjectValidator{}
	testdata:=self.findTestDataPath("projectvalidator")
	c.Check(validator.Validate(filepath.Join(testdata,"testproj1"),self.Logger),
		Equals,true)		
	c.Check(validator.Validate(filepath.Join(testdata,"testproj2"),self.Logger),
		Equals,false)
		
	c.Check(validator.Validate(filepath.Join(testdata,"testproj3"),self.Logger),
		Equals,false)
}

func (self *S) findTestDataPath(insideTestData string) string {
	dirs:= strings.Split(os.Getenv("GOPATH"),string(filepath.ListSeparator))
	for _,candidate:= range(dirs) {
		if _,err:=os.Stat(filepath.Join(candidate,"testdata")); err==nil {
			return filepath.Join(candidate,"testdata", insideTestData)
		}
	}
	msg :=
`Your GOPATH environment variable does not include the Seven5 source code
tree so we cannot find the testdata and thus cannot run tests. Some element
of GOPATH should include testdata as its direct child.`	
	self.Logger.Panic(msg)
	return "" //wont happen
}

func (self *S) TestJsonParseAndCheck(c *C) {
	c.Check(bootstrapConfiguration(basicConfig, self.Logger),Not(IsNil))
	c.Check(bootstrapConfiguration(caseInsensitiveConfig, self.Logger),Not(IsNil))
	c.Check(bootstrapConfiguration(badConfig, self.Logger),IsNil)
	c.Check(bootstrapConfiguration("", self.Logger),IsNil)
}

func (self *S) TestBootstrapPill(c *C) {
	config := bootstrapConfiguration(basicConfig, self.Logger)
	c.Check(bootstrapSeven5(config, self.Logger),Not(Equals),"")
}

