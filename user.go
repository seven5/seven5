package seven5

import (
	"bytes"
	"crypto/bcrypt"
	"fmt"
	"github.com/bradfitz/gomemcache"
	"log"
	"math/rand"
	"mongrel2"
	"net/http"
	"net/url"
	"reflect"
	"seven5/store"
	"time"
)

type User struct {
	Username    string `seven5key:"Username"`
	FirstName   string
	LastName    string
	Email       string    `json:"-",seven5key:"Email"`
	BcryptHash  []byte    `json:"-"`
	Groups      []string  `json:"-"`
	Id          uint64    `seven5All:"false",json:"id"`
	IsStaff     bool      `json:"-"`
	IsActive    bool      `json:"-"`
	IsSuperuser bool      `json:"-"`
	LastLogin   time.Time //UTC ... can be zero valued for never logged in
	Created     time.Time //UTC
	//for input only from the client... leading _ indicates should not be stored by the storage layer
	//the json layer will marshal this inbound but will not outbound if it is empty (as expected)
	_PlainTextPassword string `json:"PlainTextPassword,omitempty"`
	_Superuser         bool   `json:"Superuser,omitempty"`

	//this is for app-specific preferences
	Preference map[string]interface{}

	//standard values in the Preference map are
	//TZoneName string 
	//TZoneOffset int  
	//ImageURL string //usually http://www.gravatar.com/avatar/205e460b479e2e5b48aec07710c08d50
}

//Sessions represent a connected user. They contain a copy of the user structure at the time the
//user logged in.  Per-application storage can be put in the Info map.
type Session struct {
	Id        uint64 `seven5All:"false"`
	User      *User
	SessionId string `seven5key:"SessionId"`
	Info      map[string]interface{}
}

//CreateSuperUser is customarily done in the pwd.go file associated with your application in the
//init() function.  Such a file should never be checked into the repository!   This function
//tries to find the username given first and if a user with that username already exists, it does
//nothing.If not found, it creates a super user and returns the userid.
func CreateSuperUser(store store.T, username string, firstName string, lastName string, email string, plainTextPassword string) (uint64, error) {
	return create(store, username, firstName, lastName, email, plainTextPassword, true)
}

//CreateUser is customarily done in the pwd.go file associated with your application in the
//init() function.  Such a file should never be checked into the repository!   This function
//tries to find the username given first and if a user with that username already exists, it does
//nothing.  If not found, it creates a normal user and returns the userid.
func CreateUser(store store.T, username string, firstName string, lastName string, email string, plainTextPassword string) (uint64, error) {
	return create(store, username, firstName, lastName, email, plainTextPassword, false)
}

//create is the building block for CreateUser and CreateSuperUser.
func create(store store.T, Username string, FirstName string, LastName string, Email string, PlainTextPassword string, isSuper bool) (uint64, error) {
	var err error

	hit := make([]*User, 0, 1)
	err = store.FindByKey(&hit, "Username", Username, uint64(0))
	if err != nil {
		return uint64(0), err
	}
	if len(hit) > 0 {
		return hit[0].Id, nil //nothing to do, we found that username
	}
	user := new(User)
	user.Username = Username
	user.FirstName = FirstName
	user.LastName = LastName
	user.Email = Email
	user.BcryptHash, err = bcrypt.GenerateFromPassword([]byte(PlainTextPassword), bcrypt.DefaultCost)
	if err != nil {
		return uint64(0), err
	}
	user.IsStaff = isSuper
	user.IsActive = true
	user.IsSuperuser = isSuper
	user.Created = time.Now()
	user.Preference = make(map[string]interface{})

	err = store.Write(user)
	if err != nil {
		return uint64(0), err
	}
	fmt.Printf("created user %s with id %d\n", Username, user.Id)
	return user.Id, err
}

//UserGuise represents the user api in the URL space.  It mounts itself at /api/user
type UserGuise struct {
	//we need the implementation of the default HTTP machinery 
	*HttpRunnerDefault
	store.T
}

//Name returns "UserGuise"
func (self *UserGuise) Name() string {
	return "UserGuise" //used to generate the UniqueId so don't change this
}

//Pattern returns "/api/user" which is where it sits in the URL space of mongrel2
func (self *UserGuise) Pattern() string {
	return "/api/user"
}

func (self *UserGuise) AppStarting(log *log.Logger) error {
	self.T = &store.MemcacheGobStore{memcache.New(store.LOCALHOST)}
	return nil
}

//NewUserGuise creates a new guise... but only one should be needed in any program and this code is 
//called as the program starts by the infrastructure so user code should never need it.
func NewUserGuise() *UserGuise {
	return &UserGuise{&HttpRunnerDefault{mongrel2.HttpHandlerDefault: &mongrel2.HttpHandlerDefault{new(mongrel2.RawHandlerDefault)}}, nil}
}

//ProcessRequests handles a single request to the UserGuise. It returns a single response. This is
//an unusual REST-like service because has an "extra" method called "login" that used to convert
//a set of credentials into a logged in session.  The login method is /api/user/login, the remainder
//of the urls are standard CRUD for REST.
func (self *UserGuise) ProcessRequest(req *mongrel2.HttpRequest) *mongrel2.HttpResponse {
	var err error
	//path:=req.Path
	_ = req.Header["METHOD"]
	uri := req.Header["URI"]

	resp := new(mongrel2.HttpResponse)
	resp.ServerId = req.ServerId
	resp.ClientId = []int{req.ClientId}

	parsed, err := url.Parse(uri)
	if err != nil {
		resp.StatusCode = http.StatusBadRequest
		resp.StatusMsg = "could not understand URI"
		return resp
	}
	values := parsed.Query()
	user := ""
	pwd := ""

	for k, v := range values {
		if k == "username" {
			user = v[0]
			continue
		}
		if k == "password" {
			pwd = v[0]
			continue
		}
	}

	fmt.Printf("got u and p:'%s' and '%s'\n", user, pwd)

	badCred := `{ "err": "Username or password is incorrect"}`
	if user == "" || pwd == "" {
		return fillBody(badCred, resp)
	}
	hits := make([]*User, 0, 1)
	err = self.T.FindByKey(&hits, "Username", user, uint64(0))
	if err != nil {
		resp.StatusCode = http.StatusInternalServerError
		resp.StatusMsg = fmt.Sprintf("%v", err)
		return resp
	}
	if len(hits) == 0 {
		return fillBody(badCred, resp)
	}
	err = bcrypt.CompareHashAndPassword(hits[0].BcryptHash, []byte(pwd))
	if err != nil && err != bcrypt.MismatchedHashAndPasswordError {
		resp.StatusCode = http.StatusInternalServerError
		resp.StatusMsg = fmt.Sprintf("%v", err)
		return resp
	}
	if err == bcrypt.MismatchedHashAndPasswordError {
		return fillBody(badCred, resp)
	}

	//create the new session Id... make sure it's unique
	for {
		s := make([]*Session, 0, 1)
		r := createRandomSessionId()
		fmt.Printf("checking '%s'\n", r)
		err = self.T.FindByKey(&s, "SessionId", r, uint64(0))
		if err != nil {
			resp.StatusCode = http.StatusInternalServerError
			resp.StatusMsg = fmt.Sprintf("%v", err)
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
		resp.StatusCode = http.StatusBadRequest
		resp.StatusMsg = badCred
		return resp
	}
	fmt.Printf("successful login %s and placed in session %s\n", user, session.SessionId)

	return fillBody(fmt.Sprintf(`{"sessionId":"%s"}`, session.SessionId), resp)
}

//fillBody creates the body of the response for a message back to the client. It expects to be
//sending the client the json content provided as parameter 1.
func fillBody(jsonContent string, resp *mongrel2.HttpResponse) *mongrel2.HttpResponse {
	body := new(bytes.Buffer)
	body.WriteString(jsonContent)
	resp.Header = make(map[string]string)
	resp.Header["Content-Type"] = "text/json"
	resp.Body = body
	resp.ContentLength = body.Len()
	resp.StatusCode = 200
	resp.StatusMsg = "ok"
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

type UserSvc struct {
	//no fields! REST calls are stateless!
}

//Create can only be called if all the needed fields are supplied and the session points to a user
//who is a super user.
func (self *UserSvc) Create(store store.T, ptrToValues interface{}, session *Session) error {
	return store.Write(ptrToValues)
}

//Read can be called by any logged in user.  It does not reveal all the fields because of the
//json marshalling.  This is used by applications to do useful things like display another
//users image or name.
func (self *UserSvc) Read(store store.T, ptrToObject interface{}, id uint64, session *Session) error {
	return store.FindById(ptrToObject, id)
}

//Update can be called only on a user who is also logged into the session or by a user who is
//logged in as a super user.
func (self *UserSvc) Update(store store.T, ptrToNewValues interface{}, id uint64, session *Session) error {
	return store.Write(ptrToNewValues)

}

//Create can only be called if the session points to a user who is a super user.
func (self *UserSvc) Delete(store store.T, id uint64, session *Session) error {
	return store.Delete(self.Make(id))

}

//Find by Key can be called by any logged in user, although the set of key fields are restricted to
//just by username.  Only superuser may search by email.
func (self *UserSvc) FindByKey(store store.T, key string, value string, session *Session, max int) (interface{}, error) {
	hits := make([]*User, 0, max)
	return hits, store.FindByKey(&hits, "Username", value, uint64(0))

}

//Validate is called BEFORE any other method in this set (except Make).  This does various kinds of
//simple validation that is common to many of the methods.
func (self *UserSvc) Validate(store store.T, ptrToValues interface{}, id uint64, op RestfulOp, session *Session) map[string]string {
	//all our methods require a login!
	result := make(map[string]string)
	if session == nil {
		result["_"] = "access is restricted to users that are logged in"
		return result
	}
	//make sure the pointer is ok
	ptrValue := reflect.ValueOf(ptrToValues)
	if ptrValue.Kind() != reflect.Ptr {
		result["_"] = fmt.Sprintf("internal error: wrong type passed to validate:%v", ptrValue)
		return result
	}
	structValue := ptrValue.Elem()
	switch op {
	case OP_CREATE:
		ok := true
		ok = self.verifyFieldPresent(structValue, result, "Username", false) && ok
		ok = self.verifyFieldPresent(structValue, result, "FirstName", false) && ok
		ok = self.verifyFieldPresent(structValue, result, "LastName", false) && ok
		ok = self.verifyFieldPresent(structValue, result, "Email", false) && ok
		ok = self.verifyFieldPresent(structValue, result, "_PlainTextPassword", false) && ok
		if !ok {
			return result
		}
		//fields are ok, is this person a super user?
		if !session.User.IsSuperuser {
			result["_"] = "creation of new users is not allowed"
		}
	case OP_READ:
		if id != uint64(0) {
			return nil
		}
		//only can search by username
		if self.verifyFieldPresent(structValue, result, "Username", true) == false {
			return result
		}
	case OP_UPDATE:
		if session.User.IsSuperuser {
			return nil
		}
		if self.verifyFieldPresent(structValue, result, "Id", true) == false {
			return result
		}
		target := structValue.FieldByName("Id").Uint()
		if target != session.User.Id {
			result["_"] = "updating of user data is not allowed"
			return result
		}
	case OP_DELETE:
		if session.User.IsSuperuser {
			return nil
		}
		result["_"] = "deleting users is not allowed"
		return result
	}
	return nil
}

//Make returns a User struct.	
func (self *UserSvc) Make(id uint64) interface{} {
	return &User{Id: id}
}

var noValue reflect.Value

//verifyFieldPresent checks a structure type for a given field name.  If you pass true then the value
//cannot be the zero value for the type.  Returns true if everything is ok.
func (self *UserSvc) verifyFieldPresent(v reflect.Value, result map[string]string, name string, cannotBeZero bool) bool {

	field := v.FieldByName(name)
	if field == noValue {
		result[name] = fmt.Sprintf("internal error! %s is a required field but it's not in this structure!", name)
		return false
	}
	if cannotBeZero && field == reflect.Zero(field.Type()) {
		result[name] = fmt.Sprintf("%s not supplied, and it is a required field", name)
		return false
	}
	return true
}
