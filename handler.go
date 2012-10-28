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

//One layer of indirection around the SimpleHandler in case somebody wants different
//behavior at the HTTP level.  
type Handler interface {
	//Dispatch is the function that dispatches method calls to rest level resources.  This is
	//where to hook tests of the back end functionality because it does not have dependencies
	//on the network.
	Dispatch(method string, uriPath string, header map[string]string, queryParams map[string]string, body string) (string, *Error)
	//AddResourceByName maps the (singular) resourceName into the url space.  The name should not include 
	//any slashes or spaces as this will trigger armageddon and destroy all life on this planet.  The second
	//parameter is the underlying type that is passed over the wire; it must be a struct or a pointer to one.
	//This second parameter must have a field of type seven5.Id called Id.  All other fields must also be
	//public and the types of these fields must be from the seven5 package (since the types must be flattenable
	//nicely to JSON and mapped directly to Go, Dart, and SQL types). 
	//
	//The last parameters must implement 0 or more of the REST resource interfaces of seven5 such as Indexer,
	//Finder, or Poster.  If any of the interfaces are not implemented by the object, the server will
	//return 502 (Not Implemented).
	AddResourceByName(singular string, overTheWire interface{}, implementation interface{})
	//AddResource maps a resource into the url space.  The name of the resource is derived from the name
	//of the struct or pointer to struct (which is converted to lower case and should be singular) provided as
	//the first parameter.  The first parameter must have a 
	//field of type seven5.Id called Id.  All other fields must also be public and the types of these fields must 
	//be from the seven5 package (since the types must be flattenable nicely to JSON and mapped directly to 
	//Go, Dart, and SQL types). 
	//
	//The second parameters must implement 0 or more of the REST resource interfaces of seven5 such as Indexer,
	//Finder, or Poster.  If any of the interfaces are not implemented by the object, the server will
	//return 502 (Not Implemented).
	AddResource(overTheWireSingular interface{}, implementation interface{})
	//ServeHttp allows this type to be used as an http.Handler in the http.ListenAndServe method.  
	//However, all manipulations of the mapping (such as adding resources) must have been completed
	//before this object is used as an http.Handler because there are no concurrency guarantees
	//around the data structures internal to this object.
	ServeHTTP(http.ResponseWriter, *http.Request)
	//Describe generates a structure that describes the resource, suitable for generation of 
	//documentation or construction of an API.
	Describe(uriPath string) *ResourceDescription
	//ServerMux returns the underlying server multiplexer this object is based on.  If you want
	//to add other resources to the Handler you can do this by adding them to the returned mux.
	//Because this object is going to usually be "invoked" with ServeHTTP(), adding of resources
	//must be done before the call ServeHTTP to function properly.
	ServeMux() *http.ServeMux
	//Resources returns a slice of *ResourceDescription that describe all the resources known
	//to the handler.  This is useful during code generation.
	Resources() []*ResourceDescription
}
