package seven5

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mongrel2"
	"net/http"
	"net/url"
	"os"
	"seven5/store"
	"strconv"
	"strings"
)

const (
	json_mime = "application/json"
	//CREATE signals to the validation method that a create operation is in progress.
	OP_CREATE = iota
	//READ signals to the validation method that a read operation is in progress.
	OP_READ
	//UPDATE signals to the validation method that a update operation is in progress.
	OP_UPDATE
	//DELETE signals to the validation method that a delete operation is in progress.
	OP_DELETE
)

//RestfulOp is a type that indications the operation to be performed. It is passed to the Validate
//function of Restful before the operation actually occurs.
type RestfulOp int

//Restful is the key interface for a restful service implementor.  This should be used to make the semantic
//decisions necessary for the particular type implementor such as what user can read a particular
//record or what parameters are required to create a new record.
type Restful interface {
	//Create is called when /api/PLURALNAME is POSTed to.  The values POSTed are passed to this function
	//in the second parameter.  The currently logged in user (who supplied the params) is the last 
	//parameter and this can be nil if there is no logged in user.
	Create(store store.T, ptrToValues interface{}, session *Session) error
	//Read is called when a url like /api/PLURALNAME/72 is accessed (GET) for the object with id=72.
	//Implementations should implement the correct read semantics, such as security.  The id of the
	//desired object is set as the Id field of the ptrToObject and the callee should fill in the 
	//ptrToObject with the values of the other fields.  The
	//user who sent the request, if any, is encoded by the session object.
	Read(store store.T, ptrToObject interface{}, session *Session) error
	//Update is called when a url like /api/PLURALNAME/72 is accessed (PUT) for the object with id=72.
	//Implementations should implement the correct write semantics, such as security. The callee
	//will receive 72 in the Id field of the ptrToNewValues plus all the values supplied by
	//the client. Logged in client, if any, is represented by the session object.
	Update(store store.T, ptrToNewValues interface{}, session *Session) error
	//Update is called when a url like /api/PLURALNAME/72 is accessed (DELETE) for the object with id=72.
	//Implementations should implement the correct delete semantics, such as security. 72 will
	//be passed as the Id value of the ptrToValues but the other fileds in the structure should be ignored.
	Delete(store store.T, ptrToValues interface{}, session *Session) error
	//FindByKey is called when a url like /api/PLURALNAME is accessed (GET) with query parameters
	//that indicate the key to be searched for and the value sought.  The session can be used in
	//cases where there is a need to differentiate based on ownership of the object.  The max
	//number of objects to return is supplied in the last parameter.  In backbone terms, this is
	//is a "fetch."  The returned value should be a slice of up to max ptrs to objects of the
	//apprioprate type for this interface.
	FindByKey(store store.T, key string, value string, session *Session, max int) (interface{}, error)
	//Validate is called BEFORE any other method in this set (except Make).  It can be used to
	//centralize repeated validation checks.  If the validation passes, the receiver should return
	//nil, otherwise a map indicating the field name (key) where the problem was detected and the
	//error message to display to the user as the value.  This will be sent back to the client and
	//the intentded method is NOT called.  For errors that are of a global nature, the error
	//map returned should contain a key of _ with the error message as the value.
	Validate(store store.T, ptrToValues interface{}, op RestfulOp, session *Session) map[string]string
	//Make is called with a given id to create a new instance of the appropriate type for use with
	//this interface.  The id may be zero.  This function is needed because the seven5 library does
	//not know the true type of the structures being stored/retrieved.
	Make(id uint64) interface{}
}

//RestHandlerDefault is the default implementation of a REST processor.  Most clients will use this
//to provide default processing of the HTTP level. Client code probably wants to use
//NewRestHandlerDefault() to provide the necessary details.  A client can just implement a Restful interface
//and then use implementation to deal with all the HTTP cruft.
type RestHandlerDefault struct {
	//we need the implementation of the default HTTP machinery from Seven5
	*HttpRunnerDefault
	//client code should provide this implementation
	svc   Restful
	//name must be supplied by client code, singular and lower case english
	plural string
	//inserted once the application is running
	store store.T
	singular string
}

//NewRestHandlerDefault creates the necessary defaults to route HTTP messages that are intended
//to be restful to the svc provided as the first object.  The second param should be singular,
//english, and lowercase noun like "user" or "fart".
func NewRestHandlerDefault(svc Restful, nounSingular string) *RestHandlerDefault {
	return &RestHandlerDefault{&HttpRunnerDefault{mongrel2.HttpHandlerDefault: &mongrel2.HttpHandlerDefault{new(mongrel2.RawHandlerDefault)}},svc,Pluralize(nounSingular),nil,nounSingular}
}

//For the HTTPified interface, we have to have a name
func (self *RestHandlerDefault) Name() string {
	return self.singular;
}

//Pattern returns the URL that we are using for REST, which is not the same as the name (singular)
func (self *RestHandlerDefault) Pattern() string {
	return "/api/"+self.plural;
}

//Handle a single request of the HTTP level of mongrel. This code primarily figures out what
//REST level call the client is trying to use by looking at the URL and the method.
func (self *RestHandlerDefault) ProcessRequest(req *mongrel2.HttpRequest) *mongrel2.HttpResponse {

	//create a response to go back to the client
	response := new(mongrel2.HttpResponse)
	response.ServerId = req.ServerId
	response.ClientId = []int{req.ClientId}

	fmt.Fprintf(os.Stderr, "method %s called on %s with body '%s'\n", req.Header["METHOD"], req.Path, req.Body)

	var ct string

	//is the body json?
	if ct = req.Header["content-type"]; ct != "" && ct != json_mime {
		response.StatusCode = http.StatusBadRequest
		response.StatusMsg = fmt.Sprintf("only javascript content-type for REST services (was %s)!", ct)
		return response
	}

	//verify that they will ACCEPT json if they aren't sending us any content
	if ct == "" {
		accept := req.Header["accept"]
		if strings.Index(accept, json_mime) == -1 {
			response.StatusCode = http.StatusBadRequest
			response.StatusMsg = fmt.Sprintf("only javascript for REST services (was %s)!", accept)
			return response
		}
	}

	path := req.Path
	method := req.Header["METHOD"]

	//CREATE
	if path == self.Pattern() {
		if method == "POST" {
			return dispatchCreate(req, response, self.svc, self.store)
		} else if method == "GET" {
			return dispatchFetch(req, response, self.svc, self.store)
		}
		response.StatusCode = http.StatusMethodNotAllowed
		response.StatusMsg = "must be POST for save() method from client!"
		return response
	}

	//UPDATE/SAVE
	if method == "PUT" && strings.HasPrefix(path, "/"+self.Pattern()+"/") {
		id := pathToId(path, self.Pattern(), response)
		if id == uint64(0) {
			return response //error in the ID
		}
		return dispatchUpdate(req, response, self.svc, id, self.store)
	}

	//READ
	if method == "GET" && strings.HasPrefix(path, "/"+self.Pattern()+"/") {
		id := pathToId(path, self.Pattern(), response)
		if id == uint64(0) {
			return response //error in the ID
		}
		return dispatchRead(req, response, self.svc, id, self.store)
	}

	response.StatusCode = http.StatusInternalServerError
	response.StatusMsg = "not implemented yet"

	return response
}

//discoverSession is responsible for looking at the http request headers and deciding if a
//session is present and if a session is present, if it is valid.  Note that no matter
//the value oKForNoSession, a *bad* session is always an error (to prevent brute-force attacks)
func discoverSession(req *mongrel2.HttpRequest, response *mongrel2.HttpResponse, store store.T, okForNoSession bool) (*Session, *mongrel2.HttpResponse) {
	sessionId := req.Header["x-seven5-session"]
	sessionFailed := "bad session"

	if sessionId == "" {
		if okForNoSession {
			return nil, nil
		}
		response.StatusCode = http.StatusUnauthorized
		response.StatusMsg = sessionFailed
		return nil, response
	}

	hits := make([]*Session, 0, 1)
	err := store.FindByKey(&hits, "SessionId", sessionId, uint64(0))
	if err != nil || len(hits) == 0 {
		fmt.Printf("bad session attempt %s [no such session?%v, error?%v]\n", sessionId, len(hits) == 0, err)
		response.StatusCode = http.StatusUnauthorized
		response.StatusMsg = sessionFailed
		return nil, response
	}

	return hits[0], nil
}

//formatValidationError puts the appropriate fields into a HTTP response when the validation at
//the REST level has failed.
func formatValidationError(errMap map[string]string, response *mongrel2.HttpResponse) *mongrel2.HttpResponse {
	jsonContent,err:=json.Marshal(&errMap)
	if err!=nil {
		response.StatusCode = http.StatusInternalServerError
		response.StatusMsg = fmt.Sprintf("unable to marshal error to json error: %v",err)
		return response
	}
	response.StatusCode = http.StatusInternalServerError
	response.StatusMsg = fmt.Sprintf("failed to validate on create: %s", err)
	fillBody(string(jsonContent), response)
	return response
	
}

//dispatchCreate is called to handle a POST message to the /api/plural  url.  It is not
//idempotent as it creates a new object.
func dispatchCreate(req *mongrel2.HttpRequest, response *mongrel2.HttpResponse, svc Restful, store store.T) *mongrel2.HttpResponse {
	values := svc.Make(uint64(0))

	session, respErr := discoverSession(req, response, store, false)
	if respErr != nil {
		return respErr
	}

	var err error
	err = json.Unmarshal(req.Body, &values)
	if err != nil {
		response.StatusCode = http.StatusBadRequest
		response.StatusMsg = fmt.Sprintf("json parse error: %s", err)
		return response
	}
	if errMap:=svc.Validate(store, values, OP_CREATE, session); errMap != nil {
		return formatValidationError(errMap, response)
	}
	if err = svc.Create(store, values, session); err != nil {
		response.StatusCode = http.StatusBadRequest
		response.StatusMsg = fmt.Sprintf("failed to create: %s", err)
		return response
	}
	return marshalJsonIntoResponse(response, values)
}

//pathToId takes a path like /api/plural/129  and returns 129 or 0 for error.
func pathToId(path string, name string, response *mongrel2.HttpResponse) uint64 {
	idString := path[len(name)+2:]
	if len(idString) == 0 {
		response.StatusCode = http.StatusBadRequest
		response.StatusMsg = fmt.Sprintf("bad path, no id!(was %s)!", path)
		return uint64(0)
	}

	id, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		response.StatusCode = http.StatusBadRequest
		response.StatusMsg = fmt.Sprintf("bad id in path! (%s)!", path)
		return uint64(0)
	}
	return id
}

//dispatchUpdate is called in response to a call to PUT on /api/plural/192.  It is idempotent.
func dispatchUpdate(req *mongrel2.HttpRequest, response *mongrel2.HttpResponse, svc Restful, id uint64, store store.T) *mongrel2.HttpResponse {
	values := svc.Make(uint64(id))
	var err error

	session, respErr := discoverSession(req, response, store, false)
	if respErr != nil {
		return respErr
	}

	if err = json.Unmarshal(req.Body, &values); err != nil {
		response.StatusCode = http.StatusBadRequest
		response.StatusMsg = fmt.Sprintf("json parse error: %s", err)
		return response
	}
	if errMap:=svc.Validate(store, values, RestfulOp(OP_UPDATE), session); errMap!=nil {
		return formatValidationError(errMap, response)
	}
	if err = svc.Update(store, values, session); err != nil {
		response.StatusCode = http.StatusInternalServerError
		response.StatusMsg = fmt.Sprintf("failed to write update: %s", err)
		return response
	}
	response.StatusCode = 200
	response.StatusMsg = "ok"
	return response
}

//dispatchRead is called in response to a call to GET on /api/plural/192.  It is idempotent.
func dispatchRead(req *mongrel2.HttpRequest, response *mongrel2.HttpResponse, svc Restful, id uint64, store store.T) *mongrel2.HttpResponse {
	values := svc.Make(uint64(id))
	var err error

	session, respErr := discoverSession(req, response, store, true)
	if respErr != nil {
		return respErr
	}
	
	if errMap:=svc.Validate(store, values, RestfulOp(OP_READ), session); errMap!=nil {
		return formatValidationError(errMap, response)
	}

	if err = svc.Read(store, values, session); err != nil {
		response.StatusCode = http.StatusInternalServerError
		response.StatusMsg = fmt.Sprintf("unable to load the item: %s", err)
		return response
	}

	return marshalJsonIntoResponse(response, &values)
}

//marshalJsonIntoResponse is a utility routine used to create a response body for a response to the
//client based on the values parameter and using the json marshalling.
func marshalJsonIntoResponse(response *mongrel2.HttpResponse, values interface{}) *mongrel2.HttpResponse {
	var dataBuffer []byte
	var err error

	if dataBuffer, err = json.Marshal(values); err != nil {
		response.StatusCode = http.StatusInternalServerError
		response.StatusMsg = fmt.Sprintf("unable to compute json for item:%s", err)
		return response
	}
	response.ContentLength = len(dataBuffer)
	response.Body = bytes.NewBuffer(dataBuffer)
	response.StatusCode = 200
	response.StatusMsg = "ok"
	return response

}

//dispatch is called when a GET request is made to /api/plural?foo=bar.  This is called by a client who
//is really searching for a collection of objects and so the response is json array of objects.
func dispatchFetch(req *mongrel2.HttpRequest, response *mongrel2.HttpResponse, svc Restful, store store.T) *mongrel2.HttpResponse {
	session, respErr := discoverSession(req, response, store, true)
	if respErr != nil {
		return respErr
	}
	var hits interface{}
	var err error

	keyToSearchOn := ""
	valueToFind := ""
	uri := req.Header["URI"]
	fmt.Printf("uri to parse: '%s'\n", uri)

	parsed, err := url.Parse(uri)
	if err != nil {
		response.StatusCode = http.StatusBadRequest
		response.StatusMsg = "could not understand URI"
		return response
	}
	values := parsed.Query()

	for k, v := range values {
		if k == "keyName" {
			keyToSearchOn = v[0]
			continue
		}
		if k == "targetValue" {
			valueToFind = v[0]
			continue
		}
	}

	fmt.Printf("key '%s' and value to fetch '%s'\n", keyToSearchOn, valueToFind)

	if hits, err = svc.FindByKey(store, keyToSearchOn, valueToFind, session, 10); err != nil {
		response.StatusCode = http.StatusInternalServerError
		response.StatusMsg = fmt.Sprintf("unable to find the key in fetch: %s", err)
		return response
	}
	return marshalJsonIntoResponse(response, &hits)

}
