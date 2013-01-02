package gauth

import (
	"code.google.com/p/goauth2/oauth"
	"fmt"
	"net/http"
	"github.com/seven5/seven5"
)

//GauthSession is the implementation of seven5.Session for Gauth.  
type GauthSession struct {
	seven5.Session
	*seven5.GoogleUser
}

//GauthSessionMgr is our implementation of the SessionManager that knows about our semantics plus
//a litle bit about google.  A real implementation should deal with the concurrency issues!
type GauthSessionMgr struct {
	sessionMap map[string]seven5.Session
}

//NewSessionManager returns an instance of seven5.SessionManager that is actually a GauthSessionMgr.
func NewSessionManager() seven5.SessionManager {
	return &GauthSessionMgr {
		make(map[string]seven5.Session),
	}
}

//Generate is called when we need to create a new instance of our app-specific session type.
//This is typically called when a user successfully logs in with Google so we go ahead and
//pull in the Google data at that point.
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

//Destroy is called when a user requests to logout. The session map needs to be updated to no longer
//hold the session.
func (self *GauthSessionMgr) Destroy(id string) error{
	//XXX CONCURRENCY!
	delete(self.sessionMap,id)
	delete(knownSessions,id)
	return nil
}

//Find is called by the cookie management layer to see if a particular session is known to the
//app-specific code (our session manager).
func (self *GauthSessionMgr) Find(id string) (seven5.Session, error){
	// XXX CONCURRENCY!
	result, ok := self.sessionMap[id]
	if ok{
		return result, nil
	}
	return nil, nil
}

