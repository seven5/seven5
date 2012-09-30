// Copyright 2012 Captricity, Inc. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// seven5 is restful without remorse or pity.
//
// Source code and project home:
// https://github.com/captricity/seven5
//
package seven5

import (
	"net/http"
)
//One layer of indirection around the DefaultHandler in case somebody wants different
//behavior at the HTTP level.  
type Handler interface {
	//This is the function that actually gets called by the go-level muxer (there is 
	//a wrapper around the handler interface).  Handle is responsible for calling
	//Dispatch() so you can ignore this Dispatch if you want your own.
	Handle(response http.ResponseWriter, request *http.Request)
	//This is the function that dispatches method calls to rest level resources.  This is
	//where to hook tests of the back end functionality.
	Dispatch(method string, uriPath string, header map[string]string, queryParams map[string]string)
	//This maps a resource into the URL space.  You can write your own version to do
	//fancy things like put prefixes on the URLs or do versioning, etc. All calls to
	//add resource should be completed _before_ calling ListenAndServe because there
	//no concurrency guarantees around the mapping data structure.
	AddResource(name string, r interface{})
	//Removes all mappings currently known.  Useful for test code.
	RemoveAllResources()
}

//The currently in use handler, typically an instance of SimpleHandler
var CurrentHandler = NewSimpleHandler()

//because go-level wants a function, not an interface, so we wrap the CurrentHandler
func handlerWrapper(response http.ResponseWriter, request *http.Request) {
	CurrentHandler.Handle(response, request)
}