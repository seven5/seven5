package seven5

import (
	"bytes"
	"crypto/bcrypt"
	"fmt"
	"github.com/bradfitz/gomemcache"
	"math/rand"
	"mongrel2"
	"net/http"
	"net/url"
	"seven5/store"
	"time"
)

type User struct {
	Username    string `seven5key:"Username"`
	FirstName   string
	LastName    string
	Email       string `seven5key:"Email"`
	BcryptHash  []byte
	Groups      []string
	Id          uint64 `seven5All:"false"`
	IsStaff     bool
	IsActive    bool
	IsSuperuser bool
	LastLogin   time.Time //UTC ... can be zero valued for never logged in
	Created     time.Time //UTC

	//this is for app-specific preferences
	Preference map[string]interface{}

	//standard values in the Preference map are
	//TZoneName string 
	//TZoneOffset int  
	//ImageURL string //usually http://www.gravatar.com/avatar/205e460b479e2e5b48aec07710c08d50
}

type Session struct {
	Id        uint64 `seven5All:"false"`
	User      *User
	SessionId string `seven5key:"SessionId"`
	Info      map[string]interface{}
}

//CreateSuperUser is customarily done in the pwd.go file associated with your application in the
//init() function.  Such a file should never be checked into the repository!   This function
//tries to find the username given first and if a user with that username already exists, it does
//nothing.
func CreateSuperUser(store store.T, username string, firstName string, lastName string, email string, plainTextPassword string) error {
	return create(store, username, firstName, lastName, email, plainTextPassword, true)
}
func CreateUser(store store.T, username string, firstName string, lastName string, email string, plainTextPassword string) error {
	return create(store, username, firstName, lastName, email, plainTextPassword, false)
}

func create(store store.T, Username string, FirstName string, LastName string, Email string, PlainTextPassword string, isSuper bool) error {
	var err error

	hit := make([]*User, 0, 1)
	err = store.FindByKey(&hit, "Username", Username, uint64(0))
	if err != nil {
		return err
	}
	if len(hit) > 0 {
		return nil //nothing to do, we found that username
	}
	user := new(User)
	user.Username = Username
	user.FirstName = FirstName
	user.LastName = LastName
	user.Email = Email
	user.BcryptHash, err = bcrypt.GenerateFromPassword([]byte(PlainTextPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.IsStaff = isSuper
	user.IsActive = true
	user.IsSuperuser = isSuper
	user.Created = time.Now()
	user.Preference = make(map[string]interface{})

	return store.Write(user)
}

type UserGuise struct {
	//we need the implementation of the default HTTP machinery 
	*HttpRunnerDefault
	store.T
}

func (self *UserGuise) Name() string {
	return "UserGuise" //used to generate the UniqueId so don't change this
}

func (self *UserGuise) IsJson() bool {
	return false
}

func (self *UserGuise) Pattern() string {
	return "/user"
}

func (self *UserGuise) AppStarting(config *ProjectConfig) error {
	self.T = &store.MemcacheGobStore{memcache.New(store.LOCALHOST)}
	return nil
}

//create a new one... but only one should be needed in any program
func NewUserGuise() *UserGuise {
	return &UserGuise{&HttpRunnerDefault{mongrel2.HttpHandlerDefault: &mongrel2.HttpHandlerDefault{new(mongrel2.RawHandlerDefault)}}, nil}
}

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
	if len(hits) == 0 || err == bcrypt.MismatchedHashAndPasswordError {
		return fillBody(badCred, resp)
	}

	//create the new session Id... make sure it's unique
	for  {
		s := make([]*Session, 0, 1)
		r:=createRandomSessionId()
		fmt.Printf("checking '%s'\n",r)
		err = self.T.FindByKey(&s,"SessionId",r, uint64(0))
		if err!=nil {
			resp.StatusCode = http.StatusInternalServerError
			resp.StatusMsg = fmt.Sprintf("%v", err)
			return resp
		}
		if len(s)==0 {
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

func createRandomSessionId() string {
	l := []int{rand.Intn(len(letter)), rand.Intn(len(letter)), rand.Intn(len(letter))}
	n := rand.Intn(1000)
	return fmt.Sprintf("%c%c%c-%03d", letter[l[0]], letter[l[1]], letter[l[2]], n)
}
