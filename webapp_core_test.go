package seven5

import (
	"encoding/json"
	"launchpad.net/gocheck"
	"strings"
	//"fmt"
)

//ex1 allows INPUT ONLY for B, but does not allow output
type ex1 struct {
	A string 
	B PrivateString
}

//ex2 shows how to not reveal things to the model, although it is still writable (B) and 
//to completely hide a field, even input is not allowed (A). note leading comma on
//omitempty
// B doesn't work: http://code.google.com/p/go/issues/detail?id=2761

type ex2 struct {
	A string `json:"-"`
	B PrivateString  `json:",omitempty"`
	
}

// This is the "suite" structure for objects that need to survive the whole of the tests
// in this file.
type JSONSuite struct {
}

var s = &JSONSuite{}

// hook up suite to gocheck
var _ = gocheck.Suite(s)

func (self *JSONSuite) TestMarshalOfPrivateString(c *gocheck.C) {
	example1 := &ex1{A: "foo", B: "bar"}
	jsonContent, err := json.Marshal(&example1)
	result:=string(jsonContent)
	
	c.Assert(err, gocheck.Equals, nil)
	c.Check(strings.Index(result,"bar"),  gocheck.Equals, -1)
	//sanity check
	c.Check(strings.Index(result,"foo"),  gocheck.Not(gocheck.Equals), -1)
}

func (self *JSONSuite) TestUnMarshalOfPrivateString(c *gocheck.C) {
	var result ex1
	jsonContent :=[]byte(`{"A":"frip","B":"baz"}`)
	err:=json.Unmarshal(jsonContent, &result)
	
	c.Assert(err, gocheck.Equals, nil)
	c.Check(string(result.B),  gocheck.Equals, "baz")
	//sanity check
	c.Check(result.A, gocheck.Equals, "frip")
}

//
// http://code.google.com/p/go/issues/detail?id=2761
//
//
//It's too late to be changing the json interface right now.
//Omitempty has a clear definition.  Marshalers do not
//get to change that definition.
//
//I would suggest updating the documentation but I don't
//think there's anything to say.
//
//Russ
func (self *JSONSuite) xTestMarshalOfJsonOptions(c *gocheck.C) {
	example2 := &ex2{A: "fleazil", B: "grix"}
	jsonContent, err := json.Marshal(&example2)
	result:=string(jsonContent)
	
	c.Assert(err, gocheck.Equals, nil)
	c.Check(result,  gocheck.Equals, "{}")
}
