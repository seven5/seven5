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
func PrepareFunctionalTest(n Named, c *gocheck.C) gozmq.Context {
	//using . because 'gb -t' changes the cwd to the right place
	path, err := filepath.Abs(".")
	if err != nil {
		c.Fatalf("unable to start up, can't find absolute path to cwd!")
	}
	path = filepath.Clean(path)
	config:= BootstrapFromDir(path)
	
	if config==nil {
		c.Fatalf("unable to bootstrap project from path %s",path)
	}
	
	if err=WriteTestConfig(config,n); err!=nil {
		c.Fatalf("unable to write mongrel2 test configuration: %s",err)	
	}
	
	var ctx gozmq.Context
	
	if ctx,err=CreateNetworkResources(config); err!=nil {
		c.Fatalf("unable to create mongrel2 or 0MQ resources:",err)	
	}
	
	if !StartUp(ctx,config,[]Named{n}) {
		c.Fatalf("unable to start the handlers, no context found: %s, %s",n.Name(),path)
	} 
	return ctx	
}

//WriteTestConfig does the necessary steps to write a mongrel 2 configuration suitable for
//testing.
func WriteTestConfig(config *ProjectConfig, n Named) error {
	var err error
	//this accepts all the defaults for log placement, pid files, etc.
	if err = GenerateServerHostConfig(config, LOCALHOST, TEST_PORT); err!=nil {
		return err
	}
	//only can do HTTP testing right now
	if err= GenerateHandlerAddressAndRouteConfig(config, LOCALHOST, n); err!=nil{
		return err
	}
	//static content at /static
	if err=GenerateStaticContentConfig(config, LOCALHOST, STATIC); err!=nil {
		return err
	}

	//normally this does nothing unless the DB is completely empty
	if err=GenerateMimeTypeConfig(config); err!=nil {
		return err
	}
	
	if err=FinishConfig(config); err!=nil {
		return err
	}
	
	return nil
}

//This function performs a mapping test on the URL provided.  It checks that if
//mongrel2 sees the url, it invokes the handler you supplied.  (It actually just checks
//the name, not the pointer.) If it fails, it notes the failure in the provided gocheck.C
func MappingTest(url string, h Httpified,c *gocheck.C) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.Fatalf("failed to create request for test!")
	}
	req.Header.Add(ROUTE_TEST_HEADER, h.Name())

	client := new(http.Client)
	resp, err := client.Do(req)

	if err != nil {
		c.Fatalf("failed trying to use http to localhost:6767: %s", err)
	}
	c.Check(resp.StatusCode, gocheck.Equals, ROUTE_TEST_RESPONSE_CODE)
	c.Check(resp.Header[ROUTE_TEST_HEADER][0], gocheck.Equals, h.Name())
}
