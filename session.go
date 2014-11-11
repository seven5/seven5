package seven5

import (
	_ "fmt"
	"net/http"
)

var goroutineChannel chan *sessionPacket

//SessionManager is a type that most applications should not need to implement.  It handles the particular
//session semantics in connection with the establishment of Oauth sessions and mapping browser cookies
//to sessions.  SessionManager implementations _must_ be safe to be accessed from multiple
//goroutines.  Because the SessionManager can be accessed a via the pbundle interface, there
//will be multiple goroutines handling requests that could call through this interface.  The
//SimpleSessionManager implentation has this property and may be a useful model for other
//implementors.
type SessionManager interface {
	Find(id string) (Session, error)
	Generate(c OauthConnection, id string, r *http.Request, state string, code string) (Session, error)
	Destroy(id string) error
	Update(Session, interface{}) (Session, error)
}

//Session is the minimal interface to a session.  Most applications should not need to implement this
//method and can use the SimpleSession object.
type Session interface {
	SessionId() string
	UserData() interface{}
}

//SimpleSession is a default implementation of Session suitable for most applications.
type SimpleSession struct {
	id string
	ud interface{}
}

//SessionId returns the sessionId (usually a UDID).
func (self *SimpleSession) SessionId() string {
	return self.id
}

//UserData returns data you want to hang on the session.
func (self *SimpleSession) UserData() interface{} {
	return self.ud
}

//NewSimpleSession returns a new simple session with its SessionId initialized.
//If sid is not "", then it used as the session id and no new UDID is generated.
//This latter feature is to allow control of the session id and should only be
//used in test code.
func NewSimpleSession(userData interface{}, sid string) *SimpleSession {
	var s = sid
	if sid == "" {
		s = UDID()
	}
	return &SimpleSession{s, userData}
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
//manager.  This keeps the sessions in memory, not on disk.  Note that this implementation is
//safe in the face of being called from multiple goroutines.
func NewSimpleSessionManager() *SimpleSessionManager {
	if goroutineChannel == nil {
		goroutineChannel = make(chan *sessionPacket)
		go handleSessionChecks(goroutineChannel)
	}
	return &SimpleSessionManager{
		out: goroutineChannel,
	}
}

//counter is useful for tests
var packetsProcessed = 0

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

//Generate is called when we need to create a new session for a given browser, typically because they
//have successfully authenticated.  This method ignores all the parameters passed but they present
//in the interface for more sophisticated SessionManager implementations.
func (self *SimpleSessionManager) Generate(c OauthConnection, oldId string, r *http.Request, state string, code string) (Session, error) {
	//create the default cruft needed for any session
	result := NewSimpleSession(nil, "")
	return self.Assign(result)
}

//Assign is responsible for connecting the new session to any storage resources needed.  Convenient
//for those overridding the Generate method with their own implementation.
func (self *SimpleSessionManager) Assign(result Session) (Session, error) {

	ch := make(chan Session)

	pkt := &sessionPacket{
		del: false,
		id:  result.SessionId(),
		s:   result,
		ret: ch,
	}
	self.out <- pkt
	_ = <-ch
	close(ch)

	//this the now initialized session
	return result, nil
}

//Update is called from the actual response handlers in the web app to inform
//us that the session needs to change.  XXX Should the PBundle just expose the
//Session manager? XXX
func (self *SimpleSessionManager) Update(result Session, i interface{}) (Session, error) {
	return self.Assign(NewSimpleSession(i, ""))
}

//Destroy is called when a user requests to logout. The session map needs to be updated to no longer
//hold the session.
func (self *SimpleSessionManager) Destroy(id string) error {
	ch := make(chan Session)

	pkt := &sessionPacket{
		del: true,
		id:  id,
		s:   nil,
		ret: ch,
	}
	self.out <- pkt
	_ = <-ch
	close(ch)

	return nil
}

//Find is called by the cookie management layer to see if a particular session is known to the
//app-specific code (our session manager).
func (self *SimpleSessionManager) Find(id string) (Session, error) {

	ch := make(chan Session)

	pkt := &sessionPacket{
		del: false,
		id:  id,
		s:   nil,
		ret: ch,
	}
	self.out <- pkt
	s := <-ch
	close(ch)

	return s, nil
}
