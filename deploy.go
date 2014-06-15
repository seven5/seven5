package seven5

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

const (
	REDIRECT_HOST_TEST = "http://localhost:%d"
	HEROKU_HOST        = "https://%s.herokuapp.com"
	DEFAULT_TEST_PORT  = 3003
)

//RemoteDeployment is a very simple implementation of DeploymentEnvironment that knows the beginning of a
//remote url and uses the environment for knowing if we are running in test mode or not.
type RemoteDeployment struct {
	url      string
	isTest   bool
	testPort int
}

type HerokuDeploy struct {
	*EnvironmentVars
	name string
}

//NewHerokuDeploy returns a new HerokuDeploy object that implements DeploymentEnvironment.
func NewHerokuDeploy(herokuName string, localName string) *HerokuDeploy {
	result := &HerokuDeploy{
		EnvironmentVars: NewEnvironmentVars(localName),
		name:            herokuName,
	}
	return result
}

func (self *HerokuDeploy) GetQbsStore() *QbsStore {
	if self.IsTest() {
		return NewQbsStore(self.MustAppValue("TESTDBNAME"),
			self.MustAppValue("TESTDRIVER"), self)
	}
	db := os.Getenv("DATABASE_URL")
	if db == "" {
		panic("no DATABASE_URL found, cannot connect to a *QbsStore")
	}
	log.Printf("found a database:%s", db)
	pair := strings.SplitN(db, ":", 2)
	if len(pair) != 2 {
		panic(fmt.Sprintf("DATABASE_URL does not start with driver: %v", db))
	}
	log.Printf("pair %v %v", pair[0], pair[1])
	return NewQbsStore(pair[1], pair[0], self)
}

func (self *HerokuDeploy) IsTest() bool {
	t := self.GetAppValue("TEST")
	return t != ""
}

//Port reads the value of the environment variable PORT to get the value to return here.  It
//will panic if the environment variable is not set or it's not a number.
func (self *HerokuDeploy) Port() int {
	p := self.EnvironmentVars.GetValueOrPanic("PORT")
	i, err := strconv.Atoi(p)
	if err != nil {
		panic(err)
	}
	return i
}

//RedirectHost is needed in cases where you are using oauth because this must sent to the
//"other side" of the handshake without any extra knowlege.
func (self *HerokuDeploy) RedirectHost(string) string {
	if self.IsTest() {
		return fmt.Sprintf(REDIRECT_HOST_TEST, self.Port())
	}
	return HerokuName(self.name)

}

//NewRemoteDeployment returns an implementation of Deployment that points URLs to the fullUrl
//provided unless in test mode.  If testPort>0 then we are assumed to be in test mode.
func NewRemoteDeployment(fullUrl string, testPort int) *RemoteDeployment {
	return &RemoteDeployment{
		url:      fullUrl,
		isTest:   testPort > 1023,
		testPort: testPort,
	}
}

//HerokuName calculates the full url of a deployed Heroku app. Useful in conjunction with
//NewRemoteDeployment.
func HerokuName(n string) string {
	return fmt.Sprintf(HEROKU_HOST, n)
}

func (self *RemoteDeployment) IsTest() bool {
	return self.isTest
}
func (self *RemoteDeployment) Port() int {
	return self.testPort
}

func (self *RemoteDeployment) RedirectHost(string) string {
	if self.IsTest() {
		return fmt.Sprintf(REDIRECT_HOST_TEST, self.Port())
	}
	return self.url
}
