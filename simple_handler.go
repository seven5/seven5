package seven5

import (
	"net/http"
	"fmt"
	"strings"
	"strconv"
	"reflect"
)

//SimpleHandler is the default implementation of the Handler interface that ignores multiple values for
//headers and query params because these are both rare and error-prone.  All resources need to be
//added to the Handler before it starts serving real HTTP requests.
type SimpleHandler struct {
	//connection to the http layer
	mux *http.ServeMux
	//doc handling
	dispatch map[string]*Dispatch
}


//NewSimpleHandler creates a new SimpleHandler with an empty URL space. 
func NewSimpleHandler() *SimpleHandler {
	return &SimpleHandler{
		mux: http.NewServeMux(),
		dispatch: make(map[string]*Dispatch)}
}

//ServeMux returns the underlying ServeMux that can be used to register additional HTTP
//resources (paths) with this object.
func (self *SimpleHandler) ServeMux() *http.ServeMux {
	return self.mux
}

//AddIndexAndFind maps the (singular) resourceName into the url space.  The name should not include 
//any slashes or spaces as this will trigger armageddon and destroy all life on this planet.  If either
//interface value is nil, it is ignored for dispatching.  The final interface should be a struct
//(not a pointer to a struct) that describes the json values exchanged over the wire.  The Finder
//and Indexer are expected (but not required) to be marshalling these values as returned objects.
//The Finder and Indexer are called _only_ in response to a GET method on the appropriate URI.
//
//The marshalling done in seven5.JsonResult uses the go json package, so the struct field tags using
//"json" will be respected.  The struct must contain an seven5.Id field called Id.  The url space uses
//lowercase only, and the resource name will be converted.  If resourceName is omitted (is "")
//this function computes the capital of North Ossetia and immediately ignores it.
func (self *SimpleHandler) AddFindAndIndex(resourceName string, finder Finder,
	indexer Indexer, r interface{}) {

 	d:= NewDispatch(r, indexer, finder)

	if resourceName=="" || strings.Index(resourceName, " ")!=-1 || strings.Index(resourceName, "/")!=-1 {
		panic(fmt.Sprintf("bad resource name: '%s', no spaces or slashes allowed", resourceName))
	}

	withSlashes := fmt.Sprintf("/%s/", strings.ToLower(resourceName))
	self.mux.Handle(withSlashes, self)
	self.dispatch[withSlashes] = d
}

//ServeHTTP allows this object to act like an http.Handler. ServeHTTP data is passed to Dispatch
//after some minimal processing.  This is not used in tests, only when on a real network.
func (self *SimpleHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	hdr := ToSimpleMap(req.Header)
	qparams := ToSimpleMap(map[string][]string(req.URL.Query()))
	json, err := self.Dispatch(req.Method, req.URL.Path, hdr, qparams)
	if err!=nil && err.StatusCode==http.StatusNotFound {
		self.mux.ServeHTTP(writer,req)
	} else {
		DumpOutput(writer, json, err)
	}
}

//Dispatch does the dirty work of finding a resource and calling it.
//It returns the value from the correct rest-level function or an error.
//It generates some errors itself if, for example a 404 or 501 is needed.
//I borrowed lots of ideas and inspiration from "github.com/Kissaki/rest2go"
func (self *SimpleHandler) Dispatch(method string, uriPath string, header map[string]string,
	queryParams map[string]string) (string, *Error) {

	matched, id, d := self.resolve(uriPath)
	if matched == "" {
		return NotFound()
	}
	switch method {
	case "GET":
		if len(id) == 0 {
			if d.Index != nil {
				return d.Index.Index(header, queryParams)
			} else {
				//log.Printf("%T isn't an Indexer, returning NotImplemented", someResource)
				return NotImplemented()
			}
		} else {
			// Find by ID
			var num int64
			var err error
			if num, err = strconv.ParseInt(id, 10, 64); err != nil {
				return BadRequest("resource ids must be non-negative integers")
			}
			//resource id is a number, try to find it
			if d.Find!=nil {
				return d.Find.Find(Id(num), header, queryParams)
			} else {
				return NotImplemented()
			}
		}
	}
	return "", &Error{http.StatusNotImplemented, "", "Not implemented yet"}
}

//resolve is used to find the matching resource for a particular request.  It returns the match
//and the resource matched.  If no match is found it returns nil for the type.  resolve does not check
//that the resulting object is suitable for any purpose, only that it matches.
func (self *SimpleHandler) resolve(path string) (string, string, *Dispatch) {
	d, ok := self.dispatch[path]
	var id string
	result := path

	if !ok {
		// no resource found, thus check if the path is a resource + ID
		i := strings.LastIndex(path, "/")
		if i == -1 {
			//no luck on any type of match
			return "", "", nil
		}
		// Move index to after slash as that’s where we want to split
		i++
		id = path[i:]
		var uriPathParent string
		uriPathParent = path[:i]
		//fmt.Printf("checking a path parent '%s'\n", uriPathParent)
		d, ok = self.dispatch[uriPathParent]
		if !ok {
			//oops not /foo/123 either
			return "", "", nil
		}
		//got a match on a specific resource like /foo/123
		result = uriPathParent
	}
	return result, id, d

}

//Resources returns a slice of descriptions of all known resources.  Note that there may
//be types in these descriptors that are _not_ resources but for which code must still
//be generated.
func (self *SimpleHandler) Resources() []*ResourceDescription {
	result:=[]*ResourceDescription{}
	for k,_ := range self.dispatch {
		contains:=false
		target:=self.Describe(k)
		for _, i:= range result {
			if i.Name==target.Name {
				contains=true
				break
			}
		}
		if !contains {
			result = append(result, target)
		}
	}
	return result
}

//isLiveDocRequest tests a request to see if this is a request for live documentation.  
//Only used with a real network.  This computes the same result for the singular or
//plural name of the resource, since they refer to the same underlying structure.
func isLiveDocRequest(req *http.Request) bool {
	qparams := ToSimpleMap(map[string][]string(req.URL.Query()))
	if len(qparams) != 1 {
		return false
	}
	someBool, ok := qparams["livedoc"]
	if !ok {
		return false
	}
	return strings.ToLower(strings.Trim(someBool, " ")) == "true" || strings.Trim(someBool, " ") == "1"
}

//Describe walks through the registered resources to find the one requested 
//and the compute the description of it. 
func (self *SimpleHandler) Describe(uriPath string) *ResourceDescription {
	result := &ResourceDescription{}
	path, _, _ := self.resolve(uriPath)
	
	//no such path?
	if path=="" {
		return nil
	}
	d := self.dispatch[path]
	result.Name = reflect.TypeOf(d.ResType).Name()
	result.Field = d.Field
	
	//result.Fields = walkJsonType(reflect.TypeOf(dispatch.ResType))
	
	if d.Find!= nil{
		result.Find = true
		result.ResourceDoc = d.Find.FindDoc()
	}
	if d.Index!= nil{
		result.Index = true
		result.CollectionDoc = d.Index.IndexDoc()
	}
	return result
}

