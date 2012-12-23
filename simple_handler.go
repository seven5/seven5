package seven5

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

const MAX_FORM_SIZE = 16 * 1024

//SimpleHandler is the default implementation of the Handler interface that ignores multiple values for
//headers and query params because these are both rare and error-prone.  All resources need to be
//added to the Handler before it starts serving real HTTP requests.
type SimpleHandler struct {
	//connection to the http layer
	mux *http.ServeMux
	//doc handling
	dispatch map[string]*Dispatch
	//the cookie mapper, if you want one
	cookieMapper CookieMapper
}

//NewSimpleHandler creates a new SimpleHandler with an empty URL space.  If you pass a cookie
//mapper, it will be used to "find" the sessions from requests.  Note that this cookie mapper
//is ONLY used for resource requests.
func NewSimpleHandler(cm CookieMapper) *SimpleHandler {
	return &SimpleHandler{
		mux:            http.NewServeMux(),
		dispatch:       make(map[string]*Dispatch),
		cookieMapper: cm,
	}
}

//ServeMux returns the underlying ServeMux that can be used to register additional HTTP
//resources (paths) with this object.
func (self *SimpleHandler) ServeMux() *http.ServeMux {
	return self.mux
}

//AddExplicitResourceMethods maps the (singular) resourceName into the url space.  The name should not include 
//any slashes or spaces as this will trigger armageddon and destroy all life on this planet.  The second
//parameter is the underlying type that is passed over the wire; it must be a struct or a pointer to one.
//This second parameter must have a field of type seven5.Id called Id.  All other fields must also be
//public and the types of these fields must be from the seven5 package (since the types must be flattenable
//nicely to JSON and mapped directly to Go, Dart, and SQL types). 
//
//The latter parameters are seven5 interfaces that represent each of the REST methods that are exposed to
//clients.  Any of these can be nil; nil values cause the server to respond with 501 (Not Implemented).  
//This method is functionally equivalent to a call on AddResource but allows each of the REST interfaces
//to be explicitly tied to an implementation.
//
//This method is public in case you want to access it instead of the AddResource() method visible through
//the Handler interface.
func (self *SimpleHandler) AddExplicitResourceMethods(resourceName string, r interface{},
	indexer Indexer, finder Finder, poster Poster, puter Puter, deleter Deleter) {

	if resourceName == "" || strings.Index(resourceName, " ") != -1 || strings.Index(resourceName, "/") != -1 {
		panic(fmt.Sprintf("bad resource name: '%s', no spaces or slashes allowed", resourceName))
	}

	d := NewDispatch(resourceName, r, indexer, finder, poster, puter, deleter)

	withSlashes := fmt.Sprintf("/%s/", strings.ToLower(resourceName))
	self.mux.Handle(withSlashes, self)
	self.dispatch[withSlashes] = d
}

//AddResourceByName maps the (singular) resourceName into the url space.  The name should not include 
//any slashes or spaces as this will trigger armageddon and destroy all life on this planet.  The second
//parameter is the underlying type that is passed over the wire; it must be a struct or a pointer to one.
//This second parameter must have a field of type seven5.Id called Id.  All other fields must also be
//public and the types of these fields must be from the seven5 package (since the types must be flattenable
//nicely to JSON and mapped directly to Go, Dart, and SQL types). 
//
//The last parameters must implement 0 or more of the REST resource interfaces of seven5 such as Indexer,
//Finder, or Poster.  If any of the interfaces are not implemented by the object, the server will
//return 501 (Not Implemented).
func (self *SimpleHandler) AddResourceByName(resourceName string, r interface{}, resourceImpl interface{}) {
	indexer, _ := resourceImpl.(Indexer)
	finder, _ := resourceImpl.(Finder)
	poster, _ := resourceImpl.(Poster)
	puter, _ := resourceImpl.(Puter)
	deleter, _ := resourceImpl.(Deleter)

	self.AddExplicitResourceMethods(resourceName, r, indexer, finder, poster, puter, deleter)
}

//AddResource maps a resource into the url space.  The name of the resource is derived from the name
//of the struct (which is converted to lower case and should be singular).  The first parameter must have a 
//field of type seven5.Id called Id.  All other fields must also be public and the types of these fields must 
//be from the seven5 package (since the types must be flattenable nicely to JSON and mapped directly to 
//Go, Dart, and SQL types). 
//
//The second parameters must implement 0 or more of the REST resource interfaces of seven5 such as Indexer,
//Finder, or Poster.  If any of the interfaces are not implemented by the object, the server will
//return 501 (Not Implemented).
func (self *SimpleHandler) AddResource(overTheWireSingular interface{}, implementation interface{}) {
	t := reflect.TypeOf(overTheWireSingular)
	if t.Kind() == reflect.Struct {
		self.AddResourceByName(strings.ToLower(t.Name()), overTheWireSingular, implementation)
	} else if t.Kind() == reflect.Ptr {
		self.AddResourceByName(strings.ToLower(t.Elem().Name()), overTheWireSingular, implementation)
	} else {
		panic(fmt.Sprintf("Type of %s must be struct or pointer to struct!", overTheWireSingular))
	}
}

//ServeHTTP allows this object to act like an http.Handler. ServeHTTP data is passed to Dispatch
//after some minimal processing.  This is not used in tests, only when on a real network.
func (self *SimpleHandler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	hdr := ToSimpleMap(req.Header)
	defer req.Body.Close()

	if err := req.ParseForm(); err != nil {
		http.Error(writer, fmt.Sprintf("can't parse form data:%s", err), http.StatusBadRequest)
		return
	}

	qparams := ToSimpleMap(map[string][]string(req.Form))

	limitedData := make([]byte, MAX_FORM_SIZE)
	curr := 0
	for curr < len(limitedData) {
		n, err := req.Body.Read(limitedData[curr:])
		if err != nil && err == io.EOF {
			break
		}
		if err != nil {
			http.Error(writer, fmt.Sprintf("can't read form data:%s", err), http.StatusInternalServerError)
			return
		}
		curr += n
	}

	//if available, get a session
	var session Session
	if self.cookieMapper != nil {
		var err error
		session, err = self.cookieMapper.Session(req)
		if err!=nil && err!=NO_SUCH_COOKIE {
			http.Error(writer, fmt.Sprintf("can't create session:%s", err), http.StatusInternalServerError)
			return
		}
		if session == nil {
			fmt.Printf("dropping cookie, can't match it to a session\n")
			self.cookieMapper.RemoveCookie(writer)
		}
	}

	json, err := self.Dispatch(req.Method, req.URL.Path, hdr, qparams,
		string(limitedData[0:curr]), session)
	if err != nil && err.StatusCode == http.StatusNotFound {
		self.mux.ServeHTTP(writer, req)
	} else {
		DumpOutput(writer, json, err)
	}
}

//Dispatch does the dirty work of finding a resource and calling it.
//It returns the value from the correct rest-level function or an error.
//It generates some errors itself if, for example a 404 or 501 is needed.
//I borrowed lots of ideas and inspiration from "github.com/Kissaki/rest2go"
func (self *SimpleHandler) Dispatch(method string, uriPath string, header map[string]string,
	queryParams map[string]string, body string, session Session) (string, *Error) {

	matched, id, d := self.resolve(uriPath)
	if matched == "" {
		return NotFound()
	}
	method = strings.ToUpper(method)
	switch method {
	case "GET":
		if len(id) == 0 {
			if d.Index != nil {
				allowReader, ok := d.Index.(AllowReader)
				if ok {
					if !allowReader.AllowRead(session) {
						return NotAuthorized()
					}
				}
				return d.Index.Index(header, queryParams, session)
			} else {
				//log.Printf("%T isn't an Indexer, returning NotImplemented", someResource)
				return NotImplemented()
			}
		} else {
			// Find by ID
			num, errMessage := ParseId(id)
			if errMessage != "" {
				return BadRequest(errMessage)
			}
			//resource id is a number, try to find it
			if d.Find != nil {
				allow, ok:=d.Index.(Allower)
				if ok {
					if !allow.Allow(Id(num),"GET",session) {
						return NotAuthorized()
					}
				}
				return d.Find.Find(Id(num), header, queryParams, session)
			} else {
				return NotImplemented()
			}
		}
	case "POST":
		if id != "" {
			return BadRequest("can't POST to a particular resource, did you mean PUT?")
		}
		if d == nil {
			return NotFound()
		}
		if d.Post == nil {
			return NotImplemented()
		}
		allowWriter, ok:=d.Post.(AllowWriter)
		if ok {
			if !allowWriter.AllowWrite(session) {
				return NotAuthorized()
			}
		}
		
		return d.Post.Post(header, queryParams, body, session)
	//these two are really similar
	case "PUT", "DELETE":
		if id == "" {
			return BadRequest(fmt.Sprintf("%s requires a resource id", method))
		}
		if d == nil {
			return NotFound()
		}
		num, errMessage := ParseId(id)
		if errMessage != "" {
			return BadRequest(errMessage)
		}
		if method == "PUT" {
			if d.Put == nil {
				return NotImplemented()
			}
			allow, ok:=d.Put.(Allower)
			if ok {
				if !allow.Allow(Id(num),"PUT",session) {
					return NotAuthorized()
				}
			}
			return d.Put.Put(num, header, queryParams, body, session)
		} else {
			if d.Delete == nil {
				return NotImplemented()
			}
			allow, ok:=d.Delete.(Allower)
			if ok {
				if !allow.Allow(Id(num),"DELETE",session) {
					return NotAuthorized()
				}
			}
			return d.Delete.Delete(num, header, queryParams, session)
		}
	}

	return "", &Error{http.StatusNotImplemented, "", "Not implemented yet"}
}

//parseId returns the id contained in a string or an error message about why the id is bad.
func ParseId(candidate string) (Id, string) {
	var num int64
	var err error
	if num, err = strconv.ParseInt(candidate, 10, 64); err != nil {
		return Id(0), fmt.Sprintf("resource ids must be non-negative integers (was %s): %s", candidate, err)
	}
	return Id(num), ""
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
func (self *SimpleHandler) APIs() []*APIDoc {
	result := []*APIDoc{}
	for k, _ := range self.dispatch {
		contains := false
		target := self.Describe(k)
		for _, i := range result {
			if i.Name == target.Name {
				contains = true
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
func (self *SimpleHandler) Describe(uriPath string) *APIDoc {
	result := &APIDoc{}
	path, _, _ := self.resolve(uriPath)

	//no such path?
	if path == "" {
		return nil
	}
	d := self.dispatch[path]
	result.Name = reflect.TypeOf(d.ResType).Name()
	result.Field = d.Field
	result.ResourceName = strings.Replace(path, "/", "", -1)

	if d.Find != nil {
		result.Find = true
		result.FindDoc = d.Find.FindDoc()
	}
	if d.Index != nil {
		result.Index = true
		result.IndexDoc = d.Index.IndexDoc()
	}
	if d.Post != nil {
		result.Post = true
		result.PostDoc = d.Post.PostDoc()
	}
	if d.Put != nil {
		result.Put = true
		result.PutDoc = d.Put.PutDoc()
	}
	if d.Delete != nil {
		result.Delete = true
		result.DeleteDoc = d.Delete.DeleteDoc()
	}
	return result
}
