package seven5

import (
	"code.google.com/p/goauth2/oauth"
	"fmt"
	"net/http"
	"os"
)

//AuthServiceConnector is an abstraction of a service that can do Oauth-based authentication.
//For now this is a very thin wrapper over code.google.com/p/goauth2/oauth.
type AuthServiceConnector interface {
	AuthURL(AuthPageMapper, string) string
	CodeValueName() string
	ErrorValueName() string
	StateValueName() string
	Name() string
	ExchangeForToken(AuthPageMapper, string) (*oauth.Transport, error)
}

//AuthPageMapper is an interface for expressing what URLs should be used when dealing with
//an external service for authentication.  These are part of the application. This is necessary 
//because this must be handled  programmatically by the back end and can't use files.
type AuthPageMapper interface {
	AuthPath(AuthServiceConnector) string
	State(AuthServiceConnector) string
	RedirectPath(AuthServiceConnector) string
	ErrorPage(AuthServiceConnector, string) string
	LoginLandingPage(AuthServiceConnector, string) string
	LogoutLandingPage(AuthServiceConnector, string) string
}

//SimpleAuthPageMapper maps the authentication urls to /auth/SERVICENAME such as /auth/google for
//google authentication.  It assumes that all auth services can share the same url space for
//Errors, SuccessfulLogin and SuccessfulLogout. It does use the "state" token that is part
//of the AuthPageMapping api.
type SimpleAuthPageMapper struct {
	errorPage string
	loginOk   string
	logoutOk  string
	dep       Deployment
}

//NewSimpleAuthPageMapper returns an AuthPageMapper that has a simple mapping scheme for where
//the login "magic" urls are (e.g. /auth/google) and handles all the other page mappings as
//constants.  
func NewSimpleAuthPageMapper(errUrl string, loginOk string, logoutOk string, dep Deployment) AuthPageMapper {
	return &SimpleAuthPageMapper{
		errorPage: errUrl,
		loginOk:   loginOk,
		logoutOk:  logoutOk,
		dep:       dep,
	}
}

func (self *SimpleAuthPageMapper) ErrorPage(conn AuthServiceConnector, errorText string) string {
	return fmt.Sprintf("%s?service=%s&e=%s", toWebUIPath(self.errorPage), conn.Name(), errorText)
}

func (self *SimpleAuthPageMapper) LoginLandingPage(ignored AuthServiceConnector, ignored_state string) string {
	return toWebUIPath(self.loginOk)
}

func (self *SimpleAuthPageMapper) LogoutLandingPage(ignored AuthServiceConnector, state string) string {
	return toWebUIPath(self.logoutOk)
}

//State can be used if you need to create a token that will be passed back to you
//when the success state is reached.  It will be handed back to you when SuccessPa
func (self *SimpleAuthPageMapper) State(ignored AuthServiceConnector) string {
	return ""
}

func (self *SimpleAuthPageMapper) AuthPath(conn AuthServiceConnector) string {
	return fmt.Sprintf("/auth/%s", conn.Name())
}

func (self *SimpleAuthPageMapper) RedirectPath(conn AuthServiceConnector) string {
	return fmt.Sprintf("%s/oauth2callback", self.AuthPath(conn))
}

//AddAuthService is the "glue" that connects the URL space to your app and to an external
//oauth provider.  The session manager is used, in addition, to create new sessions when
//login is successful.  The URL space is controlled by handler, cfg expresses the information
//about your App deployment (or test mode), and the conn is the service provider interface.
func AddAuthService(handler Handler, pageMap AuthPageMapper, conn AuthServiceConnector,
	cm CookieMapper) {

	handler.ServeMux().HandleFunc(pageMap.AuthPath(conn),
		func(w http.ResponseWriter, r *http.Request) {
			//logout?
			if r.URL.Query().Get("logout") != "" {
				cm.RemoveCookie(w)
				cm.Destroy(r)
				http.Redirect(w, r, pageMap.LogoutLandingPage(conn, pageMap.State(conn)), http.StatusTemporaryRedirect)
				return
			}
			//login
			http.Redirect(w, r, conn.AuthURL(pageMap, pageMap.State(conn)), http.StatusFound)
		})

	handler.ServeMux().HandleFunc(pageMap.RedirectPath(conn),
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			code := r.URL.Query().Get(conn.CodeValueName())
			e := r.URL.Query().Get(conn.ErrorValueName())
			if e != "" {
				http.Redirect(w, r, pageMap.ErrorPage(conn, e), http.StatusTemporaryRedirect)
				return
			}
			//exchange it
			trans, err := conn.ExchangeForToken(pageMap, code)
			if err != nil {
				error_msg := fmt.Sprintf("unable to finish the token exchange with %s: %s", conn.Name(), err)
				http.Redirect(w, r, pageMap.ErrorPage(conn, error_msg), http.StatusTemporaryRedirect)
				return
			}
			state := r.URL.Query().Get(conn.StateValueName())
			session, err := cm.Generate(conn.Name(), trans, r, code)
			if err != nil {
				error_msg := fmt.Sprintf("failed to create session")
				http.Redirect(w, r, pageMap.ErrorPage(conn, error_msg), http.StatusTemporaryRedirect)
				return
			}
			cm.AssociateCookie(w, session)
			http.Redirect(w, r, pageMap.LoginLandingPage(conn, state), http.StatusTemporaryRedirect)
		})
}

func toWebUIPath(s string) string {
	return fmt.Sprintf("/out%s", s)
}

func UDID() string {
	f, err := os.Open("/dev/urandom")
	if err != nil {
		panic(fmt.Sprintf("failed to get /dev/urandom! %s", err))
	}
	b := make([]byte, 16)
	_, err = f.Read(b)
	if err != nil {
		panic(fmt.Sprintf("failed to read /dev/urandom! %s", err))
	}
	f.Close()
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
