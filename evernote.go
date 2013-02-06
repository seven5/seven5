package seven5

import (
	"errors"
	"fmt"
	//oauth1 "github.com/iansmith/go-oauth/oauth"
	oauth1 "github.com/garyburd/go-oauth/oauth"
	"net/http"
	"net/url"
	"strconv"
)

const (
	EVERNOTE_AUTH_URL_HOST = "https://sandbox.evernote.com"
	EVERNOTE_AUTH_URL_PATH = "/oauth"
	EVERNOTE_AUTH_URL      = EVERNOTE_AUTH_URL_HOST + EVERNOTE_AUTH_URL_PATH
	EVERNOTE_USER_URL      = "https://sandbox.evernote.com/OAuth.action"
)

var evClient = oauth1.Client{
	TemporaryCredentialRequestURI: EVERNOTE_AUTH_URL,
	ResourceOwnerAuthorizationURI: EVERNOTE_USER_URL,
	TokenRequestURI:               EVERNOTE_AUTH_URL,
}

//EvernoteOauth1 is the implementation of an ServiceConnector for Evernote.  XXX
//the set of known credentials should be culled every few minutes because one
//could mount an attack against this growing without bounds. XXX
type EvernoteOauth1 struct {
	host       string
	knownCreds map[string]*oauth1.Credentials
}

//NewEvernoteOauth1 returns an OauthConnector suitable for use with Evernote.  
//The OauthClientDetail is passed here because we need extract
//client id and secret from somewhere other than the code.  The Deployment is passed to
//help calculate correct hostnames.
func NewEvernoteOauth1(d OauthClientDetail, dep DeploymentEnvironment) *EvernoteOauth1 {

	//app credentials are fixed
	evClient.Credentials.Token = d.ClientId("evernote")
	evClient.Credentials.Secret = d.ClientSecret("evernote")

	result := &EvernoteOauth1{
		host:       dep.RedirectHost("evernote"),
		knownCreds: make(map[string]*oauth1.Credentials),
	}
	return result
}

func (self *EvernoteOauth1) Phase2(token string, verifier string) (OauthConnection,error) {
	//should lookup creds in map
	creds, ok := self.knownCreds[token]
	if !ok {
		return nil,errors.New(fmt.Sprintf("Unable to find token %s in known credentials!", token))
	}
	cr, v, err := evClient.RequestToken(http.DefaultClient, creds, verifier)
	if err != nil {
		return nil,err
	}
	id:=v["edam_userId"][0]
	i,err:=strconv.ParseInt(id, 10, 64)
	if err!=nil {
		return nil,err
	}
	u:=v["edam_noteStoreUrl"][0]
	url, err:=url.Parse(u)
	if err!=nil {
		return nil,err
	}
	result:=&EvernoteConnection{
		Credentials:cr,
		EvernoteId:i,
		Notestore: url,
	}
	delete(self.knownCreds,token)
	return result,nil
}

func (self *EvernoteOauth1) Name() string {
	return "evernote"
}

func (self *EvernoteOauth1) UserInteractionURL(tempCred OauthCred, state string, callbackPath string) string {
	cred := &oauth1.Credentials{Token: tempCred.Token(), Secret: tempCred.Secret()}
	values := url.Values{
		"action": {state},
	}
	return evClient.AuthorizationURL(cred, values)
}

func (self *EvernoteOauth1) Phase1(state string, callbackPath string) (OauthCred, error) {
	callback := self.host + callbackPath
	tempCred, err := evClient.RequestTemporaryCredentials(http.DefaultClient, callback, nil)
	if err != nil {
		fmt.Printf("Error getting temp cred: %s\n", err.Error())
		return nil, err
	}
	self.knownCreds[tempCred.Token] = tempCred
	return &SimpleOauthCred{tempCred}, nil
}

func (self *EvernoteOauth1) StateValueName() string {
	return "action"
}

func (self *EvernoteOauth1) ClientTokenValueName() string {
	return "oauth_token"
}

func (self *EvernoteOauth1) CodeValueName() string {
	return "oauth_verifier"
}
func (self *EvernoteOauth1) ErrorValueName() string {
	return "oauth_error"
}

type EvernoteConnection struct {
	*oauth1.Credentials
	EvernoteId int64
	Notestore *url.URL
}

func (self *EvernoteConnection) SendAuthenticated(r *http.Request) (*http.Response,error) {
	return nil,nil
}
