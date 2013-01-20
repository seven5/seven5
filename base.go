package seven5

import (
	_ "fmt"
)

//NewBaseDispatcher returns a raw dispatcher that has several defaults set.  
//* The Allow() interfaces are used for authorization checks
//* The application will keep a single cookie on the browser (that's why the name is passed in)
//* The application will keep a session associated with the cookie for each "logged in" user (in memory)
//* Json is used to encode and decode the wire types
//* Rest resources dispatched by this object are mapped to /rest in the URL space
func NewBaseDispatcher(appName string) *BaseDispatcher {
	sm := NewSimpleSessionManager()
	cm := NewSimpleCookieMapper(appName, sm)
	result :=&BaseDispatcher{}
	result.RawDispatcher = NewRawDispatcher(&JsonEncoder{}, &JsonDecoder{}, cm, result, "/rest")
	return result
}

type BaseDispatcher struct {
	*RawDispatcher
}


func (self *BaseDispatcher) Index(d *restObj, bundle PBundle) bool {
	allowReader, ok := d.index.(AllowReader)
	if !ok {
		return false
	}
	return allowReader.AllowRead(bundle)
}

func (self *BaseDispatcher) Post(d *restObj, bundle PBundle) bool {
	allowWriter, ok := d.post.(AllowWriter)
	if !ok {
		return false
	}
	return allowWriter.AllowWrite(bundle)
}

func (self *BaseDispatcher) Find(d *restObj,num Id,  bundle PBundle) bool {
	allow, ok := d.find.(Allower)
	if !ok {
		return false
	}
	return allow.Allow(num, "GET", bundle)
}

func (self *BaseDispatcher) Put(d *restObj,num Id,  bundle PBundle) bool {
	allow, ok := d.put.(Allower)
	if !ok {
		return false
	}
	return allow.Allow(num, "PUT", bundle)
}

func (self *BaseDispatcher) Delete(d *restObj, num Id, bundle PBundle) bool {
	allow, ok := d.del.(Allower)
	if !ok {
		return false
	}
	return allow.Allow(num, "DELETE", bundle)
}
