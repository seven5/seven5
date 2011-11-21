package seven5

import (
	"path/filepath"
	"github.com/alecthomas/gozmq"
	"launchpad.net/gocheck"
	"net/http"
)
//PrepareFunctionalTest configures mongrel to run the RawHandler supplied.
//Return value is nil on error, and that's a fatal error that has already
//been logged.
func PrepareFunctionalTest(rh RawHandler, c *gocheck.C) gozmq.Context {
	//using . because 'gb -t' changes the cwd to the right place
	path, err := filepath.Abs(".")
	if err != nil {
		c.Fatalf("unable to start up, can't find absolute path to cwd!")
	}
	path = filepath.Clean(path)
	context := StartUp([]RawHandler{rh}, path)
	if context == nil {
		c.Fatalf("unable to start the handlers, no context found")
	} 
	return context	
}

//This function performs a mapping test on the URL provided.  It checks that if
//mongrel2 sees the url, it invokes the handler you supplied.  (It actually just checks
//the name, not the pointer.) If it fails, it notes the failure in the provided gocheck.C
func MappingTest(url string, rh RawHandler,c *gocheck.C) {
	name := "echo"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.Fatalf("failed to create request for test!")
	}
	req.Header.Add(ROUTE_TEST_HEADER, rh.Name())

	client := new(http.Client)
	resp, err := client.Do(req)

	if err != nil {
		c.Fatalf("failed trying to use http to localhost:6767: %s", err)
	}
	c.Check(resp.StatusCode, gocheck.Equals, ROUTE_TEST_RESPONSE_CODE)
	c.Check(resp.Header[ROUTE_TEST_HEADER][0], gocheck.Equals, name)
}
