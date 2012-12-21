package seven5

import (
	"code.google.com/p/goauth2/oauth"
	"fmt"
	"net/http"
	"os"
	"strings"
	"path/filepath"
)

const (
	REDIRECT_PATH      = "/auth/%s/oauth2callback"
	REDIRECT_HOST_TEST = "http://localhost:3003"
	HEROKU_HOST        = "https://%s.herokuapp.com"
)

//AuthServiceConnector is an abstraction of a service that can do Oauth-based authentication.
//For now this is a very thin wrapper over code.google.com/p/goauth2/oauth.
type AuthServiceConnector interface {
	AuthURL(AppAuthConfig, string) string
	CodeValueName() string
	ErrorValueName() string
	StateValueName() string
	Name() string
	ExchangeForToken(AppAuthConfig, string) (*oauth.Transport, error)
}

//AppAuthConfig is a simple interface wrapping the ability to know the state of the server
//and interface the oauth provider to the site.
type AppAuthConfig interface {
	IsTest() bool
	AppName() string
	AuthPath(AuthServiceConnector) string
	State(AuthServiceConnector) string
	ClientId(AuthServiceConnector) string
	ClientSecret(AuthServiceConnector) string
	RedirectPath(AuthServiceConnector) string
	RedirectHost(AuthServiceConnector) string
	ErrorPage(AuthServiceConnector, string) string
	SuccessPage(AuthServiceConnector, string) string
}

//HerokuAppAuthConfig understands the seven5 default way of encoding all the oauth related
//information and, if desired, how to handle deployment urls for heroku.   
//In your heroku config, the latter two but not the first should be set.
type HerokuAppAuthConfig struct {
	HerokuName string // use if you are taking the defaults with heroku
}

//NewHerokuAppBasic returns an AppAuthConfig suitable for use with heroku and with the 
//default seven5 environment variable setup.  If you are using this simple interface, for an oauth
//provider named foo and an app called bar you should have the following environment variables 
//set on your development machine: FOO_TEST, BAR_FOO_CLIENT_ID, BAR_FOO_CLIENT_SECRET.
func NewHerokuAppAuthConfig(herokuName string) AppAuthConfig {
	return &HerokuAppAuthConfig{HerokuName: herokuName}
}

func (self *HerokuAppAuthConfig) IsTest() bool {
	return os.Getenv(fmt.Sprintf("%s_TEST", strings.ToUpper(self.AppName()))) != ""
}

func (self *HerokuAppAuthConfig) AppName() string {
	return self.HerokuName
}

func (self *HerokuAppAuthConfig) ClientId(conn AuthServiceConnector) string {
	return os.Getenv(fmt.Sprintf("%s_%s_CLIENT_ID", strings.ToUpper(self.AppName()),
		strings.ToUpper(conn.Name())))
}
func (self *HerokuAppAuthConfig) ClientSecret(conn AuthServiceConnector) string {
	return os.Getenv(fmt.Sprintf("%s_%s_CLIENT_SECRET", strings.ToUpper(self.AppName()),
		strings.ToUpper(conn.Name())))
}

func (self *HerokuAppAuthConfig) ErrorPage(conn AuthServiceConnector, errorText string) string {
	pg := "/error/oauth_error.html"
	return fmt.Sprintf("%s?service=%s&e=%s", toWebUIPath(pg), conn.Name(), errorText)
}
func (self *HerokuAppAuthConfig) SuccessPage(ignored AuthServiceConnector, state string) string {
	fmt.Printf("State? %s\n", state)
	return "/main.html"
}
//State can be used if you need to create a token that will be passed back to you
//when the success state is reached.
func (self *HerokuAppAuthConfig) State(ignored AuthServiceConnector) string {
	return "foobie"
}
func (self *HerokuAppAuthConfig) AuthPath(conn AuthServiceConnector) string {
	return fmt.Sprintf("/auth/%s", conn.Name())
}

func (self *HerokuAppAuthConfig) RedirectHost(conn AuthServiceConnector) string {
	r := fmt.Sprintf(REDIRECT_HOST_TEST)
	if !self.IsTest() {
		r = fmt.Sprintf(HEROKU_HOST, self.AppName())
	}
	return r
}
func (self *HerokuAppAuthConfig) RedirectPath(conn AuthServiceConnector) string {
	return fmt.Sprintf(REDIRECT_PATH, conn.Name())
}

func AddAuthService(handler Handler, cfg AppAuthConfig, conn AuthServiceConnector,
		mgr SessionManager) {

	handler.ServeMux().HandleFunc(cfg.AuthPath(conn),
		func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, conn.AuthURL(cfg, cfg.State(conn)), http.StatusFound)
		})

	handler.ServeMux().HandleFunc(cfg.RedirectPath(conn),
		func(w http.ResponseWriter, r *http.Request) {
			defer r.Body.Close()
			code := r.URL.Query().Get(conn.CodeValueName())
			e := r.URL.Query().Get(conn.ErrorValueName())
			if e != "" {
				http.Redirect(w, r, cfg.ErrorPage(conn, e), http.StatusTemporaryRedirect)
				return
			}
			//exchange it
			trans, err := conn.ExchangeForToken(cfg, code)
			if err != nil {
				error_msg := fmt.Sprintf("unable to finish the token exchange with %s: %s", conn.Name(), err)
				http.Redirect(w, r, cfg.ErrorPage(conn, error_msg), http.StatusTemporaryRedirect)
				return
			}
			state := r.URL.Query().Get(conn.StateValueName())
			session := mgr.Generate(conn.Name(), trans, r, code)
			mgr.AssociateCookie(w, session)
			http.Redirect(w, r, cfg.SuccessPage(conn, state), http.StatusTemporaryRedirect)
		})
}


func toWebUIPath(s string) string {
	dir, file := filepath.Split(s)
	pieces := strings.Split(file, ".")
	last:=pieces[1]
	return fmt.Sprintf("%s_%s.%s.%s",dir,pieces[0],last,last)
}


func UDID() string {
	f, err := os.Open("/dev/urandom")
	if err!=nil {
		panic(fmt.Sprintf("failed to get /dev/urandom! %s", err))
	}
	b := make([]byte, 16)
	_,err=f.Read(b)
	if err!=nil {
		panic(fmt.Sprintf("failed to read /dev/urandom! %s", err))
	}
	f.Close()
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
