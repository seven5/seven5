package seven5

import (
	"code.google.com/p/goauth2/oauth"
	"net/http"
)


//SessionManager is a type that most applications will need to implement.  It handles the particular
//session semantics of the application as a delegate from the SimpleSessionManagerBase
type SessionManager interface {
	Find(id string) (Session, error)
	Generate(id string, providerName string, trans *oauth.Transport, req *http.Request,
		state string) (Session, error)
	Destroy(id string) error
}

//Session is the _minimal_ interface to a session.  Most real applications should create their
//implementation of the Session interface.
type Session interface {
	SessionId() string
	Transport(provider string) *oauth.Transport
}

//SimpleSession is a simple, embeddable implementation of Session for code that wants to quickly
//handle the implementation of Session.  Typical applications will simply embed this struct
//in their particular implementation of Session.
type SimpleSession struct {
	id    string
	trans map[string]*oauth.Transport
}


//SimpleSessionGenerator is a SessionGenerator function for applications that want to use
//the SimpleSession as their Session implementation. 
func SimpleSessionGenerator(id string, providerName string, trans *oauth.Transport,
	ignored_request *http.Request, ignored_state string) Session {
	providerToTransport := make(map[string]*oauth.Transport)
	providerToTransport[providerName] = trans
	return &SimpleSession{id: id, trans: providerToTransport}
}


//SessionId returns the sessionId (usually a UDID).
func (self *SimpleSession) SessionId() string {
	return self.id
}

//SessionId returns the oauth token associated with this session.
func (self *SimpleSession) Transport(provider string) *oauth.Transport {
	return self.trans[provider]
}
