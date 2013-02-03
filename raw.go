package seven5

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

const MAX_FORM_SIZE = 16 * 1024

//NewRawDispatcher is the lower-level interface to creating a RawDispatcher.  Applications only
//need this function if they wish to substitute their own implementation for some portion of the
//handling of rest resources.  The prefix is used to tell the dispatcher where
//it is mounted so it can "strip" this prefix from any URL paths it is decoding.  This should
//be "" if the dispatcher is mounted at /; it should not end in a / or the entire world
//will come to a fiery end.
func NewRawDispatcher(io IOHook, sm SessionManager, a Authorizer,
	hold TypeHolder, prefix string) *RawDispatcher {
	return &RawDispatcher{
		Res:        make(map[string]*restObj),
		IO:         io,
		SessionMgr: sm,
		Auth:       a,
		Holder:     hold,
		Prefix:     prefix,
	}
}

//RawDispatcher is the "parent" type of dispatchers that understand REST.   This class
//is actually broken into pieces so that parts of its implementation may be changed
//by applications.
type RawDispatcher struct {
	Res        map[string]*restObj
	IO         IOHook
	SessionMgr SessionManager
	Auth       Authorizer
	Prefix     string
	Holder     TypeHolder
}

//ResourceSeparate adds a resource type to this dispatcher with each of the Rest methods 
//individually specified.  The name should be singular and camel case. The example should an example
//of the wire type to be marshalled, unmarshalled.
func (self *RawDispatcher) ResourceSeparate(name string, wireExample interface{}, index RestIndex,
	find RestFind, post RestPost, put RestPut, del RestDelete) {

	t := reflect.TypeOf(wireExample)
	if t.Kind() != reflect.Ptr {
		panic("wire example is not a pointer (should be a pointer to a struct)")
	}
	under := t.Elem()
	if under.Kind() != reflect.Struct {
		panic("wire example is not a pointer to a struct (but is a pointer)")
	}

	self.Add(name,wireExample)

	obj := &restObj{
		t:     under,
		name:  name,
		index: index,
		find:  find,
		del:   del,
		post:  post,
		put:   put,
	}
	self.Res[strings.ToLower(name)] = obj
}

//Resource is the shorter form of ResourceSeparate that allows you to pass a single resource
//in so long as it meets the interface RestAll.  Resource name must be singular and camel case and will be
//converted to all lowercase for use as a url.  The example wire type's fields must be public and must all be
//types definde by seven5.
func (self *RawDispatcher) Resource(dartClassname string, wireExample interface{}, r RestAll) {
	self.ResourceSeparate(dartClassname, wireExample, r, r, r, r, r)
}

//Rez is the really short form for adding a resource. It assumes that the dart classname is
//the same as the wire type and that the resource supports RestAll.
func (self *RawDispatcher) Rez(wireExample interface{}, r RestAll) {
	long:=reflect.TypeOf(wireExample).String()
	pieces:=strings.Split(long,".")
	dartName:=pieces[len(pieces)-1]
	self.Resource(dartName, wireExample, r)
}

//Dispatch is the entry point for the dispatcher.  Most types will want to leave this method
//intact (don't override) and instead override particular hooks to add/modify particular 
//functionality.
func (self *RawDispatcher) Dispatch(mux *ServeMux, w http.ResponseWriter, r *http.Request) *ServeMux {

	//find the resource and, if present, the id
	matched, id, d := self.resolve(r.URL.Path)
	if matched == "" {
		//typically trips the error dispatcher
		http.NotFound(w, r)
		return nil
	}
	method := strings.ToUpper(r.Method)

	//compute the parameter bundle
	bundle, err := self.IO.BundleHook(w, r, self.SessionMgr)
	if err != nil {
		http.Error(w, fmt.Sprintf("can't create or destroy session:%s", err), http.StatusInternalServerError)
		return nil
	}

	//pull anything from the body that's there
	body, err := self.IO.BodyHook(r, d)
	if err != nil {
		http.Error(w, fmt.Sprintf("badly formed body data: %s", err), http.StatusBadRequest)
		return nil
	}

	//parse the id value
	num := Id(-90191) //signal value in case it ever should end up getting used
	if len(id) > 0 {
		n, errMessage := ParseId(id)
		num = Id(n)
		if errMessage != "" {
			//typically trips the error dispatcher
			http.Error(w, fmt.Sprintf("Bad request (id): %s", errMessage), http.StatusBadRequest)
			return nil
		}
	}
	switch method {
	case "GET":
		//TWO FLAVORS OF GET: INDEXER OR FINDER?
		if len(id) == 0 { //INDEXER
			if d.index == nil {
				//typically trips the error dispatcher
				http.Error(w, "Not implemented (INDEX)", http.StatusNotImplemented)
				return nil
			}
			if self.Auth != nil && !self.Auth.Index(d, bundle) {
				//typically trips the error dispatcher
				http.Error(w, "Not authorized (INDEX)", http.StatusUnauthorized)
				return nil
			}
			result, err := d.index.Index(bundle)
			if err != nil {
				http.Error(w, fmt.Sprintf("Internal error on Index: %s", err), http.StatusInternalServerError)
			} else {
				//go through encoding
				self.IO.SendHook(d, w, bundle, result, "")
			}
			return nil
		} else { //FINDER
			if d.find == nil {
				//typically trips the error dispatcher
				http.Error(w, "Not implemented (FIND)", http.StatusNotImplemented)
				return nil
			}
			if self.Auth != nil && !self.Auth.Find(d, num, bundle) {
				//typically trips the error dispatcher
				http.Error(w, "Not authorized (FIND)", http.StatusUnauthorized)
				return nil
			}
			result, err := d.find.Find(num, bundle)
			if err != nil {
				http.Error(w, fmt.Sprintf("Internal error on Find: %s", err), http.StatusInternalServerError)
			} else {
				self.IO.SendHook(d, w, bundle, result, "")
			}
			return nil
		}
	case "POST":
		if id != "" {
			http.Error(w, "can't POST to a particular resource, did you mean PUT?", http.StatusBadRequest)
			return nil
		}
		if d.post == nil {
			http.Error(w, "Not implemented (POST)", http.StatusNotImplemented)
			return nil
		}
		if self.Auth != nil && !self.Auth.Post(d, bundle) {
			http.Error(w, "Not authorized (POST)", http.StatusUnauthorized)
			return nil
		}
		result, err := d.post.Post(body, bundle)
		if err != nil {
			http.Error(w, fmt.Sprintf("Internal error on Post: %s", err), http.StatusInternalServerError)
		} else {
			self.IO.SendHook(d, w, bundle, result, self.location(d, result))
		}
		return nil
	case "PUT", "DELETE":
		if id == "" {
			http.Error(w, fmt.Sprintf("%s requires a resource id", method), http.StatusBadRequest)
			return nil
		}
		if method == "PUT" {
			if d.put == nil {
				http.Error(w, "Not implemented (PUT)", http.StatusNotImplemented)
				return nil
			}
			if self.Auth != nil && !self.Auth.Put(d, num, bundle) {
				http.Error(w, "Not authorized (PUT)", http.StatusUnauthorized)
				return nil
			}
			result, err := d.put.Put(num, body, bundle)
			if err != nil {
				http.Error(w, fmt.Sprintf("Internal error on Put: %s", err), http.StatusInternalServerError)
			} else {
				self.IO.SendHook(d, w, bundle, result, "")
			}
		} else {
			if d.del == nil {
				http.Error(w, "Not implemented (DELETE)", http.StatusNotImplemented)
				return nil
			}
			if self.Auth != nil && !self.Auth.Delete(d, num, bundle) {
				http.Error(w, "Not authorized (DELETE)", http.StatusUnauthorized)
				return nil
			}
			result, err := d.del.Delete(num, bundle)
			if err != nil {
				http.Error(w, fmt.Sprintf("Internal error on Delete: %s", err), http.StatusInternalServerError)
			} else {
				self.IO.SendHook(d, w, bundle, result, "")
			}
		}
		return nil
	}
	panic("should not be able to reach here, probably bad method? from bad client?")
}


//Location computes the url path to the object provided
func (self *RawDispatcher) location(obj *restObj, i interface{}) string {
	//we should have already checked that this object is a pointer to a struct with an Id field
	//in the "right place" and "right type"
	result := self.Prefix + "/" + obj.name

	p := reflect.ValueOf(i)
	if p.Kind() != reflect.Ptr {
		panic("unexpected kind for p")
	}
	e := p.Elem()
	if e.Kind() != reflect.Struct {
		panic("unexpected kind for e")
	}
	f := e.FieldByName("Id")
	id := f.Int()

	return fmt.Sprintf("%s/%d", result, id)
}

//resolve is used to find the matching resource for a particular request.  It returns the match
//and the resource matched.  If no match is found it returns nil for the type.  resolve does not check
//that the resulting object is suitable for any purpose, only that it matches.
func (self *RawDispatcher) resolve(rawPath string) (string, string, *restObj) {
	path := rawPath
	if strings.HasSuffix(path,"/") && path!="/" {
		path=path[0:len(path)-1]
	}
	pre := self.Prefix + "/"
	if self.Prefix != "" {
		if !strings.HasPrefix(path, pre) {
			panic(fmt.Sprintf("expected prefix %s on the URL path but not found on: %s", pre, path))
		}
		path = path[len(pre):]
	}
	d, ok := self.Res[path]
	var id string
	result := path
	if !ok {
		i := strings.LastIndex(path, "/")
		if i == -1 {
			return "", "", nil
		}
		id = path[i+1:]
		var uriPathParent string
		uriPathParent = path[:i]
		d, ok = self.Res[uriPathParent]
		if !ok {
			return "", "", nil
		}
		result = uriPathParent
	}
	return result, id, d
}

//ParseId returns the id contained in a string or an error message about why the id is bad.
func ParseId(candidate string) (Id, string) {
	var num int64
	var err error
	if num, err = strconv.ParseInt(candidate, 10, 64); err != nil {
		return Id(0), fmt.Sprintf("resource ids must be non-negative integers (was %s): %s", candidate, err)
	}
	return Id(num), ""
}

//Add is required by the TypeHolder protocol.  Delegated into the TypeHolder passed at creation time.
func (self *RawDispatcher) Add(name string,wireType interface{}) {
	self.Holder.Add(name,wireType)
}

//All is required by the TypeHolder protocol. Delegated into the TypeHolder passed at creation time.
func (self *RawDispatcher) All() []*FieldDescription {
	return self.Holder.All()
}

