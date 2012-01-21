package seven5

import (
	"github.com/seven5/gozmq"
	"launchpad.net/gocheck"
	"net/http"
	"path/filepath"
	//"fmt"
	//"os"
)
//PrepareFunctionalTest configures mongrel to run the RawHandler supplied.
//Return value is nil on error, and that's a fatal error that has already
//been logged.
func PrepareFunctionalTest(n Routable, c *gocheck.C) gozmq.Context {
	//using . because 'gb -t' changes the cwd to the right place
	path, err := filepath.Abs(".")
	if err != nil {
		c.Fatalf("unable to start up, can't find absolute path to cwd!")
	}
	path = filepath.Clean(path)
	config := bootstrapFromDir(path)

	if config == nil {
		c.Fatalf("unable to bootstrap project from path %s", path)
	}

	if err = WriteTestConfig(config, n); err != nil {
		c.Fatalf("unable to write mongrel2 test configuration: %s", err)
	}

	var ctx gozmq.Context

	if ctx, err = createNetworkResources(config); err != nil {
		c.Fatalf("unable to create mongrel2 or 0MQ resources:", err)
	}

	if !startUp(ctx, nil, config, []Routable{n}) {
		c.Fatalf("unable to start the handlers, no context found: %s, %s", n.Name(), path)
	}
	return ctx
}

//WriteTestConfig does the necessary steps to write a mongrel 2 configuration suitable for
//testing.
func WriteTestConfig(config *projectConfig, n Routable) error {
	var err error
	//this accepts all the defaults for log placement, pid files, etc.
	if err = generateServerHostConfig(config, localhost, test_port); err != nil {
		return err
	}
	//only can do HTTP testing right now
	if err = generateHandlerAddressAndRouteConfig(config, localhost, n); err != nil {
		return err
	}
	//static content at /static
	if err = generateStaticContentConfig(config, localhost, static_dir); err != nil {
		return err
	}

	//normally this does nothing unless the DB is completely empty
	if err = generateMimeTypeConfig(config); err != nil {
		return err
	}

	if err = finishConfig(config); err != nil {
		return err
	}

	return nil
}

//This function performs a mapping test on the URL provided.  It checks that if
//mongrel2 sees the url, it invokes the handler you supplied.  (It actually just checks
//the name, not the pointer.) If it fails, it notes the failure in the provided gocheck.C
func MappingTest(url string, h Httpified, c *gocheck.C) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.Fatalf("failed to create request for test!")
	}

	req.Header.Add(route_test_header, h.Name())

	client := new(http.Client)
	resp, err := client.Do(req)

	if err != nil {
		c.Fatalf("failed trying to use http to localhost:6767: %s", err)
	}

	if len(resp.Header[route_test_header]) == 0 {
		c.Fatalf("failed trying to find response header: %s", url)
	}
	c.Check(resp.StatusCode, gocheck.Equals, route_test_response_code)
	c.Check(resp.Header[route_test_header][0], gocheck.Equals, h.Name())
}
