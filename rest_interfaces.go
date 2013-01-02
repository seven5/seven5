package seven5

import ()

//Error represents an error response from the implementation of a resource.  These should never need to be
//constructed by hand, use the helper functions like seven5.BadRequest("you called it wrong")
//In some sense, this is actually "non-200" return codes because it is also used on
//resource creation (201) and so forth.
type Error struct {
	//Status code is an HTTP level error code (e.g. 200 ok)
	StatusCode	int
	//Location is a "result" to be shown to the calling code.  If Location is not "", a header 
	//is created in the output called "Location" with this as value. Useful for 201, resource created.
	Location	string
	//Message is a string to return to the caller in the status line.
	Message	string
}

//BaseDocSet represents documentation about the parts of a resource that are the same among all requests. All
//requests in Seven5 return at least a BaseDocSet since they all can accept Headers, Query Parameters and
//return a json value (Result).
type BaseDocSet struct {
	Headers		string
	QueryParameters	string
	Result		string
}

//BodyDocSet is a slight addition to the BaseDocSet so that the value can represent additionally documentation
//about the body parameter.  It would be nice if we could nest a BaseDocSet inside this object but we can't because
//the current JSON marshal/unmarshal does not understand anonymous nested structs. Sigh.
type BodyDocSet struct {
	Headers		string
	QueryParameters	string
	Result		string
	Body		string
}

//Indexer indicates that the struct can return a list of resources.  Implementing structs should return 
//a list of resources from the Index() method.  Implementations should not hold state.  Index will be 
//called to create a response for a GET.
type Indexer interface {
	//Index returns either a json array of objects or an error.  
	//headers is a map from header name to value (not values, as in HTTP).
	//queryParams, ala (?foo=bar) is a map from query parameter name (foo) to value (bar).
	//session can be nil but if present represents the currently logged in user's session.
	Index(headers map[string]string, queryParams map[string]string, session Session) (string, *Error)
	//IndexDoc returns documentation information about the parameters and return results of the Index() call.
	IndexDoc() *BaseDocSet
}

//Finder indicates that the struct can return a particular resource denoted by an id.  Implementing 
//structs should return the data about a resource, given an id.  Implementations
//should not hold state.  This will be called to create a response for a GET.
type Finder interface {
	//Find returns either a json object representing the objects values or an error.  This will be
	//called for a URL like /foo/127 with 127 converted to an int and passed as id.  
	//id will be non-negative
	//headers is a map from header name to value (not values, as in HTTP)
	//queryParams, ala (?foo=bar) is a map from query parameter name (foo) to value (bar)
	//session can be nil but if present represents the currently logged in user's session.
	Find(id Id, headers map[string]string, queryParams map[string]string, session Session) (string, *Error)
	//FindDoc returns documentation information about the parameters and results the Find call.
	FindDoc() *BaseDocSet
}

//Poster indicates that the recevier can create new instances of the correct type.  The poster
//_must_ return a new instance of the resource, complete with a previously unassigned id.
//The instance returned should include all the necessary fields set to default (acceptable) values.
//Most implementations will take a body that is json to have the client indicate paramters
//needed for creation.  This call is obviously not idempotent, but idempotent is a cool word anyway.
//session can be nil but if present represents the currently logged in user's session.
type Poster interface {
	//Returns a new object in the string return (json encoded).  The body is used by most clients
	//to indicate parameters for the creation of the object.
	Post(headers map[string]string, queryParams map[string]string, body string, session Session) (string, *Error)
	//PostDoc returns information about the parameters to and results from the Post() call.
	PostDoc() *BodyDocSet
}

//Puter indicates that the recevier can change values of fields on a type (via PUT).  This call must 
//return the full object with all updated values.  This must be called on a particular instance named by the
//id. Typically the caller sends the new values as json in the body, although we parse the body
//as an HTTP form (up to a given size limit) and put the parsed values in queryParams. 
//This call should be idempotent, because idempotency is fun to say.
//session can be nil but if present represents the currently logged in user's session.
type Puter interface {
	//Returns the new values.
	Put(id Id, headers map[string]string, queryParams map[string]string, body string, session Session) (string, *Error)
	//PutDoc returns information about the parameters to and results from the Put() call.
	PutDoc() *BodyDocSet
}

//Delete indicates that the implementor can delete instances of the resource type.  This call should 
//return the full object state at the time of the delete.  This must be called on a particular instance 
//named by the id and that object is terminated with extreme predjudice. 
//This call is not idempotent, yet is potent.
//session can be nil but if present represents the currently logged in user's session.
type Deleter interface {
	//Returns the values at the time of the deletion.
	Delete(id Id, headers map[string]string, queryParams map[string]string, session Session) (string, *Error)
	//DeleteDoc returns information about the parameters to and results of the Delete() call.
	DeleteDoc() *BaseDocSet
}
