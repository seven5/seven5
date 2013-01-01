package gauth

import (
	"fmt"
	"seven5" //githubme:seven5:
	"strings"
	"encoding/json"
)

//WIRE resource for a user, properties must be public for JSON encoder... all the objects must
//be types known to the seven5 system... this is what goes back and forth over the wire
type GauthUser struct {
	Id       seven5.Id
	Name     seven5.String255
	GoogleId seven5.String255 //will not fit in int64
	Email    seven5.String255
	Pic      seven5.String255
}

//This is the type we use internally, includes extra fields not visible to client side.  Note
//that there is no risk of exposing the type because 1) the extra fields are lower case so
//nobody outside this package can see them, including the json encoder and 2) if returned
//to seven5 instead of a *GauthUser, seven5 will panic because isStaff is not of a type it
//defines.
type GauthUserInternal struct {
	*GauthUser
	isStaff bool
}

//rest resource for the currently logged in user, stateless in the direct sense but actually
//stateful in the sense that the user values returned are controlled by a cookie on the user's
//browser.
type GauthUserResource struct {
}

//KnownUsers maps a google id to a particular user object.  
var knownUsers = make(map[string]*GauthUserInternal)

//known sessions maps a UDID to an key in the known users... this implementation could grow
//without bound
var knownSessions = make(map[string]string)

//Find the user from the session, unless the session is unknown and we need to create the
//user.  If we create the first user, that user is marked "staff".
func GauthUserFromSession(s seven5.Session, createIfNotFound bool) *GauthUserInternal {
	if s == nil {
		return nil //with no session, there is no user!
	}
	session, ok := s.(*GauthSession)
	if !ok {
		panic(fmt.Sprintf("Unexpected session type found: %T!", s))
	}
	result, ok := knownSessions[session.SessionId()]
	if ok {
		//simple case, we have a session and we know the user it maps to (via google id)
		return knownUsers[result]
	}
	//ok, at this point we have a session, but don't know which user it maps to
	//we assume somebody fetched google data into the session
	u, ok := knownUsers[session.GoogleUser.GoogleId]

	//if we don't know the user from previous logins and don't want to create it
	if !ok && !createIfNotFound {
		return nil
	}

	//we've seen this user before, remember which session
	if ok {
		knownSessions[session.SessionId()] = session.GoogleUser.GoogleId
		return u
	}

	//brand new user, we need to create the user and the session
	u = &GauthUserInternal{
		GauthUser: &GauthUser{
			Id:       seven5.Id(len(knownUsers)),
			GoogleId: seven5.String255(session.GoogleUser.GoogleId),
			Name:     seven5.String255(session.GoogleUser.Name),
			Email:    seven5.String255(session.GoogleUser.Email),
			Pic:      seven5.String255(session.GoogleUser.Picture),
		},
		isStaff: len(knownUsers) == 0,
	}
	knownSessions[session.SessionId()] = session.GoogleUser.GoogleId
	knownUsers[session.GoogleUser.GoogleId] = u
	fmt.Printf("Gauth total number of known users:%d\n", len(knownUsers))
	return u
}

//This index a list of size one which is the currently logged in user unless the user is staff.
//Staff users are shown all users, unless the query string specifies self=true.  Because of
//AllowRead this will never be called unless the user at least has a session.
func (STATELESS *GauthUserResource) Index(headers map[string]string,
	qp map[string]string, session seven5.Session) (string, *seven5.Error) {

	internal := GauthUserFromSession(session, true)
	
	//normal case, should be a list of size one
	list := []*GauthUser{internal.GauthUser}
	_, haveSelf := qp["self"]
	if !internal.isStaff || haveSelf {
		return seven5.JsonResult(list, true)
	}
	list = []*GauthUser{}
	for _, v := range knownUsers {
		list = append(list, v.GauthUser)
	}
	return seven5.JsonResult(list, true)
}

//used to create dynamic documentation/api
func (STATELESS *GauthUserResource) IndexDoc() *seven5.BaseDocSet {
	return &seven5.BaseDocSet{
		Headers: "This resource consumes no special headers.  Seven5 applications, however, " +
			"use a cookie named appName-seven5-session to identify a user's session.",

		Result: "This resource collection normally will be of size 1.  The content in this " +
			"list is the information known about the currently logged in user.  For staff " +
			"this list has all users.",

		QueryParameters: "This resource accepts the query parameter 'self' to allow a staff " +
			"member to request only his own information. The value bound to 'self' is ignored.",
	}
}

//Because of Allow, this resource is _only_ called when the logged in user asks
//about himself or if the user is staff they can ask about anyone.
func (STATELESS *GauthUserResource) Find(id seven5.Id, hdrs map[string]string,
	query map[string]string, session seven5.Session) (string, *seven5.Error) {
	internal := GauthUserFromSession(session, false)
	//simple case avoids the search
	if internal.Id != id {
		return seven5.JsonResult(internal.GauthUser, true)
	}
	//staff users only can do this search
	for _, v := range knownUsers {
		if v.Id == id {
			fmt.Printf("Allowing staff user %d to view user %d\n", internal.Id, v.Id)
			return seven5.JsonResult(v.GauthUser, true)
		}
	}
	return seven5.JsonResult(nil, true)
}

//used to generate documentation/api
func (STATELESS *GauthUserResource) FindDoc() *seven5.BaseDocSet {
	return &seven5.BaseDocSet{
		Result: "Returns information about the currently logged in user if the proper id supplied. " +
			"Staff members can call this method on any id.",

		Headers: "This resource consumes no special headers",

		QueryParameters: "This resource ignores any query parameters supplied.",
	}
}

//Put returns the values of the object, after all changes. 
func (STATELESS *GauthUserResource)	Put(id seven5.Id, headers map[string]string, 
	queryParams map[string]string, body string, session seven5.Session) (string,*seven5.Error){
		
	var user *GauthUserInternal
	var client GauthUser
	
	for _, v:= range knownUsers {
		if v.Id == id {
			user = v
			break
		}
	}
	if user==nil {
		return seven5.BadRequest("unacceptable id")
	}
	
	dec:= json.NewDecoder(strings.NewReader(body))
	err := dec.Decode(&client)
	if err!=nil {
		return seven5.BadRequest(fmt.Sprintf("badly formed body: %s", err))
	}
	
	if user.Email!=client.Email && client.Email!="" {
		user.Email = client.Email
	}
	if user.Name!=client.Name  && user.Name!=""{
		user.Name = client.Name
	}
	
	if user.Pic!=client.Pic && user.Pic!="" {
		user.Pic = client.Pic
	}
	
	return seven5.JsonResult(&user.GauthUser,true)
}

//PutDoc returns information about the parameters to and results from the Put() call.  Note
//that this can only be called by logged in users and normal users can only change their
//own data.
func (STATELESS *GauthUserResource) PutDoc() *seven5.BodyDocSet {
	return &seven5.BodyDocSet {
		"This method ignores headers supplied by the client.",
		"This method ignores query parameters supplied by the client.",
		"The result of this method is the full description of the modified object, after all the "+
		"updates have been made.",
		"The body of the put should be a Gauth object in json form.  This method will ignore "+
		"attempts to change the Id or GoogleId fields.",
	}
	
}

///////////////////////////////////
// ALLOWED ACTION SECTION
///////////////////////////////////

//AllowRead checks to insure that you have a session before you are allowed to call
//GET (Indexer) on this resource.
func (STATELESS *GauthUserResource) AllowRead(session seven5.Session) bool {
	return session!=nil
}

//AllowWrite checks to insure that you are logged in as a staff user to do a POST
//(Poster) on this resource. The side effect of POST is to create a new user, so
//this is not allowed except for staff members.
func (STATELESS *GauthUserResource) AllowWrite(session seven5.Session) bool {
	u := GauthUserFromSession(session, false)
	//not logged in?
	if u == nil {
		return false
	}
	return u.isStaff 
}


//Users can only call Find, and Put methods on themselves.  Users cannot call DELETE.  
//Staff members can call any method on any id.
func (STATELESS *GauthUserResource) Allow(id seven5.Id, method string, session seven5.Session) bool {
	u := GauthUserFromSession(session, false)
	//not logged in?
	if u == nil {
		return false
	}
	if u.isStaff {
		return true
	}
	if method == "DELETE" {
		return false
	}
	return u.Id == id
}
