package seven5

import (
	_ "fmt"
	"net/http"
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
}

type simplePBundle struct {
	h   map[string]string
	q   map[string]string
	s   Session
	mgr SessionManager
	out map[string]string
}

func (self *simplePBundle) ReturnHeaders() []string {
	result := []string{}
	for k, _ := range self.out {
		result = append(result, k)
	}
	return result
}

func (self *simplePBundle) SetReturnHeader(k string, v string) {
	self.out[k] = v
}
func (self *simplePBundle) ReturnHeader(k string) string {
	return self.out[k]
}

func (self *simplePBundle) Header(s string) (string, bool) {
	v, ok := self.h[s]
	return v, ok
}
func (self *simplePBundle) Query(s string) (string, bool) {
	v, ok := self.q[s]
	return v, ok
}

func (self *simplePBundle) Session() Session {
	return self.s
}
func (self *simplePBundle) UpdateSession(i interface{}) (Session, error) {
	return self.mgr.Update(self.s, i)
}

//NewSimplePBundle needs to hold a reference to the session manager as well
//as the session because it must be able to update the information stored
//about a particular sesison.
func NewSimplePBundle(r *http.Request, s Session, mgr SessionManager) (PBundle, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	return &simplePBundle{
		h:   ToSimpleMap(r.Header),
		q:   ToSimpleMap(map[string][]string(r.Form)),
		s:   s,
		mgr: mgr,
		out: make(map[string]string),
	}, nil
}

//ToSimpleMap converts an http level map with multiple strings as value to single string value.
//There are a number of places in HTTP (such as headers and query parameters) where this is
//possible and legal according to the spec, but still silly so we just use single valued
//values.
func ToSimpleMap(m map[string][]string) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		result[k] = strings.TrimSpace(v[0])
	}
	return result
}
