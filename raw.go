package seven5

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"unicode"
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
		ResUdid:    make(map[string]*restObjUdid),
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
	ResUdid    map[string]*restObjUdid
	IO         IOHook
	SessionMgr SessionManager
	Auth       Authorizer
	Prefix     string
	Holder     TypeHolder
}

func (self *RawDispatcher) validateType(example interface{}) reflect.Type {
	t := reflect.TypeOf(example)
	if t.Kind() != reflect.Ptr {
		panic("wire example is not a pointer (should be a pointer to a struct)")
	}
	under := t.Elem()
	if under.Kind() != reflect.Struct {
		panic("wire example is not a pointer to a struct (but is a pointer)")
	}
	return under
}

//ResourceSeparate adds a resource type to this dispatcher with each of the Rest methods
//individually specified.  The name should be singular and camel case. The example should an example
//of the wire type to be marshalled, unmarshalled.
func (self *RawDispatcher) ResourceSeparate(name string, wireExample interface{}, index RestIndex,
	find RestFind, post RestPost, put RestPut, del RestDelete) {

	under := self.validateType(wireExample)
	self.validateType(wireExample)
	self.Add(name, wireExample)
	obj := &restObj{
		restShared: restShared{
			t:     under,
			name:  name,
			index: index,
			post:  post,
		},
		find: find,
		del:  del,
		put:  put,
	}
	self.Res[strings.ToLower(name)] = obj
}

//ResourceSeparateUdid adds a resource type to this dispatcher with each of the RestUdid methods
//individually specified.  The name should be singular and camel case. The example should an example
//of the wire type to be marshalled, unmarshalled.  The wire type can have an Id in addition
//to a Udid field.
func (self *RawDispatcher) ResourceSeparateUdid(name string, wireExample interface{}, index RestIndex,
	find RestFindUdid, post RestPost, put RestPutUdid, del RestDeleteUdid) {

	under := self.validateType(wireExample)
	self.Add(name, wireExample)
	obj := &restObjUdid{
		restShared: restShared{
			t:     under,
			name:  name,
			index: index,
			post:  post,
		},
		find: find,
		del:  del,
		put:  put,
	}
	self.ResUdid[strings.ToLower(name)] = obj
}

//Resource is the shorter form of ResourceSeparate that allows you to pass a single resource
//in so long as it meets the interface RestAll.  Resource name must be singular and camel case and will be
//converted to all lowercase for use as a url.  The example wire type's fields must be public.
func (self *RawDispatcher) Resource(name string, wireExample interface{}, r RestAll) {
	self.ResourceSeparate(name, wireExample, r, r, r, r, r)
}

//ResourceUdid is the shorter form of ResourceSeparateUdid that allows you to pass a single resource
//in so long as it meets the interface RestAllUdid.  Resource name must be singular and camel case and will be
//converted to all lowercase for use as a url.  The example wire type's fields must be public.

func (self *RawDispatcher) ResourceUdid(name string, wireExample interface{}, r RestAllUdid) {
	self.ResourceSeparateUdid(name, wireExample, r, r, r, r, r)
}

//Rez is the really short form for adding a resource. It assumes that the name is
//the same as the wire type and that the resource supports RestAll.
func (self *RawDispatcher) Rez(wireExample interface{}, r RestAll) {
	long := reflect.TypeOf(wireExample).String()
	pieces := strings.Split(long, ".")
	name := pieces[len(pieces)-1]
	self.Resource(name, wireExample, r)
}

//Dispatch is the entry point for the dispatcher.  Most types will want to leave this method
//intact (don't override) and instead override particular hooks to add/modify particular
//functionality.
func (self *RawDispatcher) Dispatch(mux *ServeMux, w http.ResponseWriter, r *http.Request) *ServeMux {

	//find the resource and, if present, the id
	matched, id, rez, rezUdid := self.resolve(r.URL.Path)
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

	var body interface{}
	var num int64

	//pull anything from the body that's there
	if rezUdid == nil {
		body, err = self.IO.BodyHook(r, &rez.restShared)
		if err != nil {
			http.Error(w, fmt.Sprintf("badly formed body data: %s", err), http.StatusBadRequest)
			return nil
		}
	} else {
		body, err = self.IO.BodyHook(r, &rezUdid.restShared)
		if err != nil {
			http.Error(w, fmt.Sprintf("badly formed body data: %s", err), http.StatusBadRequest)
			return nil
		}
	}

	//may need to parse it as an int
	if rezUdid == nil {
		//parse the id value
		num = int64(-90191) //signal value in case it ever should end up getting used
		if len(id) > 0 {
			n, errMessage := ParseId(id)
			num = n
			if errMessage != "" {
				//typically trips the error dispatcher
				http.Error(w, fmt.Sprintf("Bad request (id): %s", errMessage), http.StatusBadRequest)
				return nil
			}
		}
	}

	switch method {
	case "GET":
		//TWO FLAVORS OF GET: INDEXER OR FINDER?
		if len(id) == 0 { //INDEXER
			if rez != nil {
				if rez.index == nil {
					//typically trips the error dispatcher
					http.Error(w, "Not implemented (INDEX)", http.StatusNotImplemented)
					return nil
				}
				if self.Auth != nil && !self.Auth.Index(&rez.restShared, bundle) {
					//typically trips the error dispatcher
					http.Error(w, "Not authorized (INDEX)", http.StatusUnauthorized)
					return nil
				}
				result, err := rez.index.Index(bundle)
				if err != nil {
					self.SendError(err, w, "Internal error on Index")
				} else {
					//go through encoding
					self.IO.SendHook(&rez.restShared, w, bundle, result, "")
				}
			} else {
				//UDID INDER
				if rezUdid.index == nil {
					//typically trips the error dispatcher
					http.Error(w, "Not implemented (INDEX, UDID)", http.StatusNotImplemented)
					return nil
				}
				if self.Auth != nil && !self.Auth.Index(&rezUdid.restShared, bundle) {
					//typically trips the error dispatcher
					http.Error(w, "Not authorized (INDEX, UDID)", http.StatusUnauthorized)
					return nil
				}
				result, err := rezUdid.index.Index(bundle)
				if err != nil {
					self.SendError(err, w, "Internal error on Index (UDID)")
				} else {
					//go through encoding
					self.IO.SendHook(&rezUdid.restShared, w, bundle, result, "")
				}
			}
			return nil
		} else { //FINDER
			if rez != nil {
				if rez.find == nil {
					//typically trips the error dispatcher
					http.Error(w, "Not implemented (FIND)", http.StatusNotImplemented)
					return nil
				}
				if self.Auth != nil && !self.Auth.Find(rez, num, bundle) {
					//typically trips the error dispatcher
					http.Error(w, "Not authorized (FIND)", http.StatusUnauthorized)
					return nil
				}
				result, err := rez.find.Find(num, bundle)
				if err != nil {
					self.SendError(err, w, "Internal error on Find")
				} else {
					self.IO.SendHook(&rez.restShared, w, bundle, result, "")
				}
				return nil
			} else {
				//UDID RESOURCE
				if rezUdid.find == nil {
					//typically trips the error dispatcher
					http.Error(w, "Not implemented (FIND,UDID)", http.StatusNotImplemented)
					return nil
				}
				if self.Auth != nil && !self.Auth.FindUdid(rezUdid, id, bundle) {
					//typically trips the error dispatcher
					http.Error(w, "Not authorized (FIND, UDID)", http.StatusUnauthorized)
					return nil
				}
				result, err := rezUdid.find.Find(id, bundle)
				if err != nil {
					self.SendError(err, w, "Internal error on Find (UDID")
				} else {
					self.IO.SendHook(&rezUdid.restShared, w, bundle, result, "")
				}
				return nil
			}
		}
	case "POST":
		if rez != nil {
			if id != "" {
				http.Error(w, "can't POST to a particular resource, did you mean PUT?", http.StatusBadRequest)
				return nil
			}
			if rez.post == nil {
				http.Error(w, "Not implemented (POST)", http.StatusNotImplemented)
				return nil
			}
			if self.Auth != nil && !self.Auth.Post(&rez.restShared, bundle) {
				http.Error(w, "Not authorized (POST)", http.StatusUnauthorized)
				return nil
			}
			result, err := rez.post.Post(body, bundle)
			if err != nil {
				self.SendError(err, w, "Internal error on Post")
			} else {
				self.IO.SendHook(&rez.restShared, w, bundle, result, self.location(rez.name, false, result))
			}
			return nil
		} else {
			//UDID POST
			if id != "" {
				http.Error(w, "can't (UDID) POST to a particular resource, did you mean PUT?", http.StatusBadRequest)
				return nil
			}
			if rezUdid.post == nil {
				http.Error(w, "Not implemented (POST, UDID)", http.StatusNotImplemented)
				return nil
			}
			if self.Auth != nil && !self.Auth.Post(&rezUdid.restShared, bundle) {
				http.Error(w, "Not authorized (POST)", http.StatusUnauthorized)
				return nil
			}
			result, err := rezUdid.post.Post(body, bundle)
			if err != nil {
				self.SendError(err, w, "Internal error on Post")
			} else {
				self.IO.SendHook(&rezUdid.restShared, w, bundle, result, self.location(rezUdid.name, true, result))
			}
			return nil

		}
	case "PUT", "DELETE":
		if id == "" {
			http.Error(w, fmt.Sprintf("%s requires a resource id or UDID", method), http.StatusBadRequest)
			return nil
		}
		if method == "PUT" {
			if rez != nil {
				if rez.put == nil {
					http.Error(w, "Not implemented (PUT)", http.StatusNotImplemented)
					return nil
				}
				if self.Auth != nil && !self.Auth.Put(rez, num, bundle) {
					http.Error(w, "Not authorized (PUT)", http.StatusUnauthorized)
					return nil
				}
				result, err := rez.put.Put(num, body, bundle)
				if err != nil {
					self.SendError(err, w, "Internal error on Put")
				} else {
					self.IO.SendHook(&rez.restShared, w, bundle, result, "")
				}
			} else {
				//PUT ON UDID
				if rezUdid.put == nil {
					http.Error(w, "Not implemented (PUT, UDID)", http.StatusNotImplemented)
					return nil
				}
				if self.Auth != nil && !self.Auth.PutUdid(rezUdid, id, bundle) {
					http.Error(w, "Not authorized (PUT, UDID)", http.StatusUnauthorized)
					return nil
				}
				result, err := rezUdid.put.Put(id, body, bundle)
				if err != nil {
					self.SendError(err, w, "Internal error on Put (UDID)")
				} else {
					self.IO.SendHook(&rezUdid.restShared, w, bundle, result, "")
				}
			}
		} else {
			if rez != nil {
				if rez.del == nil {
					http.Error(w, "Not implemented (DELETE)", http.StatusNotImplemented)
					return nil
				}
				if self.Auth != nil && !self.Auth.Delete(rez, num, bundle) {
					http.Error(w, "Not authorized (DELETE)", http.StatusUnauthorized)
					return nil
				}
				result, err := rez.del.Delete(num, bundle)
				if err != nil {
					self.SendError(err, w, "Internal error on Delete")
				} else {
					self.IO.SendHook(&rez.restShared, w, bundle, result, "")
				}
			} else {
				//UDID DELETE
				if rezUdid.del == nil {
					http.Error(w, "Not implemented (DELETE, UDID)", http.StatusNotImplemented)
					return nil
				}
				if self.Auth != nil && !self.Auth.DeleteUdid(rezUdid, id, bundle) {
					http.Error(w, "Not authorized (DELETE, UDID)", http.StatusUnauthorized)
					return nil
				}
				result, err := rezUdid.del.Delete(id, bundle)
				if err != nil {
					self.SendError(err, w, "Internal error on Delete")
				} else {
					self.IO.SendHook(&rezUdid.restShared, w, bundle, result, "")
				}
			}
		}
		return nil
	}
	log.Printf("should not be able to reach here, probably bad method? from bad client?")
	http.Error(w, "bad client behavior", http.StatusBadRequest)
	return nil
}

func (self *RawDispatcher) SendError(err error, w http.ResponseWriter, msg string) {
	ours, ok := err.(*Error)
	if !ok {
		http.Error(w, fmt.Sprintf("%s: %s", msg, err), http.StatusInternalServerError)
	} else {
		http.Error(w, ours.Msg, ours.StatusCode)
	}
}

//Location computes the url path to the object provided
func (self *RawDispatcher) location(name string, isUdid bool, i interface{}) string {
	//we should have already checked that this object is a pointer to a struct with an Id field
	//in the "right place" and "right type"
	result := self.Prefix + "/" + name

	p := reflect.ValueOf(i)
	if p.Kind() != reflect.Ptr {
		panic("unexpected kind for p")
	}
	e := p.Elem()
	if e.Kind() != reflect.Struct {
		panic("unexpected kind for e")
	}
	if !isUdid {
		f := e.FieldByName("Id")
		id := f.Int()
		return fmt.Sprintf("%s/%d", result, id)
	}
	f := e.FieldByName("Udid")
	id := f.String()
	return fmt.Sprintf("%s/%s", result, id)
}

//resolve is used to find the matching resource for a particular request.  It returns the match
//and the resource matched.  If no match is found it returns nil for the type.  resolve does not check
//that the resulting object is suitable for any purpose, only that it matches.
func (self *RawDispatcher) resolve(rawPath string) (string, string, *restObj, *restObjUdid) {
	path := rawPath
	if strings.HasSuffix(path, "/") && path != "/" {
		path = path[0 : len(path)-1]
	}
	pre := self.Prefix + "/"
	if self.Prefix != "" {
		if !strings.HasPrefix(path, pre) {
			panic(fmt.Sprintf("expected prefix %s on the URL path but not found on: %s", pre, path))
		}
		path = path[len(pre):]
	}
	//case 1: simple path to a normal resource
	rez, ok := self.Res[path]
	if ok {
		return path, "", rez, nil
	}
	//case 2: simple path to a udid resource
	rezUdid, okUdid := self.ResUdid[path]
	if okUdid {
		return path, "", nil, rezUdid
	}
	//maybe we need to split it, so look for last /...
	i := strings.LastIndex(path, "/")
	if i == -1 {
		return "", "", nil, nil
	}
	id := path[i+1:]
	var uriPathParent string
	uriPathParent = path[:i]

	//case 3, path to a resource ID (int)
	rez, ok = self.Res[uriPathParent]
	if ok {
		return uriPathParent, id, rez, nil
	}

	//case 4, path to a resource ID (UDID)
	rezUdid, ok = self.ResUdid[uriPathParent]
	if ok {
		return uriPathParent, id, nil, rezUdid
	}

	//nothing, give up
	return "", "", nil, nil

}
func normalizeUdid(raw string) string {
	if len(raw) != 36 {
		log.Printf("bad length on UDID: %d", len(raw))
		return ""
	}
	var buff bytes.Buffer
	for _, ch := range raw {
		switch ch {
		case 'a', 'b', 'c', 'd', 'e', 'f', '-', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			buff.WriteRune(ch)
		case 'A', 'B', 'C', 'D', 'E', 'F':
			buff.WriteRune(unicode.ToLower(ch))
		default:
			log.Printf("bad UDID character: %s", raw)
			return ""
		}
	}
	return buff.String()
}

//ParseId returns the id contained in a string or an error message about why the id is bad.
func ParseId(candidate string) (int64, string) {
	var num int64
	var err error
	if num, err = strconv.ParseInt(candidate, 10, 64); err != nil {
		return 0, fmt.Sprintf("resource ids must be non-negative integers (was %s): %s", candidate, err)
	}
	return num, ""
}

//Add is required by the TypeHolder protocol.  Delegated into the TypeHolder passed at creation time.
func (self *RawDispatcher) Add(name string, wireType interface{}) {
	self.Holder.Add(name, wireType)
}

//All is required by the TypeHolder protocol. Delegated into the TypeHolder passed at creation time.
func (self *RawDispatcher) All() []*FieldDescription {
	return self.Holder.All()
}
