package seven5

import (
	oauth2 "code.google.com/p/goauth2/oauth"
	_ "errors"
	"fmt"
	"net/http"
	"bytes"
	"encoding/json"
	_"strconv"
)

const (
	GITHUB_AUTH_URL_HOST = "https://github.com"
	GITHUB_AUTH_URL_PATH = "/login/oauth/authorize"
	GITHUB_AUTH_URL      = GITHUB_AUTH_URL_HOST + GITHUB_AUTH_URL_PATH
	GITHUB_TOKEN_URL     = "https://github.com/login/oauth/access_token"
	GITHUB_USER_SCOPE         = "user"
)

type GithubOauth2 struct {
	cfg  *oauth2.Config
	host string
}


func NewGithubOauth2(scope string, prompt string, d OauthClientDetail, dep DeploymentEnvironment) *GithubOauth2 {
	cfg := &oauth2.Config{
		ClientId:       d.ClientId("github"),
		ClientSecret:   d.ClientSecret("github"),
		Scope:          scope,
		AuthURL:        GITHUB_AUTH_URL,
		TokenURL:       GITHUB_TOKEN_URL,
		RedirectURL:    "", //accept defaults
		ApprovalPrompt: prompt,
	}
	return &GithubOauth2{
		host: dep.RedirectHost("github"),
		cfg:  cfg,
	}
}

func (self *GithubOauth2) CodeValueName() string {
	return "code"
}
func (self *GithubOauth2) ErrorValueName() string {
	return "error"
}
func (self *GithubOauth2) StateValueName() string {
	return "state"
}

//GithubUser represents the fields that you can extract about a user who uses oauth to log
//in via their github account.  To use this, your scope should include GITHUB_USER
type GithubUser struct {
	GithubId   string `json:"id"`
	EmailAddr  string `json:"email"`
	Name       string `json:"name"`
}


//Email makes it easier to implement user.Basic
func (self *GithubUser) Email() string {
	return self.EmailAddr
}

//SetEmail makes it easier to implement user.Basic
func (self *GithubUser) SetEmail(e string) {
	self.EmailAddr = e
}

//Returns the GithubUser object onces we have connected to the service.
func (self *GithubConnection) FetchUser() (*GithubUser, error) {
	return nil, nil
}


func (self *GithubOauth2) Name() string {
	return "github"
}

func (self *GithubOauth2) ClientTokenValueName() string {
	return "notused"
}

func (self *GithubOauth2) Phase1(state string, callbackPath string) (OauthCred, error) {
	return nil, nil
}

func (self *GithubOauth2) UserInteractionURL(ignored OauthCred, state string, callbackPath string) string {
	cb := fmt.Sprintf("%s%s", self.host, callbackPath)
	self.cfg.RedirectURL = cb
	return self.cfg.AuthCodeURL(state)
}


func (self *GithubOauth2) Phase2(ignore string, code string) (OauthConnection, error) {
	transport := &oauth2.Transport{
		Config: self.cfg,
	}
	_, err := transport.Exchange(code)
	if err != nil {
		return nil, err
	}
	
	return &GithubConnection{transport}, nil
}

type GithubConnection struct {
	*oauth2.Transport
}


func (self *GithubConnection) SendAuthenticated(r *http.Request) (*http.Response, error) {
	var buff bytes.Buffer
	enc:=json.NewEncoder(&buff)
	err:=enc.Encode(self.Transport);
	if err!=nil{
		panic(err)
	}
	return self.Client().Do(r)
}

