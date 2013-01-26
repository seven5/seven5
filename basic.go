package seven5

import (
	"errors"
	_ "fmt"
	"net/http"
	"seven5/auth"//githubme:seven5:
)

var BAD_ID = errors.New("Bad id supplied in request")

//BasicUser is an interface representing a "basic" user and only understands the value of
//"email" field on the user.  It requires the implementor to return his Id field (the wire type's Id)
//and to be able to convert to a "wire type" in an application defined way.
type BasicUser interface {
	Email() string
	SetEmail(string)
	WireId() Id
	ToWire() interface{}
}

//BasicUserSupport is the storage interface for the basic user system. It requires that the implementation
//supply a way to check for Admin and Staff roles and that the implementation can supply all known
//users as a list.  This is a good place to connect to a database, if you would like your users
//to be stored that way.  This implementation should be safe to call from multiple goroutines.
//The UpdateFields method is passed a "proposed" instance of the wire type and the current value
//of that type for possible updating.
type BasicUserSupport interface {
	IsAdmin(BasicUser) bool
	IsStaff(BasicUser) bool
	KnownUsers() []BasicUser
	UpdateFields(p interface{}, e BasicUser)
	Delete(Id) BasicUser
	Generate(c auth.OauthConnection, existing Session) (Session, error)
}

//BasicResource is a REST stateless resource.  It does have a field, but this field is set once
//at creation time.  BasicResource represents a user.
type BasicResource struct {
	Sup BasicUserSupport
}

//BasicMetaResource is a REST stateless resource.  It does have a field, but this field is set once
//at creation time.  BasicMetaResource represents meta information about all users and is only accessible
//to Staff or Admin users so this is an easy way to "check" from the client side if you are running
//as a privileged user.
type BasicMetaResource struct {
	Sup BasicUserSupport
}

//This is wire type that is accessible only to staff members.  It can only be read.
type UserMetadataWire struct {
	Id          Id
	NumberUsers Integer
	NumberStaff Integer
}

//BasicManager stores a copy of the BasicUserSupport object and creates the necessary resources that are
//going to be needed by the application code.
type BasicManager struct {
	Sup     BasicUserSupport
	Wrapped *SimpleSessionManager
}

//NewBasicManager creates a new basic user manager with the given supporting object.  This should
//be the only copy of BasicUserSupport in the application.
func NewBasicManager(support BasicUserSupport) *BasicManager {
	result := &BasicManager{
		Wrapped: NewSimpleSessionManager(),
		Sup:     support,
	}
	return result
}

//Find is required by Session.  Delegated to wrapped simple session manager.
func (self *BasicManager) Find(id string) (Session, error) {
	return self.Wrapped.Find(id)
}

//Delete is required by Session.  Delegated to wrapped simple session manager.
func (self *BasicManager) Destroy(id string) error {
	return self.Wrapped.Destroy(id)
}

//Generate is our override of the default implementation in the SimpleSessionManager.  This
//ends up calling the BasicUserSupport method of the same name.
func (self *BasicManager) Generate(c auth.OauthConnection, existingId string, ignore_req *http.Request,
	ignore_state string, ignore_code string) (Session, error) {

	existing, err:=self.Find(existingId)
	if err!=nil {
		return nil, err
	}
	s, err := self.Sup.Generate(c, existing)
	if err != nil {
		return nil, err
	}
	//ignored by our Generate in Support
	if s==nil {
		return nil, nil
	}
	return self.Wrapped.Assign(s)
}

//UserResource produces an implementation of a rest resource that is hooked to the BasicUserSupport object
//that was passed to this BasicManager at creation-time.
func (self *BasicManager) UserResource() RestAll {
	return &BasicResource{self.Sup}
}

//MetaResource produces an implementation of a rest resource that is hooked to the BasicUserSupport object
//that was passed to this BasicManager at creation-time.
func (self *BasicManager) MetaResource() RestIndex {
	return &BasicMetaResource{self.Sup}
}

//This index a list of size one which is the currently logged in user unless the user is staff.
//Staff users are shown all users, unless the query string specifies self=true.  Because of
//AllowRead this will never be called unless the user at least has a session.
func (self *BasicResource) Index(bundle PBundle) (interface{}, error) {

	b := bundle.Session().(BasicUser)

	//normal case, should be a list of size one with the _current_ user's info
	list := []interface{}{b.ToWire()}
	_, haveSelf := bundle.Query("self")
	priv := self.Sup.IsStaff(b) || self.Sup.IsAdmin(b)
	if haveSelf || !priv {
		return list, nil
	}
	list = []interface{}{}
	for _, v := range self.Sup.KnownUsers() {
		list = append(list, v.ToWire())
	}
	return list, nil
}

//Because of Allow, this resource is _only_ called when the logged in user asks
//about himself or if the user is priviledged they can ask about anyone.
func (self *BasicResource) Find(id Id, bundle PBundle) (interface{}, error) {
	b := bundle.Session().(BasicUser)

	//simple case avoids the search
	if b.WireId() == id {
		return b.ToWire(), nil
	}
	//staff/admin users only can do this search
	for _, v := range self.Sup.KnownUsers() {
		if v.WireId() == id {
			return v.ToWire(), nil
		}
	}
	return nil, nil
}

//Put returns the values of the object, after all changes.   This ends up calling the support
//object to copy over the needed fields.
func (self *BasicResource) Put(id Id, proposed interface{}, bundle PBundle) (interface{}, error) {
	var user BasicUser
	for _, v := range self.Sup.KnownUsers() {
		if v.WireId() == id {
			user = v
			break
		}
	}
	if user == nil {
		return nil, BAD_ID
	}
	self.Sup.UpdateFields(proposed, user)

	return user.ToWire(), nil
}

func (self *BasicResource) Delete(id Id, ignored PBundle) (interface{}, error) {
	result := self.Sup.Delete(id).ToWire()
	return result, nil
}

func (self *BasicResource) Post(ignored interface{}, ignoredAlso PBundle) (interface{}, error) {
	//this won't be called because of AllowWrite
	return nil, nil
}

///////////////////////////////////
// ALLOWED ACTION SECTION FOR USER
///////////////////////////////////

//AllowRead checks to insure that you have a session before you are allowed to call
//GET (Indexer) on this resource.
func (self *BasicResource) AllowRead(bundle PBundle) bool {
	return bundle.Session() != nil
}

//AllowWrite refuses all requests to Post to this resource because we are assuming that users
//are not "created" in this system but are copied in from external sources like Google.
func (self *BasicResource) AllowWrite(bundle PBundle) bool {
	//if you change this implementation, you need to change Post()
	return false
}

//Users can only call Find, and Put methods on themselves.  Users cannot call DELETE, even on self.  
//Priviledged members can call any method on any id.
func (self *BasicResource) Allow(id Id, method string, bundle PBundle) bool {
	if bundle.Session()==nil {
		return false
	}
	u := bundle.Session().(BasicUser)
	if self.Sup.IsStaff(u) || self.Sup.IsAdmin(u) {
		return true
	}
	//this is to prevent someone deleting themself
	if method == "DELETE" {
		return false
	}
	return u.WireId() == id
}

///////////////////////////////////
// METADATA RESOURCE
///////////////////////////////////

//Index can just return the metadata because the AllowRead function has already been called to check
//to see if it is ok for the logged in user to read this data.
func (self *BasicMetaResource) Index(bundle PBundle) (interface{}, error) {

	staff := 0
	for _, u := range self.Sup.KnownUsers() {
		if self.Sup.IsStaff(u) || self.Sup.IsAdmin(u) {
			staff++
		}
	}

	metadata := &UserMetadataWire{
		Id(0),
		Integer(len(self.Sup.KnownUsers())),
		Integer(staff),
	}
	list := []*UserMetadataWire{metadata}
	return list, nil
}

//AllowRead checks to insure that you have a session and you are staff before you can call
//this method.  This is the indexer and only method on this resource.
func (self *BasicMetaResource) AllowRead(bundle PBundle) bool {
	//not logged in?
	if bundle.Session()==nil {
		return false
	}
	u := bundle.Session().(BasicUser)
	return self.Sup.IsStaff(u) || self.Sup.IsAdmin(u)
}
