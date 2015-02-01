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
func NewRawDispatcher(io IOHook, sm SessionManager, a Authorizer, prefix string) *RawDispatcher {
	return &RawDispatcher{
		Root:       NewRestNode(),
		IO:         io,
		SessionMgr: sm,
		Auth:       a,
		Prefix:     prefix,
	}
}

//RestNodes are tree nodes in the tree of rest resources.  A RestNode encodes
//all the items that are directly reachable from this point.  This type is
//not of interest to those not implementing their own dispatching.
type RestNode struct {
	Res          map[string]*restObj
	ResUdid      map[string]*restObjUdid
	Children     map[string]*RestNode
	ChildrenUdid map[string]*RestNode
}

//NewRestNode creates a new, empty rest node.
func NewRestNode() *RestNode {
	return &RestNode{
		Res:          make(map[string]*restObj),
		ResUdid:      make(map[string]*restObjUdid),
		Children:     make(map[string]*RestNode),
		ChildrenUdid: make(map[string]*RestNode),
	}
}

//RawDispatcher is the "parent" type of dispatchers that understand REST.   This class
//is actually broken into pieces so that parts of its implementation may be changed
//by applications.
type RawDispatcher struct {
	Root       *RestNode
	IO         IOHook
	SessionMgr SessionManager
	Auth       Authorizer
	Prefix     string
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
	return t
}

//AddResourceSeparate adds a resource to a given rest node, in a way parallel
//to ResourceSeparate.
func (self *RawDispatcher) AddResourceSeparate(node *RestNode, name string, wireExample interface{}, index RestIndex,
	find RestFind, post RestPost, put RestPut, del RestDelete) {

	t := self.validateType(wireExample)
	obj := &restObj{
		restShared: restShared{
			typ:   t,
			name:  name,
			index: index,
			post:  post,
		},
		find: find,
		del:  del,
		put:  put,
	}
	node.Res[strings.ToLower(name)] = obj
}

//ResourceSeparate adds a resource type to this dispatcher with each of the Rest methods
//individually specified.  The name should be singular and camel case. The example should an example
//of the wire type to be marshalled, unmarshalled. This is just a wrapper around adding a
//resource at the top (root) level of the dispatcher with AddResourceSeparate.
func (self *RawDispatcher) ResourceSeparate(name string, wireExample interface{}, index RestIndex,
	find RestFind, post RestPost, put RestPut, del RestDelete) {
	self.AddResourceSeparate(self.Root, name, wireExample, index, find, post, put, del)
}

//ResourceSeparateUdid adds a resource type to this dispatcher with each of the RestUdid methods
//individually specified.  The name should be singular and camel case. The example should an example
//of the wire type to be marshalled, unmarshalled.  The wire type can have an Id in addition
//to a Udid field.
func (self *RawDispatcher) ResourceSeparateUdid(name string, wireExample interface{}, index RestIndex,
	find RestFindUdid, post RestPost, put RestPutUdid, del RestDeleteUdid) {

	self.AddResourceSeparateUdid(self.Root, name, wireExample, index, find, post, put, del)
}

//AddResourceSeparateUdid adds a resource to a given rest node, in a way parallel
//to ResourceSeparateUdid.
func (self *RawDispatcher) AddResourceSeparateUdid(node *RestNode, name string, wireExample interface{}, index RestIndex,
	find RestFindUdid, post RestPost, put RestPutUdid, del RestDeleteUdid) {
	t := self.validateType(wireExample)
	obj := &restObjUdid{
		restShared: restShared{
			typ:   t,
			name:  name,
			index: index,
			post:  post,
		},
		find: find,
		del:  del,
		put:  put,
	}
	node.ResUdid[strings.ToLower(name)] = obj
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
	self.Resource(exampleTypeToName(wireExample), wireExample, r)
}

//RezUdid is the really short form for adding a resource based on Udid. It
//assumes that the name is the same as the wire type and that the resource
//supports RestAllUdid.
func (self *RawDispatcher) RezUdid(wireExample interface{}, r RestAllUdid) {
	self.ResourceUdid(exampleTypeToName(wireExample), wireExample, r)
}

func exampleTypeToName(wireExample interface{}) string {
	long := reflect.TypeOf(wireExample).String()
	pieces := strings.Split(long, ".")
	return pieces[len(pieces)-1]
}

func sanityCheckParentWireExample(parentWire interface{}) reflect.Type {
	t := reflect.TypeOf(parentWire)
	if t.Kind() != reflect.Ptr {
		panic("parent wire example is not a pointer (should be a pointer to a struct)")
	}
	under := t.Elem()
	if under.Kind() != reflect.Struct {
		panic("parent wire example is not a pointer to struct (but is a pointer)")
	}
	return t
}

//SubResource is for adding a subresource, analagous to Resource
//You must provide the name of the wire type, lower case and singular, to be used
//with this resource.  This call panics if the provided parent wire example cannot be located because this indicates that
//the program is misconfigured and cannot work.
func (self *RawDispatcher) SubResource(parentWire interface{},
	subresourcename string, wireExample interface{}, index RestIndex, find RestFind, post RestPost, put RestPut, del RestDelete) {

	parent := self.FindWireType(sanityCheckParentWireExample(parentWire), self.Root)
	if parent == nil {
		panic(fmt.Sprintf("unable to find wire type (parent) %T", parentWire))
	}
	child := NewRestNode()
	parent.Children[subresourcename] = child
	self.AddResourceSeparate(child, subresourcename, wireExample,
		index, find, post, put, del)
}

//SubResourceSeparate is for adding a subresource, analagous to ResourceSeparate.
//It assumes that the subresource name is the same as the wire type.
//This call panics if the provided parent wire example cannot be located because this indicates that
//the program is misconfigured and cannot work.
func (self *RawDispatcher) SubResourceSeparate(parentWire interface{}, wireExample interface{}, index RestIndex,
	find RestFind, post RestPost, put RestPut, del RestDelete) {

	self.SubResource(parentWire, strings.ToLower(exampleTypeToName(wireExample)),
		wireExample, index, find, post, put, del)
}

//SubResourceUdid is for adding a subresource udid, analagous to ResourceUdid.
//You must provide the subresource name, singular and lower case.
//If the provided parent wire example cannot be located because this indicates that
//the program is misconfigured and cannot work.
func (self *RawDispatcher) SubResourceUdid(parentWire interface{}, subresourcename string, wireExample interface{}, index RestIndex,
	find RestFindUdid, post RestPost, put RestPutUdid, del RestDeleteUdid) {

	parent := self.FindWireType(sanityCheckParentWireExample(parentWire), self.Root)
	if parent == nil {
		panic(fmt.Sprintf("unable to find wire type (parent) %T", parentWire))
	}
	child := NewRestNode()
	parent.ChildrenUdid[strings.ToLower(exampleTypeToName(wireExample))] = child
	self.AddResourceSeparateUdid(child, subresourcename, wireExample, index,
		find, post, put, del)
}

//SubResourceSeparateUdid is for adding a subresource udid, analagous to ResourceSeparateUdid.
//It assumes that the subresource name is the same as the wire type.
//If the provided parent wire example cannot be located because this indicates that
//the program is misconfigured and cannot work.
func (self *RawDispatcher) SubResourceSeparateUdid(parentWire interface{}, wireExample interface{}, index RestIndex,
	find RestFindUdid, post RestPost, put RestPutUdid, del RestDeleteUdid) {
	self.SubResourceUdid(parentWire, strings.ToLower(exampleTypeToName(wireExample)),
		wireExample, index, find, post, put, del)
}

//FindWireType searches the tree of rest resources trying to find one that has the
//given type as a target. This is only of interest to dispatch implementors.
func (self *RawDispatcher) FindWireType(target reflect.Type, curr *RestNode) *RestNode {
	for _, v := range curr.Res {
		if v.typ == target {
			return curr
		}
	}
	for _, v := range curr.ResUdid {
		if v.typ == target {
			return curr
		}
	}

	//we now need to recurse into children, does DFS
	for _, child := range curr.Children {
		if node := self.FindWireType(target, child); node != nil {
			return node
		}
	}
	for _, child := range curr.ChildrenUdid {
		if node := self.FindWireType(target, child); node != nil {
			return node
		}
	}
	//loser
	return nil
}

//Dispatch is the entry point for the dispatcher.  Most types will want to leave this method
//intact (don't override) and instead override particular hooks to add/modify particular
//functionality.
func (self *RawDispatcher) Dispatch(mux *ServeMux, w http.ResponseWriter, r *http.Request) *ServeMux {
	//check the prefix for sanity
	pre := self.Prefix + "/"
	path := r.URL.Path
	if strings.HasSuffix(path, "/") && path != "/" {
		path = path[0 : len(path)-1]
	}
	if self.Prefix != "" {
		if !strings.HasPrefix(path, pre) {
			panic(fmt.Sprintf("expected prefix %s on the URL path but not found on: %s", pre, path))
		}
		path = path[len(pre):]
	}
	parts := strings.Split(path, "/")
	bundle, err := self.IO.BundleHook(w, r, self.SessionMgr)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create parameter bundle:%s", err), http.StatusInternalServerError)
		return nil
	}
	self.DispatchSegment(mux, w, r, parts, self.Root, bundle)
	return nil
}

//DispatchSegment is responsible for taking a part of the url, starting from the left
//and breaking it into segments for processing.  This is called by Dispatch() to initiate
//processing at the top level of resources but will be called recursively during
//dispatch processing.
func (self *RawDispatcher) DispatchSegment(mux *ServeMux, w http.ResponseWriter, r *http.Request,
	parts []string, current *RestNode, bundle PBundle) {

	var err error

	//find the resource and, if present, the id
	matched, id, rez, rezUdid := self.resolve(parts, current)
	if matched == "" {
		//typically trips the error dispatcher
		http.NotFound(w, r)
		return
	}
	method := strings.ToUpper(r.Method)
	//compute the parameter bundle

	var body interface{}
	var num int64

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
				return
			}
		}
	}

	//
	// RECURSE LOOKING FOR NEXT SEGMENT?
	//
	count := 1
	if id != "" {
		count = 2
	}
	if len(parts) > count {
		//we need to shear off the front parts and process the id
		if rezUdid == nil {
			if num <= 0 {
				http.Error(w, fmt.Sprintf("Bad request id"), http.StatusBadRequest)
				return
			}
			if rez.find == nil {
				http.Error(w, "Not implemented (FIND)", http.StatusNotImplemented)
				return
			}
			if self.Auth != nil && !self.Auth.Find(rez, num, bundle) {
				//typically trips the error dispatcher
				http.Error(w, "Not authorized (FIND)", http.StatusUnauthorized)
				return
			}
			result, err := rez.find.Find(num, bundle)
			if err != nil {
				self.SendError(err, w, "Internal error on Find")
				return
			} else {
				bundle.SetParentValue(rez.typ, result)
				node, ok := current.Children[parts[2]]
				if !ok {
					node, ok = current.ChildrenUdid[parts[2]]
					if !ok {
						http.Error(w, fmt.Sprintf("No such subresource:%s", parts[2]),
							http.StatusNotFound)
					}
				}
				//RECURSE
				self.DispatchSegment(mux, w, r, parts[2:],
					node, bundle)
				return
			}
		}
		//it's a UDID
		if rezUdid.find == nil {
			//typically trips the error dispatcher
			http.Error(w, "Not implemented (FIND,UDID)", http.StatusNotImplemented)
			return
		}
		if self.Auth != nil && !self.Auth.FindUdid(rezUdid, id, bundle) {
			//typically trips the error dispatcher
			http.Error(w, "Not authorized (FIND, UDID)", http.StatusUnauthorized)
			return
		}
		result, err := rezUdid.find.Find(id, bundle)
		if err != nil {
			self.SendError(err, w, "Internal error on Find (UDID")
			return
		}
		bundle.SetParentValue(rezUdid.typ, result)
		node, ok := current.Children[parts[2]]
		if !ok {
			node, ok = current.ChildrenUdid[parts[2]]
			if !ok {
				http.Error(w, fmt.Sprintf("No such subresource:%s", parts[2]),
					http.StatusNotFound)
			}
		}
		//RECURSE
		self.DispatchSegment(mux, w, r, parts[2:],
			node, bundle)
		return
	}

	//
	//pull anything from the body that's there, we might need it
	//
	if rezUdid == nil {
		body, err = self.IO.BodyHook(r, &rez.restShared)
		if err != nil {
			http.Error(w, fmt.Sprintf("badly formed body data: %s", err), http.StatusBadRequest)
			return
		}
	} else {
		body, err = self.IO.BodyHook(r, &rezUdid.restShared)
		if err != nil {
			http.Error(w, fmt.Sprintf("badly formed body data: %s", err), http.StatusBadRequest)
			return
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
					return
				}
				if self.Auth != nil && !self.Auth.Index(&rez.restShared, bundle) {
					//typically trips the error dispatcher
					http.Error(w, "Not authorized (INDEX)", http.StatusUnauthorized)
					return
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
					return
				}
				if self.Auth != nil && !self.Auth.Index(&rezUdid.restShared, bundle) {
					//typically trips the error dispatcher
					http.Error(w, "Not authorized (INDEX, UDID)", http.StatusUnauthorized)
					return
				}
				result, err := rezUdid.index.Index(bundle)
				if err != nil {
					self.SendError(err, w, "Internal error on Index (UDID)")
				} else {
					//go through encoding
					self.IO.SendHook(&rezUdid.restShared, w, bundle, result, "")
				}
			}
			return
		} else { //FINDER
			if rez != nil {
				if rez.find == nil {
					//typically trips the error dispatcher
					http.Error(w, "Not implemented (FIND)", http.StatusNotImplemented)
					return
				}
				if self.Auth != nil && !self.Auth.Find(rez, num, bundle) {
					//typically trips the error dispatcher
					http.Error(w, "Not authorized (FIND)", http.StatusUnauthorized)
					return
				}
				result, err := rez.find.Find(num, bundle)
				if err != nil {
					self.SendError(err, w, "Internal error on Find")
				} else {
					self.IO.SendHook(&rez.restShared, w, bundle, result, "")
				}
				return
			} else {
				//UDID RESOURCE
				if rezUdid.find == nil {
					//typically trips the error dispatcher
					http.Error(w, "Not implemented (FIND,UDID)", http.StatusNotImplemented)
					return
				}
				if self.Auth != nil && !self.Auth.FindUdid(rezUdid, id, bundle) {
					//typically trips the error dispatcher
					http.Error(w, "Not authorized (FIND, UDID)", http.StatusUnauthorized)
					return
				}
				result, err := rezUdid.find.Find(id, bundle)
				if err != nil {
					self.SendError(err, w, "Internal error on Find (UDID")
				} else {
					self.IO.SendHook(&rezUdid.restShared, w, bundle, result, "")
				}
				return
			}
		}
	case "POST":
		if rez != nil {
			if id != "" {
				http.Error(w, "can't POST to a particular resource, did you mean PUT?", http.StatusBadRequest)
				return
			}
			if rez.post == nil {
				http.Error(w, "Not implemented (POST)", http.StatusNotImplemented)
				return
			}
			if self.Auth != nil && !self.Auth.Post(&rez.restShared, bundle) {
				http.Error(w, "Not authorized (POST)", http.StatusUnauthorized)
				return
			}
			result, err := rez.post.Post(body, bundle)
			if err != nil {
				self.SendError(err, w, "Internal error on Post")
			} else {
				self.IO.SendHook(&rez.restShared, w, bundle, result, self.location(rez.name, false, result))
			}
			return
		} else {
			//UDID POST
			if id != "" {
				http.Error(w, "can't (UDID) POST to a particular resource, did you mean PUT?", http.StatusBadRequest)
				return
			}
			if rezUdid.post == nil {
				http.Error(w, "Not implemented (POST, UDID)", http.StatusNotImplemented)
				return
			}
			if self.Auth != nil && !self.Auth.Post(&rezUdid.restShared, bundle) {
				http.Error(w, "Not authorized (POST)", http.StatusUnauthorized)
				return
			}
			result, err := rezUdid.post.Post(body, bundle)
			if err != nil {
				self.SendError(err, w, "Internal error on Post")
			} else {
				self.IO.SendHook(&rezUdid.restShared, w, bundle, result, self.location(rezUdid.name, true, result))
			}
			return

		}
	case "PUT", "DELETE":
		if id == "" {
			http.Error(w, fmt.Sprintf("%s requires a resource id or UDID", method), http.StatusBadRequest)
			return
		}
		if method == "PUT" {
			if rez != nil {
				if rez.put == nil {
					http.Error(w, "Not implemented (PUT)", http.StatusNotImplemented)
					return
				}
				if self.Auth != nil && !self.Auth.Put(rez, num, bundle) {
					http.Error(w, "Not authorized (PUT)", http.StatusUnauthorized)
					return
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
					return
				}
				if self.Auth != nil && !self.Auth.PutUdid(rezUdid, id, bundle) {
					http.Error(w, "Not authorized (PUT, UDID)", http.StatusUnauthorized)
					return
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
					return
				}
				if self.Auth != nil && !self.Auth.Delete(rez, num, bundle) {
					http.Error(w, "Not authorized (DELETE)", http.StatusUnauthorized)
					return
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
					return
				}
				if self.Auth != nil && !self.Auth.DeleteUdid(rezUdid, id, bundle) {
					http.Error(w, "Not authorized (DELETE, UDID)", http.StatusUnauthorized)
					return
				}
				result, err := rezUdid.del.Delete(id, bundle)
				if err != nil {
					self.SendError(err, w, "Internal error on Delete")
				} else {
					self.IO.SendHook(&rezUdid.restShared, w, bundle, result, "")
				}
			}
		}
		return
	}
	log.Printf("should not be able to reach here, probably bad method? from bad client?")
	http.Error(w, "bad client behavior", http.StatusBadRequest)
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
func (self *RawDispatcher) resolve(parts []string, node *RestNode) (string, string, *restObj, *restObjUdid) {
	//case 1: simple path to a normal resource
	rez, ok := node.Res[parts[0]]
	if ok && len(parts) == 1 {
		return parts[0], "", rez, nil
	}
	//case 2: simple path to a udid resource
	rezUdid, okUdid := node.ResUdid[parts[0]]
	if okUdid && len(parts) == 1 {
		return parts[0], "", nil, rezUdid
	}
	id := parts[1]
	uriPathParent := parts[0]

	//case 3, path to a resource ID (int)
	rez, ok = node.Res[uriPathParent]
	if ok {
		return parts[0], id, rez, nil
	}

	//case 4, path to a resource ID (UDID)
	rezUdid, ok = node.ResUdid[uriPathParent]
	if ok {
		return parts[0], id, nil, rezUdid
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
