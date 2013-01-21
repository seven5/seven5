package seven5

import (
	"fmt"
	"net/http"
	"os"
	"seven5/auth"
	"strings"
)

type AuthDispatcher struct {
	provider   []auth.ServiceConnector
	mux        *ServeMux
	prefix     string
	PageMap    auth.PageMapper
	CookieMap  CookieMapper
	SessionMgr SessionManager
}

//NewAuthDispatcher is a wrapper around NewAuthDispatcherRaw that provides some defaults that work for
//a simple application.  It use SimplePageMapper, SimpleCookieManager, and SimpleSession manager for
//the implementations of the needed object.  It assumes that the application login landing page
//is /login.html and similarly the logout page is /logout.html.  Authentication errors are routed
//the page /error.html.  The supplied application name is used to name the browser cookie.
func NewAuthDispatcher(appName string, prefix string, mux *ServeMux) *AuthDispatcher {
	return NewAuthDispatcherRaw(prefix, mux, auth.NewSimplePageMapper("/login.html", "/logout.html", "/error.html"),
		NewSimpleCookieMapper(appName), NewSimpleSessionManager())
}

//NewAuthDispatcher returns a new auth dispatcher which assumes it is mapped at the prefix provided.
//This should not end with / so mapping at / is passed as "".  The serve mux must be passed because
//because as providers are added to the dispatcher it has to register handlers for them.  Note that
//this dispatcher adds mappings in the mux, based on the prefix provided, so it should not be
//manually added to the ServeMux via the Dispatch() method.
func NewAuthDispatcherRaw(prefix string, mux *ServeMux, pm auth.PageMapper, cm CookieMapper,
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
func (self *AuthDispatcher) AddProvider(p auth.ServiceConnector) {
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
	var targ auth.ServiceConnector
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
	case auth.LOGIN_URL:
		return self.Login(targ, w, r)
	case auth.LOGOUT_URL:
		return self.Logout(targ, w, r)
	case auth.CALLBACK_URL:
		return self.Callback(targ, w, r)
	}

	http.Error(w, fmt.Sprintf("Could not dispatch authentication URL: %s", r.URL), http.StatusNotFound)
	return nil
}

func (self *AuthDispatcher) Login(conn auth.ServiceConnector, w http.ResponseWriter, r *http.Request) *ServeMux {
	state := r.URL.Query().Get(conn.StateValueName())
	http.Redirect(w, r, conn.AuthURL(self.callback(conn), state), http.StatusFound)
	return nil
}

func (self *AuthDispatcher) Logout(conn auth.ServiceConnector, w http.ResponseWriter, r *http.Request) *ServeMux {
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

func (self *AuthDispatcher) Callback(conn auth.ServiceConnector, w http.ResponseWriter, r *http.Request) *ServeMux {
	code := r.URL.Query().Get(conn.CodeValueName())
	e := r.URL.Query().Get(conn.ErrorValueName())

	if e != "" {
		http.Redirect(w, r, self.PageMap.ErrorPage(conn, e), http.StatusTemporaryRedirect)
		return nil
	}
	return self.Connect(conn, code, w, r)
}

func (self *AuthDispatcher) callback(conn auth.ServiceConnector) string {
	return self.prefix + "/" + conn.Name() + "/" + auth.CALLBACK_URL
}

func (self *AuthDispatcher) Connect(conn auth.ServiceConnector, code string, w http.ResponseWriter, r *http.Request) *ServeMux {
	t, err := conn.ExchangeForToken(self.callback(conn), code)
	if err != nil {
		error_msg := fmt.Sprintf("unable to finish the token exchange with %s: %s", conn.Name(), err)
		http.Redirect(w, r, self.PageMap.ErrorPage(conn, error_msg), http.StatusTemporaryRedirect)
		return nil
	}
	state := r.URL.Query().Get(conn.StateValueName())
	session, err := self.SessionMgr.Generate(t, r, state, code)
	if err != nil {
		error_msg := fmt.Sprintf("failed to create session")
		http.Redirect(w, r, self.PageMap.ErrorPage(conn, error_msg), http.StatusTemporaryRedirect)
		return nil
	}
	self.CookieMap.AssociateCookie(w, session)
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
func AuthDispatcherFromRaw(raw *RawDispatcher, mux *ServeMux)  *AuthDispatcher{
	pm:=auth.NewSimplePageMapper("/login.html","/logout.html","error.html")
	return NewAuthDispatcherRaw("/auth", mux, pm, raw.CookieMap, raw.SessionMgr)
}

//AuthDispatcherFromRaw is a convenience method that creates an auth dispatcher from an already
//existing BaseDispatcher.  It requires a handle to the ServeMux because the AuthDispatcher creates
//mappings.  It maps the AuthDispatcher functions to /auth/serviceName/{login,logout,oauth2callback}.
func AuthDispatcherFromBase(b *BaseDispatcher, mux *ServeMux) *AuthDispatcher {
	raw:=b.RawDispatcher
	return AuthDispatcherFromRaw(raw, mux)
}
