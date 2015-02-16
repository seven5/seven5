package oauth2

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
	prefix     string
	PageMap    PageMapper
	CookieMap  CookieMapper
	SessionMgr SessionManager
}

//NewAuthDispatcherRaw returns a new auth dispatcher which assumes it is mapped at the prefix provided.
//This should not end with / so mapping at / is passed as "".  Note that this Dispatcher should
//not be added to the mux since it uses the "AddConnector" method to register particular
//URLs for each provider.    Note that most applications will have a BaseDispatcher
//and so creating this type may be simpler with the function AuthDispatcherFromBase.
func NewAuthDispatcherRaw(prefix string, pm PageMapper, cm CookieMapper,
	sm SessionManager) *AuthDispatcher {
	return &AuthDispatcher{
		prefix:     prefix,
		PageMap:    pm,
		CookieMap:  cm,
		SessionMgr: sm,
	}

}

//AddConnector creates the necessary mappings in the AuthDispatcher (and the associated ServeMux)
//handle connectivity with the provider supplied.
func (self *AuthDispatcher) AddConnector(p OauthConnector, mux *ServeMux) {
	pref := self.prefix + "/" + p.Name() + "/"
	mux.Dispatch(pref+"login", self)
	mux.Dispatch(pref+"logout", self)
	mux.Dispatch(self.callback(p), self)
	self.provider = append(self.provider, p)
}

//Dispatch is the main entry point for dispating an http request.  This is typically called
//with /auth/connectorName/{login,logout,oauth2callback}
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
	p1cred, err := conn.Phase1(state, self.callback(conn))
	if err != nil {
		http.Redirect(w, r, self.PageMap.ErrorPage(conn, err.Error()), http.StatusTemporaryRedirect)
		return nil
	}
	//everything is ok, so proceed to user interaction
	http.Redirect(w, r, conn.UserInteractionURL(p1cred, state, self.callback(conn)), http.StatusFound)
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
	connection, err := conn.Phase2(clientTok, code)
	if err != nil {
		http.Redirect(w, r, self.PageMap.ErrorPage(conn, err.Error()), http.StatusTemporaryRedirect)
		return nil
	}

	state := r.URL.Query().Get(conn.StateValueName())
	v, err := self.CookieMap.Value(r)
	if err != nil && err != NO_SUCH_COOKIE {
		http.Redirect(w, r, self.PageMap.ErrorPage(conn, err.Error()), http.StatusTemporaryRedirect)
		return nil
	}
	/*
		session, err := self.SessionMgr.Generate(connection, v, r, state, code)
		if err != nil {
			error_msg := fmt.Sprintf("failed to create session")
			http.Redirect(w, r, self.PageMap.ErrorPage(conn, error_msg), http.StatusTemporaryRedirect)
			return nil
		}

		if session != nil {
			self.CookieMap.AssociateCookie(w, session)
		}
	*/

	http.Redirect(w, r, self.PageMap.LoginLandingPage(conn, state, code), http.StatusTemporaryRedirect)
	return nil
}

func toWebUIPath(s string) string {
	return fmt.Sprintf("/out%s", s)
}

//AuthDispatcherFromBase is a convenience method that creates an auth dispatcher from an already
//existing BaseDispatcher.  It requires a handle to the ServeMux because the AuthDispatcher creates
//mappings.  It maps the AuthDispatcher functions to /auth/serviceName/{login,logout,oauth2callback}.
func AuthDispatcherFromBase(b *BaseDispatcher) *AuthDispatcher {
	raw := b.RawDispatcher
	pm := NewSimplePageMapper("/oautherror.html", "/login.html", "/logout.html")
	return NewAuthDispatcherRaw("/auth", pm, raw.IO.CookieMapper(), raw.SessionMgr)
}
