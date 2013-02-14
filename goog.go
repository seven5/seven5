package seven5

import (
	oauth2 "code.google.com/p/goauth2/oauth"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"bytes"
)

const (
	GOOGLE_AUTH_URL_HOST = "https://accounts.google.com"
	GOOGLE_AUTH_URL_PATH = "/o/oauth2/auth"
	GOOGLE_AUTH_URL      = GOOGLE_AUTH_URL_HOST + GOOGLE_AUTH_URL_PATH
	GOOGLE_TOKEN_URL     = "https://accounts.google.com/o/oauth2/token"
	GOOGLE_USER_INFO     = "https://www.googleapis.com/oauth2/v1/userinfo"
)

type GoogleOauth2 struct {
	cfg  *oauth2.Config
	host string
}

func NewGoogleOauth2(scope string, prompt string, d OauthClientDetail, dep DeploymentEnvironment) *GoogleOauth2 {
	cfg := &oauth2.Config{
		ClientId:       d.ClientId("google"),
		ClientSecret:   d.ClientSecret("google"),
		Scope:          scope,
		AuthURL:        GOOGLE_AUTH_URL,
		TokenURL:       GOOGLE_TOKEN_URL,
		RedirectURL:    "", //don't know it yet
		ApprovalPrompt: prompt,
	}
	return &GoogleOauth2{
		host: dep.RedirectHost("google"),
		cfg:  cfg,
	}
}

func (self *GoogleOauth2) CodeValueName() string {
	return "code"
}
func (self *GoogleOauth2) ErrorValueName() string {
	return "error"
}
func (self *GoogleOauth2) StateValueName() string {
	return "state"
}

//GoogleUser represents the fields that you can extract about a user who uses oauth to log
//in via their gmail/google account.  To use this, your scope must include Userinfo.
type GoogleUser struct {
	GoogleId   string `json:"id"`
	EmailAddr  string `json:"email"`
	Name       string `json:"name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Link       string `json:"link"`
	Picture    string `json:"picture"`
	Gender     string `json:"gender"`
	Locale     string `json:"locale"`
	Birthday   string `json:"birthday"`
}

//Email makes it easier to implement user.Basic
func (self *GoogleUser) Email() string {
	return self.EmailAddr
}

//SetEmail makes it easier to implement user.Basic
func (self *GoogleUser) SetEmail(e string) {
	self.EmailAddr = e
}

//Returns the GoogleUser object onces we have connected to the service.
func (self *GoogleConnection) FetchUser() (*GoogleUser, error) {
	r, err := self.Client().Get(GOOGLE_USER_INFO)
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

func (self *GoogleOauth2) Name() string {
	return "google"
}

func (self *GoogleOauth2) ClientTokenValueName() string {
	return "notused"
}

func (self *GoogleOauth2) Phase1(state string, callbackPath string) (OauthCred, error) {
	return nil, nil
}

func (self *GoogleOauth2) UserInteractionURL(ignored OauthCred, state string, callbackPath string) string {
	cb := fmt.Sprintf("%s%s", self.host, callbackPath)
	self.cfg.RedirectURL = cb
	return self.cfg.AuthCodeURL(state)
}

func (self *GoogleOauth2) Phase2(ignore string, code string) (OauthConnection, error) {
	transport := &oauth2.Transport{
		Config: self.cfg,
	}
	_, err := transport.Exchange(code)
	if err != nil {
		return nil, err
	}
	
	return &GoogleConnection{transport}, nil
}

type GoogleConnection struct {
	*oauth2.Transport
}

func (self *GoogleConnection) SendAuthenticated(r *http.Request) (*http.Response, error) {
	var buff bytes.Buffer
	enc:=json.NewEncoder(&buff)
	err:=enc.Encode(self.Transport);
	if err!=nil{
		panic(err)
	}
	return self.Client().Do(r)
}
