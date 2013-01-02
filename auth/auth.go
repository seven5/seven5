package auth

import (
	"code.google.com/p/goauth2/oauth"
	"fmt"
	"net/url"
)

//ServiceConnector is an abstraction of a service that can do Oauth-based authentication.
//For now this is a very thin wrapper over code.google.com/p/goauth2/oauth.
type ServiceConnector interface {
	AuthURL(string, string) string
	CodeValueName() string
	ErrorValueName() string
	StateValueName() string
	Name() string
	ExchangeForToken(string, string) (*oauth.Transport, error)
}

//PageMapper is an interface for expressing what URLs should be used when dealing with
//an external service for authentication.  These are part of the application. This is necessary 
//because this must be handled  programmatically by the back end and can't use files.
type PageMapper interface {
	ErrorPage(ServiceConnector, string) string
	LoginLandingPage(ServiceConnector, string, string) string
	LogoutLandingPage(ServiceConnector) string
}

type AppConnector interface {
	Login()
	Logout()
}


//OauthClientDetail is an interface for finding the specific information needed to connect to
//an Oauth server.  If you don't want to use environment variables as the way you store
//these, you can provide your own implementation of this class.  Note that it can be called
//with different AuthServiceConnectors if you have multiple oauth providers.
type OauthClientDetail interface {
	ClientId(ServiceConnector) string
	ClientSecret(ServiceConnector) string
}


//SimplePageMapper maps the authentication urls to /auth/SERVICENAME such as /auth/google for
//google authentication.  It assumes that all auth services can share the same url space for
//Errors, SuccessfulLogin and SuccessfulLogout. It does use the "state" token that is part
//of the AuthPageMapping api.
type SimplePageMapper struct {
	errorPage string
	loginOk   string
	logoutOk  string
}

//NewSimplePageMapper returns an PageMapper that has a simple mapping scheme for where
//the application URLs just constants.  
func NewSimplePageMapper(errUrl string, loginUrl string, logoutUrl string) PageMapper {
	return &SimplePageMapper{
		errorPage: errUrl,
		loginOk:   loginUrl,
		logoutOk:  logoutUrl,
	}
}

const (
	LOGIN_URL    = "login"
	LOGOUT_URL   = "logout"
	CALLBACK_URL = "oauth2callback"
)

func (self *SimplePageMapper) ErrorPage(conn ServiceConnector, errorText string) string {
	v := url.Values{
		"service": []string{conn.Name()},
		"error": []string{errorText},
	}
	return fmt.Sprintf("%s?%s", self.errorPage, v.Encode())
}

func (self *SimplePageMapper) LoginLandingPage(ignored ServiceConnector, state string, code string) string {
	return self.loginOk
}

func (self *SimplePageMapper) LogoutLandingPage(ignored ServiceConnector) string {
	return self.logoutOk
}
