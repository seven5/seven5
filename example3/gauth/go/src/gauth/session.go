package gauth

import (
	"code.google.com/p/goauth2/oauth"
	"fmt"
	"net/http"
	"seven5"//githubme:seven5:
)

//GauthSession is the implementation of seven5.Seven5 for Gauth.
type GauthSession struct {
	seven5.Session
	*seven5.GoogleUser
}

//this is our implementation of the SessionManager that knows about our semantics plus
//a litle bit about google.  A real implementation should deal with the concurrency issues!
type GauthSessionMgr struct {
	sessionMap map[string]seven5.Session
}

func NewSessionManager() seven5.SessionManager {
	return &GauthSessionMgr {
		make(map[string]seven5.Session),
	}
}

func (self *GauthSessionMgr) Generate(udid string, provider string, trans *oauth.Transport,
	req *http.Request, state string) (seven5.Session, error) {
		
	//create the default cruft needed for any session
	result := &GauthSession{
		Session: seven5.SimpleSessionGenerator(udid, provider, trans, req, state),
	}
	//if we have logged in with google, we get the google user info from them
	if provider == "google" {
		u, err := result.GoogleUser.Fetch(trans)
		if err != nil {
			panic(fmt.Sprintf("can't get the user data from google! %s\n", err))
		}
		result.GoogleUser = u
	} else {
		panic("we expected a google session!")
	}
	
	//XXX CONCURRENCY!
	self.sessionMap[udid]=result
	
	//this the now initialized session, ready for connection to the browser
	return result, nil
}

func (self *GauthSessionMgr) Destroy(id string) error{
	//XXX CONCURRENCY!
	delete(self.sessionMap,id)
	delete(knownSessions,id)
	return nil
}

func (self *GauthSessionMgr) Find(id string) (seven5.Session, error){
	// XXX CONCURRENCY!
	result, ok := self.sessionMap[id]
	if ok{
		return result, nil
	}
	return nil, nil
}

