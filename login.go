package seven5

import (
	"crypto/bcrypt"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"seven5/store"
)

//allowShutdown defaults to false.  If it is true, the shutdown message will be processed
//from any successful login by a superuser.
var allowShutdown = false

var shutdownRequested = "Shutdown Requested"

//AllowShutdown sets the flag that enables or disables a client's ability to send a shutdown
//message to the system.  If true, a login request (/api/seven5/login) from a user with
//superuser privileges may supply the
//additional parameter "shutdown=true" to cause the server to shutdown.   This method
//is not one that generally should be called by user code--calls to it are automatically
//generated via the Tune program.
func AllowShutdown(allowed bool) {
	allowShutdown=allowed
}

//LoginGuise is a special http level processor that mounts itself at "/api/seven5/login".  This guise
//is responsible ONLY for checking credentials and creating new sessions for successful
//logins (or issuing errors for bad logins).  In some ways, LoginGuise manipulates users and 
//sessions but these are not done through a restful interface--only the url /api/seven5/login is used
//and only one type of message (POST) is processed.
type LoginGuise struct {
	//we need the implementation of the default HTTP machinery 
	*httpRunnerDefault
	store.T
}

//Name returns "LoginGuise"
func (self *LoginGuise) Name() string {
	return "LoginGuise" //used to generate the UniqueId so don't change this
}

//Pattern returns "/api/seven5/login" which is where it sits in the URL space of mongrel2
func (self *LoginGuise) Pattern() string {
	return "/api/seven5/login"
}

//AppStarting indicates that this object wants to be notified at startup of the application. It
//uses this "hook" to save a copy of the store.T used by the application.
func (self *LoginGuise) AppStarting(log *log.Logger, store store.T) error {
	self.T = store
	return nil
}

//NewLoginGuise creates a new guise... but only one should be needed in any program and this code is 
//called as the program starts by the infrastructure so user code should never need it.
func newLoginGuise() *LoginGuise {
	return &LoginGuise{newHttpRunnerDefault(), nil}
}

//ProcessRequests handles a single request to the LoginGuise. It returns a single response. This 
//handles only one message, a POST, to the url this LoginGuise is associated with.  If 
//the login is successful (username and password are query parameters) the response includes
//the session id.  Otherwise, the response gives an error message.
func (self *LoginGuise) ProcessRequest(req *http.Request) *http.Response {
	var err error
	
	//path:=req.Path
	//method = req.Method
	parsed := req.URL

	resp := new(http.Response)

	values := parsed.Query()
	user := ""
	pwd := ""
	shutdown:=false

	for k, v := range values {
		if k == "username" {
			user = v[0]
			continue
		}
		if k == "password" {
			pwd = v[0]
			continue
		}
		if k=="shutdown" {
			if v[0]=="true" && allowShutdown {
				shutdown=true
			}
		}
	}

	//fmt.Printf("got u and p:'%s' and '%s'\n", user, pwd)

	badCred := make(map[string]string)
	badCred["_"]="invalid username or password"
	
	if user == "" || pwd == "" {
		return formatValidationError(badCred, resp)
	}
	hits := make([]*User, 0, 1)
	err = self.T.FindByKey(&hits, "Username", user, uint64(0))
	if err != nil {
		resp.StatusCode = http.StatusInternalServerError
		resp.Status = fmt.Sprintf("%v", err)
		return resp
	}
	if len(hits) == 0 {
		return formatValidationError(badCred, resp)
	}
	err = bcrypt.CompareHashAndPassword(hits[0].BcryptHash, []byte(pwd))
	if err != nil && err != bcrypt.MismatchedHashAndPasswordError {
		resp.StatusCode = http.StatusInternalServerError
		resp.Status = fmt.Sprintf("%v", err)
		return resp
	}
	if err == bcrypt.MismatchedHashAndPasswordError {
		return formatValidationError(badCred, resp)
	}

	if shutdown && bool(hits[0].IsSuperuser) {
	/*	resp.StatusCode = http.StatusOK
		resp.Status = shutdownMsg
		close(getShutdownChannel())
		return resp*/
		fmt.Printf("Shutdown has been requested by the client (superuser)\n")
		return nil
	}

	//create the new session Id... make sure it's unique
	for {
		s := make([]*Session, 0, 1)
		r := createRandomSessionId()
		//fmt.Printf("checking '%s'\n", r)
		err = self.T.FindByKey(&s, "SessionId", r, uint64(0))
		if err != nil {
			resp.StatusCode = http.StatusInternalServerError
			resp.Status = fmt.Sprintf("%v", err)
			return resp
		}
		if len(s) == 0 {
			break
		}
	}
	session := new(Session)
	session.User = hits[0]
	session.SessionId = createRandomSessionId()
	session.Info = make(map[string]interface{})
	err = self.T.Write(session)
	if err != nil {
		fmt.Printf("error searching for  %s:%v\n", user, err)
		resp.StatusCode = http.StatusInternalServerError
		resp.Status = fmt.Sprintf("%v", err)
		return resp
	}
	//fmt.Printf("successful login %s and placed in session %s\n", user, session.SessionId)
	return fillBody(fmt.Sprintf(`{"sessionId":"%s"}`, session.SessionId), resp)
}

//fillBody creates the body of the response for a message back to the client. It expects to be
//sending the client the json content provided as parameter 1.
func fillBody(jsonContent string, resp *http.Response) *http.Response {
	body := NewBufferCloser()
	body.WriteString(jsonContent)
	resp.Header = make(map[string][]string)
	resp.Header.Set("Content-Type","text/json")
	resp.Body = body
	resp.ContentLength = body.Len()
	resp.StatusCode = 200
	resp.Status = "ok"
	return resp
}

var letter = []int{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}

//createRandomSessionId creates a random session id like xyz-123 that (hopefully) is relatively easy
//to remember when debugging but hard to guess.
func createRandomSessionId() string {
	l := []int{rand.Intn(len(letter)), rand.Intn(len(letter)), rand.Intn(len(letter))}
	n := rand.Intn(1000)
	return fmt.Sprintf("%c%c%c-%03d", letter[l[0]], letter[l[1]], letter[l[2]], n)
}
