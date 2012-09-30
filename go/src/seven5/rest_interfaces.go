package seven5

import (
)

/**
  * Represents an error response from the implementation of a resource.  These should never be
  * constructed by hand, use the helper functions like seven5.BadRequest("you called it wrong")
  * In some sense, this is actually "non-200" return codes because it is also used on
  * resource creation (201) and so forth.
  */
type Error struct {
	StatusCode int
	Location string
	Message string
}

/** 
  * Represents the type of a field at the go level, plus documentation about that field.
  */
type FieldDoc struct {
	/**
	  * Instance of the type of this field, zero valued.  Must be a simple type like int32 or string
	  * because it has be storable in a single column in the DB and must be "naturally" convertible to
	  * a json and Dart type
	  */
	ZeroValueExample interface{}
	/**
	  * Documentation about this field, should be markdown encoded.
	  */
	Doc string
}

/**
  * Implementing structs should return a list of resources from the Index method.  Implementations
  * should not hold state.  This will be called to create a response for a GET.
  */
type Indexer interface {
	/**
	  * Return either a jsons array of objects or an error.  
    * @param headers is a map from header name to value (not values, as in HTTP)
    * @param queryParams, ala (?foo=bar) is a map from query parameter name (foo) to value (bar)
	  */
	Index(headers map[string]string,queryParams map[string]string) (string,*Error)  
	/**
	  * Returns doc for, respectively: collection, headers, query params.  Returned doc strings can
	  * and should be markdown encoded.
	  */
	IndexDoc() (string, string, string) 
}


/**
  * Implementing structs should return the data about a resource, given an id.  Implementations
  * should not hold state.  This will be called to create a response for a GET.
  */
type Finder interface {
	/**
	  * Returns either a json object representing the objects values or an error.  This will be
	  * called for a URL like /foo/127 with 127 converted to an int and passed as id.  
	  * @id will be non-negative
    * @param headers is a map from header name to value (not values, as in HTTP)
    * @param queryParams, ala (?foo=bar) is a map from query parameter name (foo) to value (bar)
	  */
	Find(id int32, headers map[string]string, queryParams map[string]string) (string,*Error)
	/**
	  * Returns doc for, respectively: resource, headers, query params.  Returned doc strings can
	  * and should be markdown encoded.
	  */
	FindDoc() string 
	/**
	  * Returns documentation for each field of the object.  The doc for each field includes
	  * a zero-valued instance of the
	  */
	FindFields() map[string]*FieldDoc
}




