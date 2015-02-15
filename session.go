package seven5

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	s5CookiePrefix = "s5" //helps for detecting keys have changed and shuffling attacks
)

//Generator is a type that converts from a small amount of unique info to the
//data that should be stored in a session.  It is called when an http request
//is received (see IOHook) and the user has previously visited this website.
//The uniqueInfo provide is recovered from the session id, typically using the
//SERVER_SESSION_KEY for decryption, so it cannot be forged if this type is
//used with SimpleSessionManager.  The value returned will be associated with
//a newly created session and http processing will continue.   If an error is
//returned, it should be one created with s5.HttpError() to provide a correct
//HTTP status code and message to the client.
type Generator interface {
	Generate(uniqueInfo string) (interface{}, error)
}

//SessionManager is a type that most applications should not need to implement.  It handles the particular
//session semantics in connection with the establishment of user sessions and mapping browser cookies
//to sessions.  SessionManager implementations _must_ be safe to be accessed from multiple
//goroutines.  Because the SessionManager can be accessed a via the pbundle interface, there
//will be multiple goroutines handling requests that could call through this interface.  The
//SimpleSessionManager implentation has this property and may be a useful model for other
//implementors.
type SessionManager interface {
	Assign(id string, ud interface{}, expires time.Time) (Session, error)
	Find(id string) (*SessionReturn, error)
	Destroy(id string) error
	Update(Session, interface{}) (Session, error)
	Generate(string) (interface{}, error)
}

//Session is the minimal interface to a session.  Most applications should not need to implement this
//type and can use the SimpleSession object.
type Session interface {
	SessionId() string
	UserData() interface{}
}

//SimpleSession is a default implementation of Session suitable for most applications.
type SimpleSession struct {
	id string
	ud interface{}
}

//SessionId returns the sessionId. To make sessions stable across runs, the
//SimpleSessionManager encodes a unique id and a expiration time into the
//SessionId and then encrypts that with a key only the session manager knows
//(so the client side cannot see it).
func (self *SimpleSession) SessionId() string {
	return self.id
}

//UserData returns data you want hanging on the session on the session.  Note
//this does not control/affect the unique id that is typically encoded in the
//session id.
func (self *SimpleSession) UserData() interface{} {
	return self.ud
}

//NewSimpleSession returns a new simple session with its SessionId initialized.
//If the sid is "", a new UDID is generated as the session ID, but most applications
//will want to control this so that sessions are stable across runs.
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
	generator Generator
	out       chan *sessionPacket
}

//NewSimpleSessionManager returns an instance of seven5.SessionManager.
//This keeps the sessions in memory, not on disk or database but does try to
//insure that sessions are stable across runs by encrypting the session ids
//with a key only the session manager knows. The key must be supplied in
//the environment variable SERVER_SESSION_KEY or this function panics. That
//key should be a 32 character hex string (see key2hex).  If you pass nil
//as your generator, we are assuming that you will explicitly connect each
//user session via a call to Assign.
func NewSimpleSessionManager(g Generator) *SimpleSessionManager {
	if os.Getenv("SERVER_SESSION_KEY") == "" {
		log.Fatalf("unable to find environment variable SERVER_SESSION_KEY")
	}
	keyRaw := strings.TrimSpace(os.Getenv("SERVER_SESSION_KEY"))
	if len(keyRaw) != aes.BlockSize*2 {
		log.Fatalf("expected SERVER_SESSION_KEY length to be %d, but was %d", aes.BlockSize*2, len(keyRaw))
	}
	buf := make([]byte, aes.BlockSize)
	l, err := hex.Decode(buf, []byte(keyRaw))
	if err != nil {
		log.Fatalf("Unable to decode SERVER_SESSION_KEY, maybe it's not in hex? %v", err)
	}
	if l != aes.BlockSize {
		log.Fatalf("expected SERVER_SESSION_KEY decoded length to be %d, but was %d", aes.BlockSize, l)
	}
	key := buf[0:l]

	result := &SimpleSessionManager{
		out:       make(chan *sessionPacket),
		generator: g,
	}
	go handleSessionChecks(result.out, key)
	return result
}

//NewDumbSessionManager returns a session manager that makes no attempt
//to conceal the client session id from the client, so this is probably only
//useful for tests.
func NewDumbSessionManager() *SimpleSessionManager {
	result := &SimpleSessionManager{
		out:       make(chan *sessionPacket),
		generator: nil,
	}
	go handleSessionChecks(result.out, []byte{})
	return result
}

//counter is useful for tests
var packetsProcessed = 0

type sessionOp int

const (
	_SESSION_OP_DEL sessionOp = iota
	_SESSION_OP_CREATE
	_SESSION_OP_FIND
	_SESSION_OP_UPDATE
)

//sessionPacket is the type exchanged over the channel from the session manager to the go routine
//that needs to handle the (single) mapping from IDs->sessions.
type sessionPacket struct {
	op         sessionOp
	sessionId  string
	uniqueInfo string
	expires    time.Time
	userData   interface{}
	ret        chan *SessionReturn
}

//SessionReturn is returned from a call to Find.  It contains either a Session
//or a unique id, never both.  The uniqueId will be the one recovered from the
//cookie that was originally passed to Assign(), although perhaps not on this run
//of the program.  When Find() returns nil, then there was either no session data
//to recover or the session expired, keys changed or some other event that means
//you better re-check the user.
type SessionReturn struct {
	Session  Session
	UniqueId string
}

//handleSessionChecks is the goroutine that reads session manager requests and responds based on its
//map.  Each operation has a sessionPacket and that has on op to tell us how to
//process each one.
func handleSessionChecks(ch chan *sessionPacket, key []byte) {
	hash := make(map[string]Session)

	var err error
	var block cipher.Block

	if len(key) != 0 {
		block, err = aes.NewCipher(key)
		if err != nil {
			log.Fatalf("unable to get AES cipher: %v", err)
		}
	}

	var result *SessionReturn
	for {
		pkt := <-ch
		packetsProcessed++

		result = nil //safety
		switch pkt.op {

		case _SESSION_OP_DEL:
			_, ok := hash[pkt.sessionId]
			if ok {
				delete(hash, pkt.sessionId)
			}
			result = nil
		case _SESSION_OP_CREATE:
			var sid string
			if block == nil {
				sid = pkt.uniqueInfo
			} else {
				sessionId := computeRawSessionId(pkt.uniqueInfo, pkt.expires)
				sid = encryptSessionId(sessionId, block)
			}
			s := NewSimpleSession(pkt.userData, sid)
			hash[sid] = s
			result = &SessionReturn{Session: s}
		case _SESSION_OP_UPDATE:
			_, ok := hash[pkt.sessionId]
			if !ok {
				result = nil
			} else {
				s := NewSimpleSession(pkt.userData, pkt.sessionId)
				result = &SessionReturn{Session: s}
			}
		case _SESSION_OP_FIND:
			s, ok := hash[pkt.sessionId]
			if !ok {
				if block == nil {
					//this is the dodgy bit
					result = &SessionReturn{UniqueId: pkt.sessionId}
					break
				}
				uniq, ok := decryptSessionId(pkt.sessionId, block)
				if !ok {
					result = nil
					break
				}
				result = &SessionReturn{UniqueId: uniq}
			} else {
				if block == nil {
					result = &SessionReturn{Session: s}
					break
				}
				//expired?
				_, ok := decryptSessionId(pkt.sessionId, block)
				if !ok {
					delete(hash, pkt.sessionId)
					result = nil
				} else {
					result = &SessionReturn{Session: s}
				}
			}
		}
		pkt.ret <- result

	}
}

//Assign is responsible for connecting the unique key for the user to a session.
//The unique key should not contain colon or comma, email address or primary key from
//the database are good choices. The userData will be initially assigned to the
//new session.  Note that you can't change the uniqueInfo later without some
//work, so making it the email address can be trying.  The expiration time
//can be in the past, that is useful for testing. If the expiration time is
//the time zero value, the expiration time of one day from now will be used.
func (self *SimpleSessionManager) Assign(uniqueInfo string, userData interface{}, expires time.Time) (Session, error) {

	ch := make(chan *SessionReturn)

	if expires.IsZero() {
		expires = time.Now().Add(24 * time.Hour)
	}

	pkt := &sessionPacket{
		op:         _SESSION_OP_CREATE,
		uniqueInfo: uniqueInfo,
		userData:   userData,
		expires:    expires,
		ret:        ch,
	}
	self.out <- pkt
	sr := <-ch
	close(ch)

	//this the now initialized session
	return sr.Session, nil
}

//Update is called from the actual response handlers in the web app to inform
//us that the session's user data needs to change. Note that a different
//Session will be returned here and this is the new session for that id.
//The returned session should not be cached but should be looked up with
//Find each time.  Note that you may not change the value of the unique id
//via this method or everything will go very badly wrong.
func (self *SimpleSessionManager) Update(session Session, i interface{}) (Session, error) {
	ch := make(chan *SessionReturn)

	pkt := &sessionPacket{
		op:        _SESSION_OP_UPDATE,
		userData:  i,
		sessionId: session.SessionId(),
		ret:       ch,
	}
	self.out <- pkt
	sr := <-ch
	close(ch)

	//this the now initialized session
	return sr.Session, nil
}

//Destroy is called when a user requests to logout. The value provided should be
//the session id, not the unique user info.
func (self *SimpleSessionManager) Destroy(id string) error {
	ch := make(chan *SessionReturn)

	pkt := &sessionPacket{
		op:        _SESSION_OP_DEL,
		sessionId: id,
		ret:       ch,
	}
	self.out <- pkt
	_ = <-ch
	close(ch)

	return nil
}

//Find is called by the cookie management layer to see if a particular session
//is known.  In the case where the session is not known, the returned value may
//be nil for no information or have the UniqueId that was extracted from the
//sessionid (created on a previous run).  If a uniqueId is returned, not
//a session, it would be wise to create a session immediately since we have
//confirmed that at some point in the past that sesison existed for this user.
func (self *SimpleSessionManager) Find(id string) (*SessionReturn, error) {

	if id == "" {
		log.Printf("[SESSION] likely programming error, called find with id=\"\"")
		return nil, nil
	}
	ch := make(chan *SessionReturn)

	pkt := &sessionPacket{
		op:        _SESSION_OP_FIND,
		sessionId: id,
		ret:       ch,
	}
	self.out <- pkt
	s := <-ch
	close(ch)

	return s, nil
}

//given a uniqueId, compute a related blob of stuff that can be used to
//shove into the session (currently a prefix and an expiration time)
func computeRawSessionId(uniqueId string, t time.Time) string {
	return fmt.Sprintf("%s:%s,%d", s5CookiePrefix, uniqueId, t.Unix())
}

//given a blob of text to encode, returns a hex-encoded string with
//the provided block cipher's ouptut.  note that a random initialization
//vector is used and placed at the front of the cleartext.  this iv
//does not need to be secure, but does need to be random.
func encryptSessionId(cleartext string, block cipher.Block) string {
	ciphertext := make([]byte, len(cleartext)+aes.BlockSize)
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		log.Panicf("failed to read the random stream: %v", err)
	}
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(cleartext))
	output := make([]byte, len(ciphertext)*2)
	hex.Encode(output, ciphertext)
	return string(output)
}

//given a blob of text to decode, checks a few things and returns either
//the originally given unique id and true or "" and false.
func decryptSessionId(encryptedHex string, block cipher.Block) (string, bool) {
	ciphertext := make([]byte, len(encryptedHex)/2)
	l, err := hex.Decode(ciphertext, []byte(encryptedHex))
	if err != nil {
		log.Printf("unable to decode the hex bytes of session id (%s,%d): %v", encryptedHex, len(encryptedHex), err)
		return "", false
	}
	iv := ciphertext[:aes.BlockSize]
	stream := cipher.NewCTR(block, iv)

	cleartext := make([]byte, l-len(iv))
	stream.XORKeyStream(cleartext, ciphertext[aes.BlockSize:])
	s := string(cleartext)
	if !strings.HasPrefix(s, s5CookiePrefix) {
		log.Printf("No cookie prefix found, probably keys changed")
		return "", false
	}
	s = strings.TrimPrefix(s, s5CookiePrefix+":")
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		log.Printf("Failed to understand parts of session id: %s", s)
		return "", false
	}
	t, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		log.Printf("Could not understand expiration time in session id: %s", s)
		return "", false
	}
	expires := time.Unix(t, 0)
	if expires.Before(time.Now()) {
		return "", false
	}
	return parts[0], true
}

//
// Generate returns nil,nil if no Generator was provided at the time of this
// object's creation. If a Generator was provided is it invoked to create
// the user data from this session.
//
func (self *SimpleSessionManager) Generate(uniq string) (interface{}, error) {
	if self.generator == nil {
		return nil, nil
	}
	return self.generator.Generate(uniq)
}
