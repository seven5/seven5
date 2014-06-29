package seven5

import (
	_ "fmt"
)

//NewBaseDispatcher returns a raw dispatcher that has several defaults set.
//* The Allow() interfaces are used for authorization checks
//* The application will keep a single cookie on the browser (that's why the name is passed in)
//* The application will keep a session associated with the cookie for each "logged in" user (in memory)
//* Json is used to encode and decode the wire types
//* Rest resources dispatched by this object are mapped to /rest in the URL space.
//You can pass a SessionManager to this method if you want to use your own implementation and this
//is common for applications that involve users.
func NewBaseDispatcher(appName string, optionalSm SessionManager) *BaseDispatcher {
	var sm SessionManager
	prefix := "/rest"
	if optionalSm != nil {
		sm = optionalSm
	} else {
		sm = NewSimpleSessionManager()
	}
	cm := NewSimpleCookieMapper(appName)
	holder := NewSimpleTypeHolder()
	result := &BaseDispatcher{}
	io := &RawIOHook{&JsonDecoder{}, &JsonEncoder{}, cm}
	result.RawDispatcher = NewRawDispatcher(io, sm, result, holder, prefix)
	return result
}

//BaseDispatcher is a slight "specialization" of the RawDispatcher for REST resources.  BaseDispatcher
//understands how to dispatch to REST resources (like Raw) but can also handle the Allower protocol for
//primitive, coarse-grained authorization.  Additionally, it allows easy creation of a BaseDispatcher
//with a custom SessionManager, as this is often used with user roles (and Allow protocol).
type BaseDispatcher struct {
	*RawDispatcher
}

//Index checks with AllowReader.AllowRead to allow/refuse access to this method on _any_ resource
//associated with this BaseDispatcher.
func (self *BaseDispatcher) Index(d *restObj, bundle PBundle) bool {
	allowReader, ok := d.index.(AllowReader)
	if !ok {
		return false
	}
	return allowReader.AllowRead(bundle)
}

//Post checks with AllowWriter.AllowWrite to allow/refuse access to this method on _any_ resource
//associated with this BaseDispatcher.
func (self *BaseDispatcher) Post(d *restObj, bundle PBundle) bool {
	allowWriter, ok := d.post.(AllowWriter)
	if !ok {
		return false
	}
	return allowWriter.AllowWrite(bundle)
}

//Find checks with Allower.Allow(FIND) to allow/refuse access to this method on _any_ resource
//associated with this BaseDispatcher.
func (self *BaseDispatcher) Find(d *restObj, num int64, bundle PBundle) bool {
	allow, ok := d.find.(Allower)
	if !ok {
		return false
	}
	return allow.Allow(num, "GET", bundle)
}

//Find checks with Allower.Allow(PUT) to allow/refuse access to this method on _any_ resource
//associated with this BaseDispatcher.
func (self *BaseDispatcher) Put(d *restObj, num int64, bundle PBundle) bool {
	allow, ok := d.put.(Allower)
	if !ok {
		return false
	}
	return allow.Allow(num, "PUT", bundle)
}

//Find checks with Allower.Allow(DELETE) to allow/refuse access to this method on _any_ resource
//associated with this BaseDispatcher.
func (self *BaseDispatcher) Delete(d *restObj, num int64, bundle PBundle) bool {
	allow, ok := d.del.(Allower)
	if !ok {
		return false
	}
	return allow.Allow(num, "DELETE", bundle)
}
