package seven5

import (
	"crypto/bcrypt"
	"fmt"
	"reflect"
	"seven5/store"
	"time"
)

//User is a structure representing a user.  The fields are largely ripped off from the Django
//user model.  Note that to allow user creation there are two fields that clients send to the
//server but which are not stored directly.  Many of the fields of this object are not shared
//with clients.
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
	return store.Write(ptrToValues)
}

//Read can be called by any logged in user.  It does not reveal all the fields because of the
//json marshalling annotations.  This is used by applications to do useful things like display another
//users image or name.
func (self *UserSvc) Read(store store.T, ptrToObject interface{}, session *Session) error {
	return store.FindById(ptrToObject, ptrToObject.(User).Id)
}

//Update can be called only on a user who is also logged into the session or by a user who is
//logged in as a super user.
func (self *UserSvc) Update(store store.T, ptrToNewValues interface{}, session *Session) error {
	return store.Write(ptrToNewValues)

}

//Create can only be called if the session points to a user who is a super user.
func (self *UserSvc) Delete(store store.T, ptrToValues interface{}, session *Session) error {
	return store.Delete(ptrToValues)

}

//Find by Key can be called by any logged in user, although the set of key fields are restricted to
//just by username.  Only superuser may search by email.
func (self *UserSvc) FindByKey(store store.T, key string, value string, session *Session, max int) (interface{}, error) {
	hits := make([]*User, 0, max)
	return hits, store.FindByKey(&hits, "Username", value, uint64(0))

}

//Validate is called BEFORE any other method in this set (except Make).  This does various kinds of
//simple validation that is common to many of the methods.
func (self *UserSvc) Validate(store store.T, ptrToValues interface{}, op RestfulOp, session *Session) map[string]string {
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
		v := reflect.ValueOf(*user)
		ok = self.verifyFieldPresent(v, result, "Username", true) && ok
		ok = self.verifyFieldPresent(v, result, "FirstName", true) && ok
		ok = self.verifyFieldPresent(v, result, "LastName", true) && ok
		ok = self.verifyFieldPresent(v, result, "Email", true) && ok
		ok = self.verifyFieldPresent(v, result, "_PlainTextPassword", true) && ok
		if !ok {
			return result
		}
	case OP_READ:
		if user.Id == uint64(0) {
			result["id"] = "no id value supplied!"
			return result
		}
		//only can search by username
		if self.verifyFieldPresent(reflect.ValueOf(*user), result, "Username", true) == false {
			return result
		}
	case OP_UPDATE:
		if user.Id == uint64(0) {
			result["id"] = "no id value supplied!"
			return result
		}
		if user.Id != session.User.Id || session.User.IsSuperuser {
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
	}
	return nil
}

//Make returns a User struct.	
func (self *UserSvc) Make(id uint64) interface{} {
	return &User{Id: id}
}

var noValue reflect.Value

//verifyFieldPresent checks a structure type for a given field name.  If you pass true for the
//last parameter, then the value cannot be the zero value for the type.  Returns true if everything is ok.
//If there is an error, it returns false and updates the result map with an error message as the value
//and the field name as the key.
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
