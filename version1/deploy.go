package seven5

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	REDIRECT_HOST_TEST = "http://localhost:%d"
	HEROKU_HOST        = "https://%s.herokuapp.com"
)

//HerokuDeploy is an implementation of DeploymentEnvironment that understands
//about Heroku and reads its configuration from environment variables.
type HerokuDeploy struct {
	herokuName string
	appName    string
}

//NewHerokuDeploy returns a new HerokuDeploy object that implements DeploymentEnvironment.
//The first parameter is the _host_ name on heroku, typically like damp-sierra-7161.
//The second parameter is the application's name locally, like myproject.
func NewHerokuDeploy(herokuName string, appName string) *HerokuDeploy {
	result := &HerokuDeploy{
		herokuName: herokuName,
		appName:    appName,
	}
	return result
}

//GetQbsStore returns a Qbs store suitable for use with tihs application. The
//implementation uses GetDSNOrDie which ends up looking for the environment
//variable DATABASE_URL which must be set or we panic.
func (self *HerokuDeploy) GetQbsStore() *QbsStore {
	dsn := GetDSNOrDie()
	return NewQbsStoreFromDSN(dsn)
}

//IsTest returns true if the environment variable localname_TEST is set to
//a value that's not "".
func (self *HerokuDeploy) IsTest() bool {
	t := os.Getenv(strings.ToUpper(self.appName) + "_TEST")
	return t != ""
}

//Port reads the value of the environment variable PORT to get the value to return here.  It
//will panic if the environment variable is not set or it's not a number.
func (self *HerokuDeploy) Port() int {
	p := os.Getenv("PORT")
	if p == "" {
		panic("PORT not defined")
	}
	i, err := strconv.Atoi(p)
	if err != nil {
		panic(err)
	}
	return i
}

//RedirectHost is needed in cases where you are using oauth because this must sent to the
//"other side" of the handshake without any extra knowlege.
func (self *HerokuDeploy) RedirectHost() string {
	if self.IsTest() {
		return fmt.Sprintf(REDIRECT_HOST_TEST, self.Port())
	}
	return fmt.Sprintf(HEROKU_HOST, self.herokuName)
}

// Url returns the string that points to the application itself.  Note that
// this will not have a / on the end.
func (self *HerokuDeploy) Url() string {
	return "http://" + self.RedirectHost()
}

//DeploymentEnvironment encodes information that cannot be obtained from the source code but can only
//be determined from "outside" the application.  This is here to provide a
//a vague hope of supporting deployments that are not based on environment
//variables.
type DeploymentEnvironment interface {
	//GetQbsStore returns the correct flavor of QbsStore for the deployment
	//environment, taking into account possible test settings.
	GetQbsStore() *QbsStore
	//IsTest returns true if this application is running on
	//a local system, not deployed to remote system.
	IsTest() bool
	//Returns the port number for this application.
	Port() int
	//RedirectHost is needed in cases where you are using oauth because this must sent to the
	//"other side" of the handshake without any extra knowlege.
	RedirectHost() string
}
