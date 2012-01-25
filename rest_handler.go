package seven5

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
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
	//SEARCH signals to the validation method that a fetch (search) operation is in progress
	OP_SEARCH
)

//RestError is used to indicate that an error occurred in a REST method that should NOT be 
//considered an internal error--e.g. the data presented to the storage method was not suitable.
//This results in a 200 response to the client, not a 500 response.  Every effort should be
//made to detect problems in the Validate() method of Restful, rather than using this mechanism as it
//complicates the REST implementation.
type RestError interface {
	//so we are an "error" also
	Error() string
	//ErrorMap returns the error message as the value and the key is the field with the problem,
	//or "_" for the whole thing is really busted.  This is the same format as the
	//result of Validate() on the Restful type.
	ErrorMap() map[string]string
}

//NewRestError creates a default implementation of RestError and returns it.   Pass "_" as the
//fieldName if you mean there is a problem with the whole request.
func NewRestError(field string, errorMessage string) RestError {
	m := make(map[string]string)
	m[field] = errorMessage
	return &restErrorDefault{m, errorMessage}
}

//RestfulOp is a type that indications the operation to be performed. It is passed to the Validate
//function of Restful before the operation actually occurs.
type RestfulOp int

//Restful is the key interface for a restful service implementor.  This should be used to make the semantic
//decisions necessary for the particular type implementor such as what user can read a particular
//record or what parameters are required to create a new record.
//
//It is expected that most of the semantics are expressed in the Validate() method and that errors
//from the storage methods should passed through to the client as errors (500 class) rather than
//200 class resposes that *contain* an error message in the payload.  If you want to return an error
//from a storage method you should return something of type RESTError and not error and the
//REST infrastructure will call the ErrorMap() function to get the "user level" error information.
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
	FindByKey(store store.T, key string, value string, session *Session, max uint16) (interface{}, error)
	//Validate is called BEFORE any other method in this set (except Make).  It can be used to
	//centralize repeated validation checks.  If the validation passes, the receiver should return
	//nil, otherwise a map indicating the field name (key) where the problem was detected and the
	//error message to display to the user as the value.  This will be sent back to the client and
	//the intentded method is NOT called.  For errors that are of a global nature, the error
	//map returned should contain a key of _ with the error message as the value.  The
	//fields of key and value are only used on OP_SEARCH.  OP_SEARCH does not use the ptrToValues
	//parameter but supplies a zero-valued object for that parameter.
	Validate(store store.T, ptrToValues interface{}, op RestfulOp, key string, value string, session *Session) map[string]string
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
	*httpRunnerDefault
	//client code should provide this implementation
	svc Restful
	//name must be supplied by client code, singular and lower case english
	plural string
	//inserted once the application is running
	store    store.T
	singular string
}

//numberedRestFunc is a type that indicates you are going to perform some action at the rest
//layer that just needs an id
type numberedRestFunc func(st store.T, s *Session, service Restful, values interface{}) error

//NewRestHandlerDefault creates the necessary defaults to route HTTP messages that are intended
//to be restful to the svc provided as the first object.  The second param should be singular,
//english, and lowercase noun like "user" or "fart".
func NewRestHandlerDefault(svc Restful, nounSingular string) *RestHandlerDefault {
	return &RestHandlerDefault{newHttpRunnerDefault(), svc, Pluralize(nounSingular), nil, nounSingular}
}

//For the HTTPified interface, we have to have a name
func (self *RestHandlerDefault) Name() string {
	return self.singular
}

//AppStarting is called by the infrastructure as the application is booted.  The REST handler
//saves a copy of the store interface that the application is using so it can do lookups later.
func (self *RestHandlerDefault) AppStarting(log *log.Logger, store store.T) error {
	self.store = store
	return nil
}

//Pattern returns the URL that we are using for REST, which is not the same as the name (singular)
func (self *RestHandlerDefault) Pattern() string {
	return "/api/" + self.plural
}

//Handle a single request of the HTTP level. This code primarily figures out what
//REST level call the client is trying to use by looking at the URL and the method.
func (self *RestHandlerDefault) ProcessRequest(req *http.Request) *http.Response {

	//create a response to go back to the client
	response := new(http.Response)

	//we don't want browser caching in a concurrent app
	if response.Header == nil {
		response.Header = make(map[string][]string)
	}
	response.Header.Set("Cache-Control", "no-cache")

	response.Header.Set("Pragma", "no-cache")
	response.Header.Set("Expires", "0")

	//fmt.Fprintf(os.Stderr, "method %s called on %s with body '%s'\n", req.Header["METHOD"], req.Path, req.Body)

	var ct string

	//is the body json?
	if ct = req.Header.Get(http.CanonicalHeaderKey("Content-Type")); ct != "" && ct != json_mime {
		response.StatusCode = http.StatusBadRequest
		response.Status = fmt.Sprintf("only javascript content-type for REST services (was %s)!", ct)
		return response
	}

	//verify that they will ACCEPT json if they aren't sending us any content
	if ct == "" {
		accept := req.Header.Get(http.CanonicalHeaderKey("accept"))
		if strings.Index(accept, json_mime) == -1 {
			response.StatusCode = http.StatusBadRequest
			response.Status = fmt.Sprintf("only javascript for REST services (was %s)!", accept)
			return response
		}
	}

	path := req.URL.Path
	method := req.Method
	//fmt.Printf("PATH: '%s'  --- Method %s  inside %s\n", path, method, self.Name())

	session, respErr := discoverSession(req, response, self.store)
	if respErr != nil {
		return respErr
	}

	//CREATE
	if path == self.Pattern() {
		if method == "POST" {
			return dispatchCreate(req, response, self.svc, self.store, session)
		} else if method == "GET" {
			return dispatchFetch(req, response, self.svc, self.store, session)
		}
		response.StatusCode = http.StatusMethodNotAllowed
		response.Status = "must be POST for create() method from client or GET for fetch()!"
		return response
	}

	patPlus := self.Pattern() + "/"

	//UPDATE/SAVE
	if method == "PUT" && strings.HasPrefix(path, patPlus) {
		id := pathToId(path, patPlus, response)
		if id == uint64(0) {
			return response //error in the ID
		}
		return dispatchUpdate(req, response, self.svc, id, self.store, session)
	}

	//READ
	if method == "GET" && strings.HasPrefix(path, patPlus) {
		id := pathToId(path, patPlus, response)
		if id == uint64(0) {
			return response //error in the ID
		}
		return dispatchRead(req, response, self.svc, id, self.store, session)
	}

	//DELETE
	if method == "DELETE" && strings.HasPrefix(path, patPlus) {
		id := pathToId(path, patPlus, response)
		if id == uint64(0) {
			return response //error in the ID
		}
		return dispatchDelete(req, response, self.svc, id, self.store, session)
	}

	response.StatusCode = http.StatusInternalServerError
	response.Status = "not implemented yet"

	return response
}

//discoverSession is responsible for looking at the http request headers and deciding if a
//session is present and if a session is present, if it is valid.  Note that no matter
//the value oKForNoSession, a *bad* session is always an error (to prevent brute-force attacks)
func discoverSession(req *http.Request, response *http.Response, store store.T) (*Session, *http.Response) {
	sessionId := req.Header.Get("x-seven5-session")
	sessionFailed := "bad session"

	if sessionId == "" {
		return nil, nil
	}

	hits := make([]*Session, 0, 1)

	err := store.FindByKey(&hits, "SessionId", sessionId, uint64(0))
	if err != nil || len(hits) == 0 {
		response.StatusCode = http.StatusUnauthorized
		response.Status = sessionFailed
		return nil, response
	}

	return hits[0], nil
}

//formatValidationError puts the appropriate fields into a HTTP response when the validation at
//the REST level has failed.
func formatValidationError(errMap map[string]string, response *http.Response) *http.Response {
	shell := make(map[string]map[string]string)
	shell["error"] = errMap
	jsonContent, err := json.Marshal(&shell)
	if err != nil {
		response.StatusCode = http.StatusInternalServerError
		response.Status = fmt.Sprintf("unable to marshal error to json error: %v", err)
		return response
	}
	fillBody(string(jsonContent), response)
	return response

}

//bodyToBytes converts the contents of the body of the request into a set of bytes.  This
//reads the body to EOF, so don't call it when the body might be large.
func bodyToBytes(req *http.Request) ([]byte, error) {
	b := new(bytes.Buffer)
	_, err := b.ReadFrom(req.Body)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

//dispatchCreate is called to handle a POST message to the /api/plural  url.  It is not
//idempotent as it creates a new object.
func dispatchCreate(req *http.Request, response *http.Response, svc Restful, store store.T, session *Session) *http.Response {
	values := svc.Make(uint64(0))
	var err error

	b, err := bodyToBytes(req)
	if err != nil {
		response.StatusCode = http.StatusBadRequest
		response.Status = fmt.Sprintf("buffer conversion error: %s", err)
		return response
	}
	err = json.Unmarshal(b, &values)
	if err != nil {
		response.StatusCode = http.StatusBadRequest
		response.Status = fmt.Sprintf("json parse error: %s", err)
		return response
	}

	//fmt.Printf("values that have been unmarshalled in dispatch: %+v\n",values);
	if errMap := svc.Validate(store, values, OP_CREATE, "", "", session); errMap != nil {
		return formatValidationError(errMap, response)
	}
	if err = svc.Create(store, values, session); err != nil {
		restError, ok := err.(RestError)
		if ok {
			return formatValidationError(restError.ErrorMap(), response)
		} else {
			response.StatusCode = http.StatusInternalServerError
			response.Status = fmt.Sprintf("failed to create: %s", err)
		}
		return response
	}
	return marshalJsonIntoResponse(response, values)
}

//pathToId takes a path like /api/plural/129  and returns 129 or 0 for error.
func pathToId(path string, name string, response *http.Response) uint64 {
	idString := path[len(name):]
	//fmt.Printf("path to Id '%s'\n",idString)

	if len(idString) == 0 {
		response.StatusCode = http.StatusBadRequest
		response.Status = fmt.Sprintf("bad path, no id!(was %s)!", path)
		return uint64(0)
	}

	id, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		response.StatusCode = http.StatusBadRequest
		response.Status = fmt.Sprintf("bad id in path! (%s)!", path)
		return uint64(0)
	}
	return id
}

//dispatchUpdate is called in response to a call to PUT on /api/plural/192.  It is idempotent.
func dispatchUpdate(req *http.Request, response *http.Response,
	svc Restful, id uint64, ugh store.T, session *Session) *http.Response {

	b, err := bodyToBytes(req)
	if err != nil {
		response.StatusCode = http.StatusInternalServerError
		response.Status = fmt.Sprintf("byte conversion error: %s", err)
		return response
	}
	//extra step: we have to unmarshal the content into the object... and ignore the
	//one provided to us by disp
	updateValues := svc.Make(id)
	err = json.Unmarshal(b, &updateValues)
	if err != nil {
		response.StatusCode = http.StatusInternalServerError
		response.Status = fmt.Sprintf("json parse error: %s", err)
		return response
	}

	return dispatchNumbered(req, response, updateValues, session, RestfulOp(OP_UPDATE), svc, ugh, "update",

		func(st store.T, s *Session, service Restful, IGNORED interface{}) error {
			return service.Update(st, updateValues, s)
		})
}

//dispatchDelete is called in response to a call to DELETE on /api/plural/192. 
func dispatchDelete(req *http.Request, response *http.Response,
	svc Restful, id uint64, ugh store.T, session *Session) *http.Response {

	return dispatchNumbered(req, response, svc.Make(id), session, RestfulOp(OP_DELETE), svc, ugh, "delete",

		func(st store.T, s *Session, service Restful, values interface{}) error {
			return service.Delete(st, values, s)
		})
}

//dispatchRead is called in response to a call to GET on /api/plural/192.  It is idempotent.
func dispatchRead(req *http.Request, response *http.Response,
	svc Restful, id uint64, ugh store.T, session *Session) *http.Response {

	return dispatchNumbered(req, response, svc.Make(id), session, RestfulOp(OP_READ), svc, ugh, "read",

		func(st store.T, s *Session, service Restful, values interface{}) error {
			return service.Read(st, values, s)
		})
}

var noSuchResource = map[string]string{"_": "no such resource"}

//centralize all the code that handles the simple flavor of rest requests like /api/plural/192
func dispatchNumbered(req *http.Request, response *http.Response, values interface{}, session *Session, op RestfulOp, svc Restful, someStore store.T, errorName string, fn numberedRestFunc) *http.Response {
	var err error

	if errMap := svc.Validate(someStore, values, op, "", "", session); errMap != nil {
		return formatValidationError(errMap, response)
	}

	if err = fn(someStore, session, svc, values); err != nil && err != store.ErrorNotFoundInStore {
		restError, ok := err.(RestError)
		if ok {
			return formatValidationError(restError.ErrorMap(), response)
		}
		response.StatusCode = http.StatusInternalServerError
		response.Status = fmt.Sprintf("unable to %s the item: %s", errorName, err)
		return response
	}

	if err == store.ErrorNotFoundInStore {
		return formatValidationError(noSuchResource, response)
	}

	return marshalJsonIntoResponse(response, &values)
}

//marshalJsonIntoResponse is a utility routine used to create a response body for a response to the
//client based on the values parameter and using the json marshalling.
func marshalJsonIntoResponse(response *http.Response, values interface{}) *http.Response {
	var dataBuffer []byte
	var err error

	if dataBuffer, err = json.Marshal(values); err != nil {
		response.StatusCode = http.StatusInternalServerError
		response.Status = fmt.Sprintf("unable to compute json for item:%s", err)
		return response
	}
	b := NewBufferCloserFromBytes(dataBuffer)
	response.ContentLength = b.Len()
	response.Body = b
	response.StatusCode = 200
	response.Status = "ok"
	return response

}

//dispatch is called when a GET request is made to /api/plural?foo=bar.  This is called by a client who
//is really searching for a collection of objects and so the response is json array of objects.
func dispatchFetch(req *http.Request, response *http.Response, svc Restful, store store.T, session *Session) *http.Response {
	var err error

	//1:check the request for errors
	parsed := req.URL
	//fmt.Printf("uri to parse: '%s' '%s'\n", parsed.Path, parsed.RawQuery)

	values := parsed.Query()
	jsonText := values.Get("query")

	searchData := make(map[string]interface{})
	err = json.Unmarshal([]byte(jsonText), &searchData)
	if err != nil {
		response.StatusCode = http.StatusBadRequest
		response.Status = fmt.Sprintf("json parse error: %s", err)
		return response
	}
	//fmt.Printf("object parsed out in fetch is %+v\n", searchData)

	//set these to be the "defaults" if the client doesn't specify them
	keyToSearchOn := ""
	valueToFind := ""
	var max = uint16(10)

	//2:decode search params
	for k, v := range searchData {
		if k == "max" {
			val := reflect.ValueOf(v)
			if val.Kind() != reflect.Float64 {
				response.StatusCode = http.StatusBadRequest
				response.Status = fmt.Sprintf("expected max number of results to be float64 [%v]", val.Kind())
				return response
			}
			max = uint16(val.Float())
			continue
		}
		//normal case of key and value being search terms
		keyToSearchOn = k
		val := reflect.ValueOf(v)
		if val.Kind() != reflect.String {
			response.StatusCode = http.StatusBadRequest
			response.Status = fmt.Sprintf("expected value to search for to be a string[%v]", val.Kind())
			return response
		}
		valueToFind = v.(string)
	}


	//3: run validation
	if errMap := svc.Validate(store, svc.Make(0), OP_SEARCH, keyToSearchOn, valueToFind, session); errMap != nil {
		return formatValidationError(errMap, response)
	}
	//4: run the query
	var hits interface{}

	if hits, err = svc.FindByKey(store, keyToSearchOn, valueToFind, session, max); err != nil {
		restError, ok := err.(RestError)
		if ok {
			return formatValidationError(restError.ErrorMap(), response)
		} else {
			response.StatusCode = http.StatusInternalServerError
			response.Status = fmt.Sprintf("failed to do findByKey: %s", err)
		}
		return response
	}

	//5:everythings ok, send result to client
	return marshalJsonIntoResponse(response, hits)
}

//restErrorDefault is a simple implementation of the RestError type that knows how to print
//out an error message.
type restErrorDefault struct {
	restError    map[string]string
	errorMessage string
}

//Error makes a restErrorDefault meet the contract of Error
func (self restErrorDefault) Error() string {
	return self.errorMessage
}

//ErrorMap() makes a restErrorDefault meet the contract of a RestError
func (self restErrorDefault) ErrorMap() map[string]string {
	return self.restError
}
