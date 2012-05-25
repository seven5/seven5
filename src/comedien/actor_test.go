package comedien

import (
	. "launchpad.net/gocheck"
	"testing"
	"net/http"
)

// Hook up gocheck into the gotest runner.
func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

//Cant' run nil as a message
func (s *S) TestSupervisorsDontTakeNil(c *C) {
	actor := &HttpRcvActor{};
	sup := Supervise(actor)
	c.Check(false, Equals, sup.ProcessOne(nil))
}

type bogus struct {
}

//Cant' run a wrong type
func (s *S) TestSupervisorsWontTakeWrongType(c *C) {
	actor := &HttpRcvActor{};
	sup := Supervise(actor)
	c.Check(false, Equals, sup.ProcessOne(&bogus{}))
}

//Can run an HTTP request
func (s *S) TestHTTPActorBasic(c *C) {
	actor := &HttpRcvActor{}
	sup := Supervise(actor)
	req := &http.Request{}
	
	req.RequestURI="/foo/bar"
	req.Method="GET"
	
	c.Check(true, Equals, sup.ProcessOne(req))
}
