package seven5

import (
	"fmt"
)

const (
	REDIRECT_HOST_TEST = "http://localhost:3003"
	HEROKU_HOST        = "https://%s.herokuapp.com"
)

//Deployment represents information about a particular deployment that may be needed for
//calculating URLs.
type Deployment interface {
	IsTest() bool
	RedirectHost(AuthServiceConnector) string
}

//RemoteDeployment is a very simple implementation of Deployment that knows the beginning of a
//remote url and uses the environment for knowing if we are running in test mode or not.
type RemoteDeployment struct {
	url string
	env Environment
}

//NewRemoteDeployment returns an implementation of Deployment that points URLs to the fullUrl
//provided unless in test mode, in which case they point to localhost.
func NewRemoteDeployment(fullUrl string, env Environment) Deployment {
	return &RemoteDeployment{
		url: fullUrl,
		env: env,
	}
}

//HerokuName calculates the full url of a deployed Heroku app. Useful in conjunction with
//NewRemoteDeployment.
func HerokuName(n string) string {
	return fmt.Sprintf(HEROKU_HOST, n)
}

func (self *RemoteDeployment)	IsTest() bool {
	return self.env.IsTest()
}

func (self *RemoteDeployment)	RedirectHost(AuthServiceConnector) string {
	if self.IsTest() {
		return REDIRECT_HOST_TEST
	} 
	return self.url
}
