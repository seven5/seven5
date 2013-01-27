package seven5

import (
	"os"
	"fmt"
	"strings"
	"path/filepath"
	"net/http"
)

//FileFlavor is used to "find" parts of your application at run-time. Use of these constants
//means that if the default project layout changes you only have to move your files around,
//not change your code.
type FileFlavor int

const (
	GO_SOURCE_FLAVOR	= iota
	DART_FLAVOR
	ASSET_FLAVOR
	TOP_LEVEL_FLAVOR
)

//DeploymentEnvironment encodes information that cannot be obtained from the source code but can only
//be determined from "outside" the application.
type DeploymentEnvironment interface {
	IsTest() bool
	Port() int
	//RedirectHost is needed in cases where you are using oauth because this must sent to the 
	//"other side" of the handshake without any extra knowlege.
	RedirectHost(string) string
}

//PublicSettings is an interface representing information that you want the client to have
//access to, usually via a URL, but do not want stored in the source code.  A common example
//of this is an API key that you use with a particular web service that needs to be available
//to the client side browser, but should be checked in to source code revision.  Note that
//you can use the returned value of PublicSettingHandler to map particular public settings
//into the URL space for easy access by a client.
type PublicSettings interface {
	PublicSettingsHandler(n string) func(http.ResponseWriter, *http.Request)
}

//ProjectFinder is an abstraction of the typical seven5 project layout. It knows how to 
//find objects of different types in their standard locations within a project.
type ProjectFinder interface {
	ProjectFind(target string, projectName string, flavor FileFlavor) (string, error)
}

//EnvironmentVars is ProjectFinder, OauthClientDetail, and PublicSettings handler
//implementation that reads values from a standard arrangement of unix-ish environment variables.  
//Typically the enivornment variables are prefixed with the application name and that must be
//provided to NewEnviromentVars. 
type EnvironmentVars struct {
	name string
}

//NewEnvironmentVars returns an initialied EnvironmentVars based on the name provided.
func NewEnvironmentVars(appName string) *EnvironmentVars {
	return &EnvironmentVars{
		name: strings.ToUpper(appName),
	}
}

//PublicSettingHandler returns a function suitable for insertion into an http or seven5 
//ServeMux as a handler for a particular URL.  It only calculates the value of the
//result once, at the time it returns the function here.  It calls GetValueOrPanic to compute
//it's result.
func (self *EnvironmentVars) PublicSettingHandler(n string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r*http.Request) {
		w.Write([]byte(self.GetValueOrPanic(fmt.Sprintf("%s_%s", self.name, strings.ToUpper(n)))))
	}
}

//GetAppValue returns a value "inside" the application namespace of environment vars or APPNAME_KEY 
//(key converted to upper case) and then fetched.  This value may be "". 
func (self *EnvironmentVars) GetAppValue(key string) string {
	return os.Getenv(fmt.Sprintf("%s_%s", self.name, strings.ToUpper(key)))
}

//ClientId returns the value of the client id that
//has been given out the by the service associated with service.  The environment variable
//is APPNAME_SERVICENAME_CLIENT_ID and is read only once.
func (self *EnvironmentVars) ClientId(name string) string {
	return self.GetValueOrPanic(fmt.Sprintf("%s_%s_CLIENT_ID", self.name, strings.ToUpper(name)))
}

//ClientSecret returns the value of the environment value of the client secret that
//has been given out the by the service associated with service.  The environment variable
//is APPNAME_SERVICENAME_CLIENT_SECRET and is read only once.
func (self *EnvironmentVars) ClientSecret(name string) string {
	return self.GetValueOrPanic(fmt.Sprintf("%s_%s_CLIENT_SECRET", self.name, strings.ToUpper(name)))
}

//ProjectFind defaults to looking at the GOPATH environment variable to work out the location
//of other objects in the project.  It calls GetValueOrPanic("GOPATH").
func (self *EnvironmentVars) ProjectFind(target string, projectName string, flavor FileFlavor) (string, error) {
	env := self.GetValueOrPanic("GOPATH")
	pieces := strings.Split(env, ":")
	if len(pieces) > 1 {
		env = pieces[0]
	}
	env = filepath.Clean(env)
	if strings.HasSuffix(env, "/") && env != "/" {
		env = env[0 : len(env)-1]
	}
	switch flavor {
	case GO_SOURCE_FLAVOR:
		return filepath.Join(env, "src", target), nil
	case DART_FLAVOR:
		return filepath.Join(filepath.Dir(env), "dart", projectName, target), nil
	case ASSET_FLAVOR:
		return filepath.Join(filepath.Dir(env), "dart", projectName, "assets", target), nil
	case TOP_LEVEL_FLAVOR:
		return filepath.Join(filepath.Dir(env), target), nil
	}
	panic("unknown type of object searched for in the project!")
}


//GetValueOrPanic returns the environment variable based on the exact value supplied (it is not modified)
//and it panics if the value cannot be found.
func (self *EnvironmentVars) GetValueOrPanic(n string) string {
	value := os.Getenv(n)
	if value == "" {
		panic(fmt.Sprintf("expected to find environment variable %s but did not!", n))
	}
	return value
}
