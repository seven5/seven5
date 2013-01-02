package auth

import (
	"fmt"
)

const (
	REDIRECT_HOST_TEST	= "http://localhost:%d"
	HEROKU_HOST		= "https://%s.herokuapp.com"
	DEFAULT_TEST_PORT = 3003
)

//RemoteDeployment is a very simple implementation of DeploymentEnvironment that knows the beginning of a
//remote url and uses the environment for knowing if we are running in test mode or not.
type RemoteDeployment struct {
	url	string
	isTest	bool
	testPort int
}

//NewRemoteDeployment returns an implementation of Deployment that points URLs to the fullUrl
//provided unless in test mode.  If testPort>0 then we are assumed to be in test mode.
func NewRemoteDeployment(fullUrl string, testPort int) DeploymentEnvironment {
	return &RemoteDeployment{
		url:	fullUrl,
		isTest:	testPort>0,
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
func (self *RemoteDeployment) TestPort() int {
	return self.testPort
}

func (self *RemoteDeployment) RedirectHost(ServiceConnector) string {
	if self.IsTest() {
		return fmt.Sprintf(REDIRECT_HOST_TEST, self.testPort)
	}
	return self.url
}
