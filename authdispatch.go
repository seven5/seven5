package seven5

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

const (
	callbackURL = "oauth2callback"
)

//AuthDispatcher is a special dispatcher that understands how to interact with Oauth2-based services for 
//authenticating the currently logged in user.  
type AuthDispatcher struct {
	provider   []OauthConnector
	mux        *ServeMux
	prefix     string
	PageMap    PageMapper
	CookieMap  CookieMapper
	SessionMgr SessionManager
}


//NewAuthDispatcher is a wrapper around NewAuthDispatcherRaw that provides some defaults that work for
//a simple application.  It uses SimplePageMapper, SimpleCookieManager, and SimpleSession manager for
//the implementations of the needed object.  It assumes that the application login landing page
//is /login.html and similarly the logout page is /logout.html.  Authentication errors are routed
//the page /oautherror.html.  The supplied application name is used to name the browser cookie.
func NewAuthDispatcher(appName string, prefix string, mux *ServeMux) *AuthDispatcher {
	return NewAuthDispatcherRaw(prefix, mux, 
		NewSimplePageMapper("/oautherror.html", "/login.html", "/logout.html", ),
		NewSimpleCookieMapper(appName), 
		NewSimpleSessionManager())
}

//NewAuthDispatcher returns a new auth dispatcher which assumes it is mapped at the prefix provided.
//This should not end with / so mapping at / is passed as "".  The serve mux must be passed because
//because as providers are added to the dispatcher it has to register handlers for them.  Note that
//this dispatcher adds mappings in the mux, based on the prefix provided, so it should not be
//manually added to the ServeMux via the Dispatch() method.
func NewAuthDispatcherRaw(prefix string, mux *ServeMux, pm PageMapper, cm CookieMapper,
	sm SessionManager) *AuthDispatcher {
	return &AuthDispatcher{
		prefix:     prefix,
		mux:        mux,
		PageMap:    pm,
		CookieMap:  cm,
		SessionMgr: sm,
	}

}

//AddProvider creates the necessary mappings in the AuthDispatcher (and the associated ServeMux)
//handle connectivity with the provider supplied.
func (self *AuthDispatcher) AddProvider(p OauthConnector) {
	pref := self.prefix + "/" + p.Name() + "/"
	self.mux.Dispatch(pref+"login", self)
	self.mux.Dispatch(pref+"logout", self)
	self.mux.Dispatch(self.callback(p), self)

	self.provider = append(self.provider, p)
}

func (self *AuthDispatcher) Dispatch(mux *ServeMux, w http.ResponseWriter, r *http.Request) *ServeMux {
	split := strings.Split(r.URL.Path, "/")
	if split[0] == "" {
		split = split[1:]
	}
	if len(split) < 3 {
		http.Error(w, fmt.Sprintf("Could not dispatch authentication URL: %s", r.URL), http.StatusNotFound)
		return nil
	}
	if split[0] != self.prefix[1:] {
		http.Error(w, fmt.Sprintf("Could not dispatch authentication URL: %s", r.URL), http.StatusNotFound)
		return nil
	}
	var targ OauthConnector
	for _, c := range self.provider {
		if c.Name() == split[1] {
			targ = c
			break
		}
	}
	if targ == nil {
		http.Error(w, fmt.Sprintf("Could not dispatch authentication URL: %s", r.URL), http.StatusNotFound)
		return nil
	}
	switch split[2] {
	case "login":
		return self.Login(targ, w, r)
	case "logout":
		return self.Logout(targ, w, r)
	case callbackURL:
		return self.Callback(targ, w, r)
	}

	http.Error(w, fmt.Sprintf("Could not dispatch authentication URL: %s", r.URL), http.StatusNotFound)
	return nil
}

func (self *AuthDispatcher) Login(conn OauthConnector, w http.ResponseWriter, r *http.Request) *ServeMux {
	state := r.URL.Query().Get(conn.StateValueName())
	p1cred, err:=conn.Phase1(state, self.callback(conn))
	if err!=nil {
		http.Redirect(w, r, self.PageMap.ErrorPage(conn, err.Error()), http.StatusTemporaryRedirect)
		return nil
	}
	//everything is ok, so proceed to user interaction
	http.Redirect(w, r, conn.UserInteractionURL(p1cred,state,self.callback(conn)), http.StatusFound)
	return nil
}

func (self *AuthDispatcher) Logout(conn OauthConnector, w http.ResponseWriter, r *http.Request) *ServeMux {
	id, err := self.CookieMap.Value(r)
	if err != nil && err != NO_SUCH_COOKIE {
		fmt.Fprintf(os.Stderr, "Problem understanding cookie in request: %s", err)
		return nil
	}
	if err != NO_SUCH_COOKIE {
		self.CookieMap.RemoveCookie(w)
		self.SessionMgr.Destroy(id)
	}
	http.Redirect(w, r, self.PageMap.LogoutLandingPage(conn), http.StatusTemporaryRedirect)
	return nil
}

func (self *AuthDispatcher) Callback(conn OauthConnector, w http.ResponseWriter, r *http.Request) *ServeMux {
	code := r.URL.Query().Get(conn.CodeValueName())
	e := r.URL.Query().Get(conn.ErrorValueName())
	tok := r.URL.Query().Get(conn.ClientTokenValueName())
	if e != "" {
		http.Redirect(w, r, self.PageMap.ErrorPage(conn, e), http.StatusTemporaryRedirect)
		return nil
	}
	return self.Connect(conn, tok, code, w, r)
}

func (self *AuthDispatcher) callback(conn OauthConnector) string {
	return self.prefix + "/" + conn.Name() + "/" + callbackURL
}

func (self *AuthDispatcher) Connect(conn OauthConnector, clientTok string, code string, w http.ResponseWriter, r *http.Request) *ServeMux {
	connection, err := conn.Phase2(clientTok, code,)
	if err != nil {
		http.Redirect(w, r, self.PageMap.ErrorPage(conn, err.Error()), http.StatusTemporaryRedirect)
		return nil
	}
	state := r.URL.Query().Get(conn.StateValueName())
	v, err:=self.CookieMap.Value(r)
	if err!=nil && err!=NO_SUCH_COOKIE {
		http.Redirect(w, r, self.PageMap.ErrorPage(conn, err.Error()), http.StatusTemporaryRedirect)
		return nil
	}
	session, err := self.SessionMgr.Generate(connection, v, r, state, code)
	if err != nil {
		error_msg := fmt.Sprintf("failed to create session")
		http.Redirect(w, r, self.PageMap.ErrorPage(conn, error_msg), http.StatusTemporaryRedirect)
		return nil
	}
	if session!=nil {
		self.CookieMap.AssociateCookie(w, session)
	}
	http.Redirect(w, r, self.PageMap.LoginLandingPage(conn, state, code), http.StatusTemporaryRedirect)
	return nil
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
		panic(fmt.Sprintf("failed to read  16 bytes from /dev/urandom! %s", err))
	}
	f.Close()
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

//AuthDispatcherFromRaw is a convenience method that creates an auth dispatcher from an already
//existing RawDispatcher.  It requires a handle to the ServeMux because the AuthDispatcher creates
//mappings. It maps the AuthDispatcher functions to /auth/serviceName/{login,logout,oauth2callback}.
//The application should have pages at oautherror.html, login.html, and logout.html as landing zones
//for the respective actions.
func AuthDispatcherFromRaw(raw *RawDispatcher, mux *ServeMux)  *AuthDispatcher{
	pm:=NewSimplePageMapper("/oautherror.html","/login.html","/logout.html",)
	return NewAuthDispatcherRaw("/auth", mux, pm, raw.CookieMap, raw.SessionMgr)
}

//AuthDispatcherFromRaw is a convenience method that creates an auth dispatcher from an already
//existing BaseDispatcher.  It requires a handle to the ServeMux because the AuthDispatcher creates
//mappings.  It maps the AuthDispatcher functions to /auth/serviceName/{login,logout,oauth2callback}.
func AuthDispatcherFromBase(b *BaseDispatcher, mux *ServeMux) *AuthDispatcher {
	raw:=b.RawDispatcher
	return AuthDispatcherFromRaw(raw, mux)
}
