package seven5

import (
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

type PBundle interface {
	Header(string) (string, bool)
	Query(string) (string, bool)
	Session() Session
	ReturnHeader(string) string
	SetReturnHeader(string, string)
	ReturnHeaders() []string
	UpdateSession(interface{}) (Session, error)
	DestroySession() error
	ParentValue(interface{}) interface{}
	SetParentValue(reflect.Type, interface{})
	IntQueryParameter(string, int64) int64
}

type simplePBundle struct {
	h      map[string]string
	q      map[string]string
	s      Session
	mgr    SessionManager
	out    map[string]string
	parent map[reflect.Type]interface{}
}

//ReturnHeaders gets all the header _keys_ that should be returned the client.
func (self *simplePBundle) ReturnHeaders() []string {
	result := []string{}
	for k, _ := range self.out {
		result = append(result, k)
	}
	return result
}

//SetReturnHeader associates the value v with the header k in the result returned
//to the client. This method does not allow setting multiple values for a single
//key because that's silly.
func (self *simplePBundle) SetReturnHeader(k string, v string) {
	self.out[k] = v
}

//ReturnHeader retrieves the header associated with k.  This interface
//does not support the (terribly poorly thought out) HTTP model where a given
//key (k) in the header set can have multiple values.
func (self *simplePBundle) ReturnHeader(k string) string {
	return self.out[k]
}

//Header returns the header sent from the client named s. If there is no such
//header, false will be returned in the second argument.
func (self *simplePBundle) Header(s string) (string, bool) {
	v, ok := self.h[strings.ToLower(s)]
	return v, ok
}
func (self *simplePBundle) Query(s string) (string, bool) {
	v, ok := self.q[s]
	return v, ok
}

//Session returns the Session object associated with this Pbundle (usally
//computed on each request by the SessionManager).
func (self *simplePBundle) Session() Session {
	return self.s
}

//UpdateSession associates a new data blob (i) with the currently in use session.
//This will blow up if there is no associated SessionManager.
func (self *simplePBundle) UpdateSession(i interface{}) (Session, error) {
	return self.mgr.Update(self.s, i)
}

//DestroySession removes the session associated with this Pbundle from the
//associated SessionManager.  Note that this wil blow up if you have either
//no session manager but is ignored if there is no session.
func (self *simplePBundle) DestroySession() error {
	if self.Session() == nil {
		return nil
	}
	return self.mgr.Destroy(self.s.SessionId())
}

//ParentValue returns a parent resource's contribution a child resource. So, for
//a url like /rest/foo/23/bar/98 the child resource at bar/98 can find the information
//about the parent resource at foo/23 with ParentValue(foosWireType).
//The value returned is constructed by calling Find() on the parent resource
//and that Find() succeeding.
func (self *simplePBundle) ParentValue(wire interface{}) interface{} {
	return self.parent[reflect.TypeOf(wire)]
}

//SetParentValue associates a particular parent value with the given wire type.
//Clients typically don't need this method, because it is called via the dispatch
//mechanism as the URL is being processed.  Because it's primarily for internal
//use, the call takes the _type_ of the client wire type, not an example of it.
func (self *simplePBundle) SetParentValue(t reflect.Type, value interface{}) {
	self.parent[t] = value
}

//IntQueryParameter returns the value of the query parameter name with a
//default value of def.  The default value is used if either the parameter
//is not present, or cannot be parsed as an int.
func (self *simplePBundle) IntQueryParameter(name string, def int64) int64 {
	raw, ok := self.Query(name)
	if !ok {
		return def
	}
	i, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return def
	}
	return i
}

//NewSimplePBundle needs to hold a reference to the session manager as well
//as the session because it must be able to update the information stored
//about a particular sesison.
func NewSimplePBundle(r *http.Request, s Session, mgr SessionManager) (PBundle, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	return &simplePBundle{
		h:      ToSimpleMap(r.Header),
		q:      ToSimpleMap(map[string][]string(r.Form)),
		s:      s,
		mgr:    mgr,
		out:    make(map[string]string),
		parent: make(map[reflect.Type]interface{}),
	}, nil
}

//ToSimpleMap converts an http level map with multiple strings as value to single string value.
//There are a number of places in HTTP (such as headers and query parameters) where this is
//possible and legal according to the spec, but still silly so we just use single valued
//values.
func ToSimpleMap(m map[string][]string) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		result[strings.ToLower(k)] = strings.TrimSpace(v[0])
	}
	return result
}

//NewTestPBundle makes a Pbundle from the given constants.  Note that you
//can supply a session manager of nil and the consumer of this object doesn't
//try to update the current session, this is ok.
func NewTestPBundle(headers map[string]string, query map[string]string, session Session,
	mgr SessionManager, output map[string]string, parent map[reflect.Type]interface{}) PBundle {

	return &simplePBundle{
		h:      headers,
		q:      query,
		s:      session,
		mgr:    mgr,
		out:    output,
		parent: parent,
	}
}
