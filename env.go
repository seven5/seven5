package seven5

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

//DeploymentEnvironment encodes information that cannot be obtained from the source code but can only
//be determined from "outside" the application.
type DeploymentEnvironment interface {
	//GetQbsStore returns the correct flavor of QbsStore for the deployment
	//environment, taking into account possible test settings.
	GetQbsStore() *QbsStore
	IsTest() bool
	Port() int
	//RedirectHost is needed in cases where you are using oauth because this must sent to the
	//"other side" of the handshake without any extra knowlege.
	RedirectHost(string) string
	GetAppValue(string) string
	ClientId(string) string
	MustAppValue(string) string
	GetValueOrPanic(n string) string
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
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(self.GetValueOrPanic(fmt.Sprintf("%s_%s", self.name, strings.ToUpper(n)))))
	}
}

//GetAppValue returns a value "inside" the application namespace of environment vars or APPNAME_KEY
//(key converted to upper case) and then fetched.  This value may be "".
func (self *EnvironmentVars) GetAppValue(key string) string {
	return os.Getenv(fmt.Sprintf("%s_%s", self.name, strings.ToUpper(key)))
}

//MustAppValue returns a value "inside" the application namespace of environment vars or APPNAME_KEY
//(key converted to upper case) and then fetched.  It panics if the value is not found.
func (self *EnvironmentVars) MustAppValue(key string) string {
	return self.GetValueOrPanic(fmt.Sprintf("%s_%s", self.name, strings.ToUpper(key)))
}

//ClientId returns the value of the client id that
//has been given out the by the service associated with name.  The environment variable
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

//GetValueOrPanic returns the environment variable based on the exact value supplied (it is not modified)
//and it panics if the value cannot be found.
func (self *EnvironmentVars) GetValueOrPanic(n string) string {
	value := os.Getenv(n)
	if value == "" {
		panic(fmt.Sprintf("expected to find environment variable %s but did not!", n))
	}
	return value
}

type LocalhostEnvironment struct {
	*EnvironmentVars
	test bool
}

func NewLocalhostEnvironment(appname string, test bool) *LocalhostEnvironment {

	result := &LocalhostEnvironment{EnvironmentVars: NewEnvironmentVars(appname), test: test}
	return result
}

func (self *LocalhostEnvironment) GetQbsStore() *QbsStore {
	if self.IsTest() {
		return NewQbsStore(self.MustAppValue("testdbname"), self.GetAppValue("testdriver"), self)
	}
	return NewQbsStore(self.MustAppValue("dbname"), self.GetAppValue("driver"), self)
}

func (self *LocalhostEnvironment) IsTest() bool {
	return self.test
}

func (self *LocalhostEnvironment) Port() int {
	portString := self.GetValueOrPanic("PORT")
	i, err := strconv.ParseInt(portString, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("PORT environment variable is not parseable: %s", err))
	}
	return int(i)
}

func (self *LocalhostEnvironment) RedirectHost(string) string {
	return fmt.Sprintf(REDIRECT_HOST_TEST, self.Port())
}
