package seven5

import (
	"fmt"
	"net/url"
)

//PageMapper is an interface for expressing the "landing" pages for a particular action in the user application.
//These are part of the application. This is necessary because this must be handled  programmatically by the 
//back end and can't necessarily use files.  Note that the ErrorPage receives a parameter which is the error
//text and the login page receives both the state parameter and the code (via oauth) that was received
//on successfully getting an oauth token.    These functions should return a URL as a string.
type PageMapper interface {
	ErrorPage(OauthConnector, string) string
	LoginLandingPage(OauthConnector, string, string) string
	LogoutLandingPage(OauthConnector) string
}

//SimplePageMapper encodes the "landing page" URLs for a particular application as constants.
type SimplePageMapper struct {
	errorPage string
	loginOk   string
	logoutOk  string
}

//NewSimplePageMapper returns an PageMapper that has a simple mapping scheme for where
//the application URLs just constants.  
func NewSimplePageMapper(errUrl string, loginUrl string, logoutUrl string) *SimplePageMapper {
	return &SimplePageMapper{
		errorPage: errUrl,
		loginOk:   loginUrl,
		logoutOk:  logoutUrl,
	}
}

//Returns the error page and passes the error text message and service name as query parameters to the 
//page.
func (self *SimplePageMapper) ErrorPage(conn OauthConnector, errorText string) string {
	v := url.Values{
		"service": []string{conn.Name()},
		"error":   []string{errorText},
	}
	return fmt.Sprintf("%s?%s", self.errorPage, v.Encode())
}

//Returns the login landing page constant with the service name and state passed through as a query parameter.  
//Note, the 'code' probably should not be passed through the client side code!
func (self *SimplePageMapper) LoginLandingPage(conn OauthConnector, state string, code string) string {
	v := url.Values{
		"service": []string{conn.Name()},
		"state":   []string{state},
	}
	return fmt.Sprintf("%s?%s", self.loginOk, v.Encode())
}

//Returns the logout landing page constant.  Add's the name of the service to the URL as a query parameter.
func (self *SimplePageMapper) LogoutLandingPage(conn OauthConnector) string {
	v := url.Values{
		"service": []string{conn.Name()},
	}
	return fmt.Sprintf("%s?%s", self.logoutOk, v.Encode())
}
