package auth

import (
	"code.google.com/p/goauth2/oauth"
	"fmt"
	oauth1 "github.com/garyburd/go-oauth/oauth"
	"net/http"
)

const (
	EVERNOTE_AUTH_URL_HOST = "https://sandbox.evernote.com"
	EVERNOTE_AUTH_URL_PATH = "/oauth"
	EVERNOTE_AUTH_URL      = EVERNOTE_AUTH_URL_HOST + EVERNOTE_AUTH_URL_PATH
	EVERNOTE_TOKEN_URL     = EVERNOTE_AUTH_URL
)

//EvernoteAuthConn is the implementation of an ServiceConnector for Evernote.
type EvernoteAuthConn struct {
	host   string
	client oauth1.Client
}

//NewEvernoteConnector returns an ServiceConnector suitable for use with Evernote.  
//The OauthClientDetail is passed here because we need extract
//client id and secret from somewhere other than the code.  The Deployment is passed to
//help calculate correct hostnames.
func NewEvernoteAuthConnector(d OauthClientDetail, dep DeploymentEnvironment) ServiceConnector {
	result := &EvernoteAuthConn{}
	result.host = dep.RedirectHost(result)
	result.client = oauth1.Client{
		TemporaryCredentialRequestURI: "https://sandbox.evernote.com/oauth",
		ResourceOwnerAuthorizationURI: "https://sandbox.evernote.com/OAuth.action",
		TokenRequestURI:               "https://sandbox.evernote.com/oauth",
		Credentials: oauth1.Credentials{
			Token:  d.ClientId(result),
			Secret: d.ClientSecret(result)},
	}
	return result
}

//ExchangeForToken returns an oauth.Token structure from a code received in the handshake
//plus the basic information in the AuthPageMapper.  This will be called after the
//first phase of the oauth exchange is done and we want to exchange a code for a token.
func (self *EvernoteAuthConn) ExchangeForToken(path string, code string) (*oauth.Transport, error) {
	return nil, nil
}

func (self *EvernoteAuthConn) Name() string {
	return "evernote"
}

func (self *EvernoteAuthConn) AuthURL(path string, state string) string {
	callback := self.host + "/auth/evernote/oauth2callback"
	tempCred, err := self.client.RequestTemporaryCredentials(http.DefaultClient, callback, nil)
	if err != nil {
		fmt.Printf("Error getting temp cred: %s\n", err.Error())
		return ""
	}
	fmt.Printf("I've got temp cred: %+v\n",tempCred)
	return self.client.AuthorizationURL(tempCred, nil)
}

func (self *EvernoteAuthConn) CodeValueName() string {
	return "oauth_token"
}
func (self *EvernoteAuthConn) ErrorValueName() string {
	return "oauth_error"
}
func (self *EvernoteAuthConn) StateValueName() string {
	return "action"
}

//EvernoteUser represents the fields that you can extract about a user who uses oauth to log
//in via their evernote account.
type EvernoteUser struct {
	EDAM_NoteStoreURL string
	EDAM_userId       string
}

func (self *EvernoteAuthConn) Fetch(t *oauth.Transport) (interface{}, error) {
	return nil, nil
}
