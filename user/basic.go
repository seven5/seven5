// Package user implements two resources that can be used as a basic user
// authentication and manipulation system.  This is intended more as a "getting started quickly"
// tool than a proper user user system.  User systems tend to be very tightly coupled to the
// application, so attempted to build a user system that can handle any possible application is
// pointless.
// 
// The two resources provided are BasicResource and BasicMetaResource.  These _must_ be used in 
// conjunction with a BaseDispatcher, or other Dispatcher that understand the "Allow" protocol since 
// this is how the resources enforce user roles.  
//
// This package assumes that you are going to implement your "users" as a aggregrate that is also
// a seven5 "session." 
// 
package user

import (
	"seven5"
	"errors"
	"code.google.com/p/goauth2/oauth"
	"net/http"
)

var BAD_ID = errors.New("Bad id supplied in request")

//Basic is an interface representing a "basic" user (user.Basic) and only understands the value of
//"email" field on the user.  It requires the implementor to return his Id field (the wire type's Id)
//and to be able to convert to a "wire type" in an application defined way.
type Basic interface {
	Email() string
	SetEmail(string) 
	WireId() seven5.Id
	ToWire() interface{}
}

//Support is the storage interface for the basic user system. It requires that the implementation
//supply a way to check for Admin and Staff roles and that the implementation can supply all known
//users as a list.  This is a good place to connect to a database, if you would like your users
//to be stored that way.  This implementation should be safe to call from multiple goroutines.
//The UpdateFields method is passed a "proposed" instance of the wire type and the current value
//of that type for possible updating.
type Support interface {
	IsAdmin(Basic) bool
	IsStaff(Basic) bool
	KnownUsers() []Basic
	UpdateFields(p interface{}, e Basic)
	Delete(seven5.Id) Basic
	Generate(*oauth.Transport) (seven5.Session,error)
}

//BasicResource is a REST stateless resource.  It does have a field, but this field is set once
//at creation time.  BasicResource represents a user.
type BasicResource struct {
	Sup Support
}

//BasicMetaResource is a REST stateless resource.  It does have a field, but this field is set once
//at creation time.  BasicMetaResource represents meta information about all users and is only accessible
//to Staff or Admin users so this is an easy way to "check" from the client side if you are running
//as a privileged user.
type BasicMetaResource struct {
	Sup Support
}

//This is wire type that is accessible only to staff members.  It can only be read.
type UserMetadataWire struct {
	Id seven5.Id
	NumberUsers seven5.Integer
	NumberStaff seven5.Integer
}

//BasicManager stores a copy of the Support object and creates the necessary resources that are
//going to be needed by the application code.
type BasicManager struct {
	Sup Support
	Wrapped *seven5.SimpleSessionManager
}

//NewBasicManager creates a new basic user manager with the given supporting object.  This should
//be the only copy of Support in the application.
func NewBasicManager(support Support) *BasicManager {
	result:=&BasicManager {
		Wrapped: seven5.NewSimpleSessionManager(),
		Sup:support,
	}
	return result
}

//Find is required by seven5.Session.  Delegated to wrapped simple session manager.
func (self *BasicManager) Find(id string) (seven5.Session, error) {
	return self.Wrapped.Find(id)
}

//Delete is required by seven5.Session.  Delegated to wrapped simple session manager.
func (self *BasicManager) Destroy(id string) error {
	return self.Wrapped.Destroy(id)
}

//Generate is our override of the default implementation in the SimpleSessionManager.  This
//ends up calling the Support method of the same name.
func (self *BasicManager) Generate(t *oauth.Transport, ignore_req *http.Request,
	ignore_state string, ignore_code string) (seven5.Session, error) {
	
	s,err:=self.Sup.Generate(t)
	if err!=nil {
		return nil, err
	}
	return self.Wrapped.Assign(s)
}


//UserResource produces an implementation of a rest resource that is hooked to the Support object
//that was passed to this BasicManager at creation-time.
func (self *BasicManager) UserResource() seven5.RestAll {
	return &BasicResource{self.Sup}
}

//MetaResource produces an implementation of a rest resource that is hooked to the Support object
//that was passed to this BasicManager at creation-time.
func (self *BasicManager) MetaResource() seven5.RestIndex {
	return &BasicMetaResource{self.Sup}
}

//This index a list of size one which is the currently logged in user unless the user is staff.
//Staff users are shown all users, unless the query string specifies self=true.  Because of
//AllowRead this will never be called unless the user at least has a session.
func (self *BasicResource) Index(bundle seven5.PBundle) (interface{}, error) {

	b := bundle.Session().(Basic)
	
	//normal case, should be a list of size one with the _current_ user's info
	list := []interface{}{b.ToWire()}
	_, haveSelf := bundle.Query("self")
	priv:=self.Sup.IsStaff(b) || self.Sup.IsAdmin(b)
	if haveSelf || !priv {
		return list,nil
	}
	list = []interface{}{}
	for _, v := range self.Sup.KnownUsers() {
		list = append(list, v.ToWire())
	}
	return list,nil
}


//Because of Allow, this resource is _only_ called when the logged in user asks
//about himself or if the user is priviledged they can ask about anyone.
func (self *BasicResource) Find(id seven5.Id, bundle seven5.PBundle) (interface{}, error) {
	b := bundle.Session().(Basic)
	
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
	return nil,nil
}

//Put returns the values of the object, after all changes.   This ends up calling the support
//object to copy over the needed fields.
func (self *BasicResource)	Put(id seven5.Id, proposed interface{}, bundle seven5.PBundle)(interface{},error){
		
	var user Basic
	
	for _, v:= range self.Sup.KnownUsers() {
		if v.WireId() == id {
			user = v
			break
		}
	}
	if user==nil {
		return nil, BAD_ID
	}
	self.Sup.UpdateFields(proposed,user)
	
	return user.ToWire(), nil
}

func (self *BasicResource) Delete(id seven5.Id, ignored seven5.PBundle) (interface{}, error) {
	result := self.Sup.Delete(id).ToWire()
	return result, nil
}

func (self *BasicResource) Post(ignored interface{}, ignoredAlso seven5.PBundle) (interface{}, error) {
	//this won't be called because of AllowWrite
	return nil,nil
}

///////////////////////////////////
// ALLOWED ACTION SECTION FOR USER
///////////////////////////////////

//AllowRead checks to insure that you have a session before you are allowed to call
//GET (Indexer) on this resource.
func (self *BasicResource) AllowRead(bundle seven5.PBundle) bool {
	return bundle.Session()!=nil
}

//AllowWrite refuses all requests to Post to this resource because we are assuming that users
//are not "created" in this system but are copied in from external sources like Google.
func (self *BasicResource) AllowWrite(bundle seven5.PBundle) bool {
	//if you change this implementation, you need to change Post()
	return false
}


//Users can only call Find, and Put methods on themselves.  Users cannot call DELETE, even on self.  
//Priviledged members can call any method on any id.
func (self *BasicResource) Allow(id seven5.Id, method string, bundle seven5.PBundle) bool {
	u := bundle.Session().(Basic)
	//not logged in?
	if u == nil {
		return false
	}
	if self.Sup.IsStaff(u) || self.Sup.IsAdmin(u){
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
func (self *BasicMetaResource) Index(bundle seven5.PBundle) (interface{}, error) {
	
	staff:=0
	for _, u:=range self.Sup.KnownUsers() {
		if self.Sup.IsStaff(u) || self.Sup.IsAdmin(u) {
			staff++
		}
	}
	
	metadata := &UserMetadataWire {
		seven5.Id(0),
		seven5.Integer(len(self.Sup.KnownUsers())),
		seven5.Integer(staff),
	}
	list:=[]*UserMetadataWire{metadata}
	return &list,nil
}

//AllowRead checks to insure that you have a session and you are staff before you can call
//this method.  This is the indexer and only method on this resource.
func (self *BasicMetaResource) AllowRead(bundle seven5.PBundle) bool {
	u := bundle.Session().(Basic)
	//not logged in?
	if u == nil {
		return false
	}
	return self.Sup.IsStaff(u)  || self.Sup.IsAdmin(u)
}
