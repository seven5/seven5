package seven5

import (
	"net/http"
)
/**
  * One layer of indirection around the DefaultHandler in case somebody wants different
  * behavior at the HTTP level.  
  */
type Handler interface {
	/**
	  * This is the function that actually gets called by the go-level muxer (there is 
	  * a wrapper around the handler interface).  Handle is responsible for calling
	  * Dispatch() so you can ignore this Dispatch if you want your own.
	  */
	Handle(response http.ResponseWriter, request *http.Request)
	/**
	  * This is the function that dispatches method calls to rest level resources.
	  */
	Dispatch(method string, uriPath string, header map[string]string, queryParams map[string]string)
	/**
	  * This maps a resource into the URL space.  You can write your own version to do
	  * fancy things like put prefixes on the URLs or do versioning, etc. All calls to
	  * add resource should be completed _before_ calling ListenAndServe because there
	  * no concurrency guarantees around the mapping data structure.
	  */
	AddResource(name string, r interface{})
	/**
	  * Removes all mappings currently known.  Useful for test code.
	  */
	RemoveAllResources()
}

/**
  * The currently in use handler, typically an instance of SimpleHandler
  */
var CurrentHandler = NewSimpleHandler()

//because go-level wants a function, not an interface
func handlerWrapper(response http.ResponseWriter, request *http.Request) {
	CurrentHandler.Handle(response, request)
}