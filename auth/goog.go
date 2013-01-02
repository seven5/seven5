package auth

import (
	"code.google.com/p/goauth2/oauth"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
)



const (
	GOOGLE_AUTH_URL_HOST = "https://accounts.google.com"
	GOOGLE_AUTH_URL_PATH = "/o/oauth2/auth"
	GOOGLE_AUTH_URL		= GOOGLE_AUTH_URL_HOST + GOOGLE_AUTH_URL_PATH
	GOOGLE_TOKEN_URL	= "https://accounts.google.com/o/oauth2/token"
	GOOGLE_USER_INFO	= "https://www.googleapis.com/oauth2/v1/userinfo"
)

//GoogleAuthConn is the implementation of an ServiceConnector for Google.
type GoogleAuthConn struct {
	scope		string
	prompt		string
	clientId	string
	clientSecret	string
	host		string
}

//NewGoogleAuthConnector returns an ServiceConnector suitable for use with Google.  Note
//that the scope variable is very important to google and should be set based on the needs
//of your app.  The prompt values can be "auto" or "force" to force a re-prompt from google
//on each authentication handshake. The OauthClientDetail is passed here because we need extract
//client id and secret from somewhere other than the code.  The Deployment is passed to
//help calculate correct hostnames.
func NewGoogleAuthConnector(scope string, prompt string, d OauthClientDetail, dep DeploymentEnvironment) ServiceConnector {
	result := &GoogleAuthConn{scope: scope, prompt: prompt}
	result.clientId = d.ClientId(result)
	result.clientSecret = d.ClientSecret(result)
	result.host = dep.RedirectHost(result)
	return result
}

//NewGoogleAuth simplifies the interaction with a google auth connector to the call just providing
//the google scope for the connection and the application.  This assumes you are using Heroku for any 
//delpoyment; if you are not doing a deployment, and only running in test mode, the heroku url will not
//be used.  This assumes that your google client and id and secret are in the normal place 
//(such as APPNAME_GOOGLE_CLIENT_ID or ...CLIENT_SECRET) and that
//APPNAME_TEST is not-empty for test mode.  The test port is the default Seven5 port for testing.
func NewGoogleAuth(scope string, appname string) ServiceConnector {
	env:=NewEnvironmentVars(appname)
	rem:=NewRemoteDeployment(appname, DEFAULT_TEST_PORT)	
	return NewGoogleAuthConnector(scope, "auto", env, rem)
}

//OauthConfig creates the config structure needed by the code.google.com/p/goauth2 library.
func (self *GoogleAuthConn) OauthConfig(callbackpath string) *oauth.Config {
	return &oauth.Config{
		ClientId:	self.clientId,
		ClientSecret:	self.clientSecret,
		Scope:		self.scope,
		AuthURL:	GOOGLE_AUTH_URL,
		TokenURL:	GOOGLE_TOKEN_URL,
		RedirectURL:	fmt.Sprintf("%s%s", self.host, callbackpath),
		ApprovalPrompt:	self.prompt,
	}
}

//ExchangeForToken returns an oauth.Token structure from a code received in the handshake
//plus the basic information in the AuthPageMapper.  This will be called after the
//first phase of the oauth exchange is done and we want to exchange a code for a token.
func (self *GoogleAuthConn) ExchangeForToken(path string, code string) (*oauth.Transport, error) {
	config := self.OauthConfig(path)
	t := &oauth.Transport{Config: config}
	_, err := t.Exchange(code)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (self *GoogleAuthConn) Name() string {
	return "google"
}

func (self *GoogleAuthConn) AuthURL(path string, state string) string {
	return self.OauthConfig(path).AuthCodeURL(state)
}

func (self *GoogleAuthConn) CodeValueName() string {
	return "code"
}
func (self *GoogleAuthConn) ErrorValueName() string {
	return "error"
}
func (self *GoogleAuthConn) StateValueName() string {
	return "state"
}

//GoogleUser represents the fields that you can extract about a user who uses oauth to log
//in via their gmail/google account.
type GoogleUser struct {
	GoogleId	string	`json:"id"`
	Email		string	`json:"email"`
	Name		string	`json:"name"`
	GivenName	string	`json:"given_name"`
	FamilyName	string	`json:"family_name"`
	Link		string	`json:"link"`
	Picture		string	`json:"picture"`
	Gender		string	`json:"gender"`
	Locale		string	`json:"locale"`
	Birthday	string	`json:"birthday"`
}

func (self *GoogleUser) Fetch(transport *oauth.Transport) (*GoogleUser, error) {
	r, err := transport.Client().Get(GOOGLE_USER_INFO)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	s := string(body)
	decoder := json.NewDecoder(strings.NewReader(s))
	result := &GoogleUser{}
	err = decoder.Decode(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
