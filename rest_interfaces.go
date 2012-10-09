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




