package seven5

//NewBaseDispatcher returns a raw dispatcher that has several defaults set.
//* The Allow() interfaces are used for authorization checks
//* The application will keep a single cookie on the browser (that's why the cookie mapper is passed in)
//* The application will keep a session associated with the cookie for each "logged in" user (via the SessionManager)
//* Json is used to encode and decode the wire types
//* Rest resources dispatched by this object are mapped to /rest in the URL space.
//You must pass an already created session manager into this method
//(see NewSimpleSessionManager(...))
func NewBaseDispatcher(sm SessionManager, cm CookieMapper) *BaseDispatcher {
	prefix := "/rest"
	result := &BaseDispatcher{}
	io := &RawIOHook{&JsonDecoder{}, &JsonEncoder{}, cm}
	result.RawDispatcher = NewRawDispatcher(io, sm, result, prefix)
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
func (self *BaseDispatcher) Index(d *restShared, bundle PBundle) bool {
	allowReader, ok := d.index.(AllowReader)
	if !ok {
		return true
	}
	return allowReader.AllowRead(bundle)
}

//Post checks with AllowWriter.AllowWrite to allow/refuse access to this method on _any_ resource
//associated with this BaseDispatcher.
func (self *BaseDispatcher) Post(d *restShared, bundle PBundle) bool {
	allowWriter, ok := d.post.(AllowWriter)
	if !ok {
		return true
	}
	return allowWriter.AllowWrite(bundle)
}

//Find checks with Allower.Allow(FIND) to allow/refuse access to this method on _any_ resource
//associated with this BaseDispatcher.
func (self *BaseDispatcher) Find(d *restObj, num int64, bundle PBundle) bool {
	allow, ok := d.find.(Allower)
	if !ok {
		return true
	}
	return allow.Allow(num, "GET", bundle)
}

//Put checks with Allower.Allow(PUT) to allow/refuse access to this method on _any_ resource
//associated with this BaseDispatcher.
func (self *BaseDispatcher) Put(d *restObj, num int64, bundle PBundle) bool {
	allow, ok := d.put.(Allower)
	if !ok {
		return true
	}
	return allow.Allow(num, "PUT", bundle)
}

//Find checks with Allower.Allow(DELETE) to allow/refuse access to this method on _any_ resource
//associated with this BaseDispatcher.
func (self *BaseDispatcher) Delete(d *restObj, num int64, bundle PBundle) bool {
	allow, ok := d.del.(Allower)
	if !ok {
		return true
	}
	return allow.Allow(num, "DELETE", bundle)
}

//Find checks with Allower.AllowUdid(GET) to allow/refuse access to this method on _any_ resource
//associated with this BaseDispatcher.
func (self *BaseDispatcher) FindUdid(d *restObjUdid, id string, bundle PBundle) bool {
	allow, ok := d.find.(AllowerUdid)
	if !ok {
		return true
	}
	return allow.Allow(id, "GET", bundle)
}

//Put checks with Allower.AllowUdid(PUT) to allow/refuse access to this method on _any_ resource
//associated with this BaseDispatcher.
func (self *BaseDispatcher) PutUdid(d *restObjUdid, id string, bundle PBundle) bool {
	allow, ok := d.put.(AllowerUdid)
	if !ok {
		return true
	}
	return allow.Allow(id, "PUT", bundle)
}

//Find checks with AllowerUdid.Allow(DELETE) to allow/refuse access to this method on _any_ resource
//associated with this BaseDispatcher.
func (self *BaseDispatcher) DeleteUdid(d *restObjUdid, id string, bundle PBundle) bool {
	allow, ok := d.del.(AllowerUdid)
	if !ok {
		return true
	}
	return allow.Allow(id, "DELETE", bundle)
}
