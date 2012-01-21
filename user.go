package seven5

import (
	"crypto/bcrypt"
	"errors"
	"fmt"
	"reflect"
	"seven5/store"
	"time"
)

//User is a structure representing a user.  The fields are largely ripped off from the Django
//user model.  Note that to allow user creation there are fields that clients send to the
//server but which are not stored directly.  Many of the fields of this object are not shared
//with clients.
type User struct {
	Username    string `seven5key:"Username"`
	FirstName   string `seven5key:"FirstName"`
	LastName    string `seven5key:"LastName"`
	Email       string    `json:"-" seven5key:"Email"`
	BcryptHash  []byte    `json:"-"`
	Groups      []string  `json:"-"`
	Id          uint64    `seven5All:"false" json:"id"`
	IsStaff     bool      `json:"-"`
	IsActive    bool      `json:"-"`
	IsSuperuser bool      `json:"-"`
	LastLogin   time.Time //UTC ... can be zero valued for never logged in
	Created     time.Time //UTC
	//for input only from the client... annotation indicates should not be stored by the storage layer
	//the json layer will marshal this inbound but will not outbound if it is empty (as expected)
	UserInput_Pwd   string `json:"PlainTextPassword,omitempty" seven5store:"false"`
	UserInput_Super bool   `json:"Superuser,omitempty" seven5store:"false"`
	UserInput_Email string `json:"Email,omitempty" seven5store:"false"`

	//this is for app-specific preferences
	Preference map[string]interface{}

	//standard values in the Preference map are
	//TZoneName string 
	//TZoneOffset int  
	//ImageURL string //usually http://www.gravatar.com/avatar/205e460b479e2e5b48aec07710c08d50
}

var (
	USER_EXISTS = errors.New("username already exists")
)

//Sessions represent a connected user. They contain a copy of the user structure at the time the
//user logged in.  Per-application storage can be put in the Info map.  There is no REST api to
//sessions, they are entirely on the server side except for the session id that is sent to a
//connected client.
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
		return hit[0].Id, USER_EXISTS //nothing to do, we found that username, signal error
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
	//fmt.Printf("created user %s with id %d\n", Username, user.Id)
	return user.Id, err
}

//This struct has no fields because REST implementations should not hold state.  The implementation of
//the methods of this service are a fairly simple example service.  
type UserSvc struct {
	//no fields! REST calls are stateless!
}

//NewUserSvc creates a rest service to process requests about users.  The method's name NewXXXSvc()
//is the conventional way to export a service that deals with the struct XXX. The tune
//tool expects this to be exported by any module that exports User.  (Note that this particular
//service is automatically imported by tune, not discovered, but if it were it would be expected
//to follow this convention.)
func NewUserSvc() Httpified {
	return NewRestHandlerDefault(&UserSvc{}, "user")
}

//Create can only be called if all the needed fields are supplied and the session points to a user
//who is a super user.
func (self *UserSvc) Create(store store.T, ptrToValues interface{}, session *Session) error {
	u := ptrToValues.(*User)
	pwd := u.UserInput_Pwd
	super := u.UserInput_Super
	email := u.UserInput_Email

	u.UserInput_Pwd = ""      //XXX SHOULD BE AUTOMATIC
	u.UserInput_Email = ""    //XXX SHOULD BE AUTOMATIC
	u.UserInput_Super = false //XXX SHOULD BE AUTOMATIC

	id, err := create(store, u.Username, u.FirstName, u.LastName, email, pwd, super)
	if err != nil && err != USER_EXISTS {
		return err
	}

	if err == USER_EXISTS { 
		return NewRestError("Username", "username already exists")
	}
	return store.FindById(ptrToValues, id)
}

//Read can be called by any logged in user.  It does not reveal all the fields because of the
//json marshalling annotations.  This is used by applications to do useful things like display another
//users image or name.
func (self *UserSvc) Read(store store.T, ptrToObject interface{}, session *Session) error {
	return store.FindById(ptrToObject, ptrToObject.(*User).Id)
}

//Update can be called only on a user who is also logged into the session or by a user who is
//logged in as a super user.
func (self *UserSvc) Update(store store.T, ptrToNewValues interface{}, session *Session) error {
	u:=ptrToNewValues.(*User)

	//does it exist?
	other:=&User{Id:u.Id}
	if err:=store.FindById(other,u.Id); err!=nil {
		return err;
	}
	
	if u.FirstName!=""{
		other.FirstName=u.FirstName
	}
	if u.LastName!="" {
		other.LastName=u.LastName
	}
	//you can change some properties of yourself but not these
	noCanDo:=false
	if !session.User.IsSuperuser {
		if u.IsStaff!=other.IsStaff {
			noCanDo=true
		}
		if u.IsSuperuser==other.IsSuperuser {
			noCanDo=true
		}
		if u.IsActive==other.IsActive {
			noCanDo=true
		}
		if noCanDo {
			return NewRestError("_", "operation not permitted")
		}
	} else {
		//you can ONLY change these as superuser
		if u.UserInput_Super!=other.IsStaff {
			other.IsStaff=u.UserInput_Super
		}
		if u.UserInput_Super==other.IsSuperuser {
			other.IsSuperuser=u.UserInput_Super
		}
		//XXX NO WAY TO CHANGE THE ACTIVE FIELD!!!
	}
	
	///XXX concurency bug see seven5/seven5/#10
	if u.UserInput_Email!="" && u.UserInput_Email!=other.Email {
		hits:=make([]*User,0,1)
		if err:=store.FindByKey(&hits,"Email",u.UserInput_Email,uint64(0)); err!=nil {
			return err
		}
		if len(hits)>0 {
			return NewRestError("Email", "email address already in use")
		}
		other.Email = u.UserInput_Email
	}
	if u.Username!="" && u.Username!=other.Username {
		hits:=make([]*User,0,1)
		if err:=store.FindByKey(&hits,"Username",u.Username,uint64(0)); err!=nil {
			return err
		}
		if len(hits)>0 {
			return NewRestError("Username", "username already in use")
		}
		other.Username = u.Username
	}
	
	var err error
	
	if u.UserInput_Pwd !="" {
		other.BcryptHash, err = bcrypt.GenerateFromPassword([]byte(u.UserInput_Pwd), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
	}

	if err=store.Write(other); err!=nil {
		fmt.Printf("error on write in update %v",err)
		return err
	}
	v:=reflect.ValueOf(ptrToNewValues).Elem()
	o:=reflect.ValueOf(other).Elem()
	v.Set(o)
	return nil	
}

//Create can only be called if the session points to a user who is a super user.
func (self *UserSvc) Delete(store store.T, ptrToValues interface{}, session *Session) error {
	u:=ptrToValues.(*User)
	other:=&User{Id:u.Id}
	if err:=store.FindById(other,u.Id); err!=nil {
		return err;
	}
	if err:=store.Delete(other);err!=nil {
		return err
	}
	v:=reflect.ValueOf(ptrToValues).Elem()
	o:=reflect.ValueOf(other).Elem()
	v.Set(o)
	return nil
}

//Find by Key can be called by any logged in user, although the set of key fields are restricted to
//just by username.  Only superuser may search by email.
func (self *UserSvc) FindByKey(store store.T, key string, value string, session *Session, max uint16) (interface{}, error) {
	hits := make([]*User, 0, max)
	err := store.FindByKey(&hits, key, value, uint64(0))
	return hits,err
}

//Validate is called BEFORE any other method in this set (except Make).  This does various kinds of
//simple validation that is common to many of the methods.
func (self *UserSvc) Validate(store store.T, ptrToValues interface{}, op RestfulOp, key string, value string, session *Session) map[string]string {
	//all our methods require a login!
	result := make(map[string]string)
	if session == nil {
		result["_"] = "access is restricted to users that are logged in"
		return result
	}
	user, ok := ptrToValues.(*User)
	if !ok {
		result["_"] = fmt.Sprintf("internal error: unexpected type:%v", ptrToValues)
		return result
	}
	//
	//Note: All the code in this switch is to detect failures, not successes.  Success means exiting
	//the switch.
	//
	switch op {
	case OP_CREATE:
		if !session.User.IsSuperuser {
			result["_"] = "creation of new users is not allowed"
			return result
		}

		ok := true
		if user.Username == "" {
			ok = false
			result["Username"] = "Username is required"
		}
		if user.FirstName == "" {
			ok = false
			result["FirstName"] = "FirstName is required"
		}
		if user.LastName == "" {
			ok = false
			result["LastName"] = "LastName is required"
		}
		if user.UserInput_Email == "" {
			ok = false
			result["Email"] = "Email address is required"
		}
		if user.UserInput_Pwd == "" {
			ok = false
			result["PlainTextPassword"] = "Password is required"
		}
		if !ok {
			return result
		}
	case OP_READ:
		if user.Id == uint64(0) {
			result["id"] = "no id value supplied!"
			return result
		}
	case OP_UPDATE:
		if user.Id == uint64(0) {
			result["id"] = "no id value supplied!"
			return result
		}
		if user.Id != session.User.Id && !session.User.IsSuperuser {
			result["_"] = "updating of user data is not allowed"
			return result
		}
	case OP_DELETE:
		if user.Id == uint64(0) {
			result["id"] = "no id value supplied!"
			return result
		}
		if !session.User.IsSuperuser {
			result["_"] = "deleting users is not allowed"
			return result
		}
	case OP_SEARCH:
		if session == nil {
			result["_"] = "only logged in users may search for other users"
			return result
		}
		if key != "Username" && key!="LastName" && key!="FirstName"{
			result["_"] = "Only searching by Username, FirstName, or LastName is allowed"
			return result;
		}
		if value == "" {
			result[key] = "Must provide a value to search for (can't search for all values yet)"
			return result;
		}
	}
	return nil
}

//Make returns a User struct.	
func (self *UserSvc) Make(id uint64) interface{} {
	return &User{Id: id}
}
