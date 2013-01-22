package seven5

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const MAX_FORM_SIZE = 16 * 1024

//NewRawDispatcher is the lower-level interface to creating a RawDispatcher.  Applications only
//need this function if they wish to change the encode/decode strategy for wire objects, the
//cookie handling, or the session handling.  The prefix is used to tell the dispatcher where
//it is mounted so it can "strip" this prefix from any URL paths it is decoding.  This should
//be "" if the dispatcher is mounted at /; it should not end in a /.
func NewRawDispatcher(enc Encoder, dec Decoder, cm CookieMapper, sm SessionManager, a Authorizer,
	hold TypeHolder, prefix string) *RawDispatcher {
	return &RawDispatcher{
		Res:        make(map[string]*restObj),
		Enc:        enc,
		Dec:        dec,
		CookieMap:  cm,
		SessionMgr: sm,
		Auth:       a,
		Holder:     hold,
		Prefix:     prefix,
	}
}

//RawDispatcher is the "parent" type of dispatchers that understand REST.  It has many 
//"hooks" that allow other types to override particular behaviors.
type RawDispatcher struct {
	Res        map[string]*restObj
	Enc        Encoder
	Dec        Decoder
	CookieMap  CookieMapper
	SessionMgr SessionManager
	Auth       Authorizer
	Prefix     string
	Holder     TypeHolder
}

//ResourceSeparate adds a resource type to this dispatcher with each of the Rest methods 
//individually specified.  The name should be singular and the example should an example
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
//in so long as it meets the interface RestAll.  Resource name must be singular and will be
//converted to all lowercase.  The example wire type's fields must be public and must all be
//types definde by seven5.
func (self *RawDispatcher) Resource(name string, wireExample interface{}, r RestAll) {
	self.ResourceSeparate(name, wireExample, r, r, r, r, r)
}

//SendHook is called to encode and write the object provided onto the output via the response
//writer.  The last parameter if not "" is assumed to be a location header.  If the location
//parameter is provided, then the response code is "Created" otherwise "OK" is returned.
//SendHook calls the encoder stored within the dispatcher for the encoding of the object
//into a sequence of bytes for transmission.
func (self *RawDispatcher) SendHook(d *restObj, w http.ResponseWriter, i interface{}, location string) {
	if err := self.verifyReturnType(d, i); err != nil {
		http.Error(w, fmt.Sprintf("%s", err), http.StatusExpectationFailed)
		return
	}
	encoded, err := self.Enc.Encode(i, true)
	if err != nil {
		http.Error(w, fmt.Sprintf("unable to encode: %s", err), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "text/json")
	if location != "" {
		w.Header().Add("Location", location)
		w.WriteHeader(http.StatusCreated)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	_, err = w.Write([]byte(encoded))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to write to client connection: %s\n", err)
	}
}

//BundleHook is called to create the bundle of parameters from the request. It often will be
//using cookies and sessions to compute the bundle.  Note that the ResponseWriter is passed
//here but the BundleHook _must_ be careful to not force it out the server--it should only
//add headers.
func (self *RawDispatcher) BundleHook(w http.ResponseWriter, r *http.Request) (PBundle, error) {
	var session Session
	if self.CookieMap != nil {
		var err error
		id, err := self.CookieMap.Value(r)
		if err != nil && err != NO_SUCH_COOKIE {
			return nil, err
		}
		var find_err error
		if self.SessionMgr != nil {
			session, find_err = self.SessionMgr.Find(id)
			if find_err != nil {
				return nil, find_err
			}
			if session == nil && err != NO_SUCH_COOKIE {
				fmt.Printf("dropping cookie, can't match it to a session\n")
				self.CookieMap.RemoveCookie(w)
			}
		}
	}
	pb, err := NewSimplePBundle(r, session)
	if err != nil {
		return nil, err
	}
	return pb, nil
}

//BodyHook is called to create a wire object of the appopriate type and fill in the values
//in that object from the request body.  BodyHook calls the decoder inside the dispatcher
//take the bytes provided by the body and initialize the object that is ultimately returned.
func (self *RawDispatcher) BodyHook(r *http.Request, obj *restObj) (interface{}, error) {
	limitedData := make([]byte, MAX_FORM_SIZE)
	curr := 0
	gotEof := false
	for curr < len(limitedData) {
		n, err := r.Body.Read(limitedData[curr:])
		if err != nil && err == io.EOF {
			gotEof = true
			break
		}
		if err != nil {
			return nil, err
		}
		curr += n
	}
	//if curr==0 then we are done because there is no body
	if curr == 0 {
		return nil, nil
	}
	if !gotEof {
		return nil, errors.New(fmt.Sprintf("Body is too large! max is %d", MAX_FORM_SIZE))
	}
	//we have a body of data, need to decode it... first allocate one
	wireObj := reflect.New(obj.t)
	if err := self.Dec.Decode(limitedData[:curr], wireObj.Interface()); err != nil {
		return nil, err
	}

	return wireObj.Interface(), nil
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
	bundle, err := self.BundleHook(w, r)
	if err != nil {
		http.Error(w, fmt.Sprintf("can't create or destroy session:%s", err), http.StatusInternalServerError)
		return nil
	}

	//pull anything from the body that's there
	body, err := self.BodyHook(r, d)
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
				self.SendHook(d, w, result, "")
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
				self.SendHook(d, w, result, "")
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
			self.SendHook(d, w, result, self.location(d, result))
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
				self.SendHook(d, w, result, "")
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
				self.SendHook(d, w, result, "")
			}
		}
		return nil
	}
	panic("should not be able to reach here, probably bad method? from bad client?")
}

func (self *RawDispatcher) verifyReturnType(obj *restObj, w interface{}) error {
	if w == nil {
		return nil
	}
	p := reflect.TypeOf(w)
	if p.Kind() != reflect.Ptr {
		//could be a slice of these pointers
		if p.Kind() != reflect.Slice {
			return errors.New(fmt.Sprintf("Marshalling problem: expected a pointer/slice type but got a %v", p.Kind()))
		}
		//check that the _inner_ type is a pointer
		p = p.Elem()
		if p.Kind() != reflect.Ptr {
			return errors.New(fmt.Sprintf("Marshalling problem: expected a slice of point type but got slice of %v", p.Kind()))
		}
	}
	e := p.Elem()
	if e != obj.t {
		return errors.New(fmt.Sprintf("Marshalling problem: expected pointer to %v but got pointer to %v",
			obj.t, e))
	}
	return nil
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

