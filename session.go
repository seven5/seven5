package seven5

import (
	_ "code.google.com/p/goauth2/oauth"
	_ "net/http"
	"errors"
	_ "fmt"
)

var PROVIDER_NOT_READY = errors.New("Provider is currently not ready in the session")

//SessionManager is a type that most applications should not need to implement.  It handles the particular
//session semantics in connection with the establishment of Oauth sessions and mapping browser cookies
//to sessions.
type SessionManager interface {
	Find(id string) (Session, error)
	Generate() (Session, error)
	Destroy(id string) error
}

//Session is the minimal interface to a session.  Most applications should not need to implement this
//method and can use the SimpleSession object.
type Session interface {
	SessionId() string
	ProviderReady(string, Fetcher)
	Fetch(string) (interface{},error)
}

//SimpleSession is a default implementation of Session suitable for most applications.
type SimpleSession struct {
	id    string
	fetch map[string]Fetcher
}

//SessionId returns the sessionId (usually a UDID).
func (self *SimpleSession) SessionId() string {
	return self.id
}

func (self *SimpleSession) ProviderReady(provider string, fetcher Fetcher){
	self.fetch[provider]=fetcher
}

//SessionId returns the oauth token associated with this session.
func (self *SimpleSession) Fetch(provider string) (interface{},error) {
	f, ok:= self.fetch[provider]
	if !ok {
		return nil, PROVIDER_NOT_READY
	}
	return f.Fetch()
}

//Fetcher is a type used to represent an authenticated "tunnel" that allows user-specific
//data to be received from a remote service that speaks oauth.  The return value if not nil,
//is specific to the particular fetcher.
type Fetcher interface {
	Fetch() (interface{}, error)
}

//SimpleSessionManager is an implementation of the SessionManager that knows about the semantics
//of getting data from a remote location as part of session creation.
type SimpleSessionManager struct {
	out chan *sessionPacket
}

//sessionPacket is the type exchanged over the channel from the session manager to the go routine
//that needs to handle the (single) mapping from IDs->sessions.
type sessionPacket struct {
	del bool
	id  string
	s   Session
	ret chan Session
}

//NewAuthSessionManager returns an instance of seven5.SessionManager that is actually a Auth session
//manager.  This keeps the sessions in memory, not on disk.
func NewSimpleSessionManager() SessionManager {
	ch := make(chan *sessionPacket)
	go handleSessionChecks(ch)

	return &SimpleSessionManager{
		out: ch,
	}
}

//counter is useful for tests
var packetsProcessed=0

//handleSessionChecks is the goroutine that reads session manager requests and responds based on its
//map.  This assumes that you want to delete the session id if you pass del as true.  If you pass
//a non-nil session we assume you want to create a session.  Otherwise, we do a lookup of the session
//based on the id and return the session (or nil, if not found) through the channel you supplied.
func handleSessionChecks(ch chan *sessionPacket) {
	hash := make(map[string]Session)

	for {
		pkt := <-ch
		packetsProcessed++
		
		//are we doing a delete?
		if pkt.del {
			delete(hash, pkt.id)
			pkt.ret <- nil
			continue
		}

		//are we doing a create?
		if pkt.s != nil {
			hash[pkt.id] = pkt.s
			pkt.ret <- nil
			continue
		}

		//simple query
		s, ok := hash[pkt.id]
		if !ok {
			pkt.ret <- nil
			continue
		}
		pkt.ret <- s
	}
}

func (self *SimpleSessionManager) Generate() (Session, error) {

	//create the default cruft needed for any session
	result := &SimpleSession {
		id: UDID(),
		fetch: make(map[string]Fetcher),
	}
	
	ch:=make(chan Session)
	
	pkt := &sessionPacket{
		del:false,
		id : result.id,
		s: result,
		ret: ch,
	}
	self.out <- pkt
	_ =  <- ch
	close(ch)
		
	//this the now initialized session, ready for connection to the browser
	return result, nil
}

//Destroy is called when a user requests to logout. The session map needs to be updated to no longer
//hold the session.
func (self *SimpleSessionManager) Destroy(id string) error {
	ch:=make(chan Session)
	
	pkt := &sessionPacket{
		del:true,
		id :id,
		s: nil,
		ret: ch,
	}
	self.out <- pkt
	_ =  <- ch
	close(ch)
	
	return nil
}

//Find is called by the cookie management layer to see if a particular session is known to the
//app-specific code (our session manager).
func (self *SimpleSessionManager) Find(id string) (Session, error) {
	
	ch:=make(chan Session)
	
	pkt := &sessionPacket{
		del:false,
		id :id,
		s: nil,
		ret: ch,
	}
	self.out <- pkt
	s:= <- ch
	close(ch)
	
	return s, nil
}
