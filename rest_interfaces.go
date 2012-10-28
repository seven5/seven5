package seven5

import (
)

//Error represents an error response from the implementation of a resource.  These should never be
//constructed by hand, use the helper functions like seven5.BadRequest("you called it wrong")
//In some sense, this is actually "non-200" return codes because it is also used on
//resource creation (201) and so forth.
type Error struct {
	//Status code is an HTTP level error code (e.g. 200 ok)
	StatusCode int
	//Location is a "result" to be shown to the calling code.  If Location is not "", a header 
	//is created in the output called "Location" with this as value. Useful for 201, resource created.
	Location string
	//Message is a string to return to the caller in the status line.
	Message string
}

//Indexer indicates that the struct can return a list of resources.  Implementing structs should return 
//a list of resources from the Index() method.  Implementations should not hold state.  Index will be 
//called to create a response for a GET.
type Indexer interface {
	//Index returns either a json array of objects or an error.  
  //headers is a map from header name to value (not values, as in HTTP).
  //queryParams, ala (?foo=bar) is a map from query parameter name (foo) to value (bar)
	Index(headers map[string]string,queryParams map[string]string) (string,*Error)  
	//IndexDoc returns doc for, respectively: collection, headers, query params.  Returned doc strings can
	//and should be markdown encoded.
	IndexDoc() []string
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
	Find(id Id, headers map[string]string, queryParams map[string]string) (string,*Error)
	//FindDoc returns doc for, respectively: resource, headers, query params.  Returned doc strings can
	//and should be markdown encoded.
	FindDoc() []string 
}

//Poster indicates that the recevier can create new instances of the correct type.  The poster
//_must_ return a new instance of the resource, complete with a previously unassigned id.
//The instance returned should include all the necessary fields set to default (acceptable) values.
//Most implementations will take a body that is json to have the client indicate paramters
//needed for creation.  This call is obviously not idempotent, but idempotent is a cool word anyway.
type Poster interface {
	//Returns a new object in the string return (json encoded).  The body is used by most clients
	//to indicate parameters for the creation of the object.
	Post(headers map[string]string, queryParams map[string]string, body string) (string,*Error)
	//Find returns doc for respectively, returned resource, accepted headers, accepted query parameters
	//and body parameter.  Four total entries in the resultings slice of strings.  Strings can and should
	//be markdown encoded.
	PostDoc() []string 
}

//Puter indicates that the recevier can change values of fields on a type (via PUT).  This call must 
//return the full object with all updated values.  This must be called on a particular instance named by the
//id. Typically the caller sends the new values as json in the body, although we parse the body
//as an HTTP form (up to a given size limit) and put the parsed values in queryParams. 
//This call should be idempotent, because idempotency is fun to say.
type Puter interface {
	//Returns the new values.
	Put(id Id, headers map[string]string, queryParams map[string]string, body string) (string,*Error)
	//Find returns doc for respectively, returned values, accepted headers, accepted query parameters
	//and body parameter.  Four total entries in the resultings slice of strings.  Strings can and should
	//be markdown encoded.
	PutDoc() []string 
}
