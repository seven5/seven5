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
	Dispatch(method string, uriPath string, header map[string]string, queryParams map[string]string) (string, *Error)
	//AddIndexAndFind maps the singular and plural resource names into URL space. If singular
	//is not "", then GET on /singular/id calls finder.Find().  If plural is not "" then
	//GET /plural/ calls the Indexer.  r should be a struct that describes the json exchanged
	//between the client and server.  This struct should have only simple field types or
	//substructs that are similarly composed.  
	AddFindAndIndex(singular string, finder Finder, plural string, indexer Indexer, r interface{})
	//ServeHttp allows this type to be used as an http.Handler in the http.ListenAndServe method.  However,
	//all manipulations of the mapping (such as adding resources) must have been completed
	//before this object is used as an http.Handler because there are no concurrency guarantees
	//around the data structures internal to this object.
	ServeHTTP(http.ResponseWriter, *http.Request)
	//Doc generates a JSON representation of the resource, suitable for generation of documentation
	//or construction of an API on the fly.
	GenerateDoc(uriPath string)*ResourceDescription
}
