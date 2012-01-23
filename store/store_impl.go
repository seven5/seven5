package store

import (
	"errors"
	"fmt"
	"github.com/seven5/gomemcache/memcache"
	"net"
)

//StoreImpl is the beginnings of a type that can express all the operations need to be a
//store implementation.  For now we just mimic the methods of the one implementation.
type StoreImpl interface {
	//Get reads the store (lookup)
	Get(key string) (StoredItem, error)
	//GetMulti does not guaratee the returned results are in the same order as requsted in
	//the parameter.
	GetMulti(keys []string) (map[string]StoredItem, error)
	//Set is an unconditional write.
	Set(item StoredItem) error
	//CompareAndSwap is the write part of the CAS operation.  The same item must have been used
	//in a previous call to Get() or GetMulti() because it holds state needed to decide if a
	//CAS violation has occured.
	CompareAndSwap(item StoredItem) error
	//Delete the key from the store.
	Delete(key string) error
	//Increment is an atomic update of a key value. delta is the amount to increment by
	Increment(key string, delta uint64) (uint64, error)
	//Add only writes the value if nothing is present already.  If something is present,
	//you get ErrorAddFailedInStore
	Add(StoredItem) error

	//DestroyAll is used by "out of band" code to do go around the normal API and destroy
	//everything.  Typically, such code is a test that is running against localhost.  This
	//is too convenient to not have, but probably should not be considered part of the 
	//API proper.
	DestroyAll(host string) error
}

//StoredItem is the beginnings of a type that can express the returned results of a get
//and the value to store for a put.
type StoredItem interface {
	//Key gets the key. Size is limited to 250 bytes.
	Key() string
	//Value gets the value. Size is limited to 1MB.
	Value() []byte

	//SetKey sets the key for this item. Size is limited to 250 bytes.
	SetKey(string)
	//SetValue sets the value of this item. Size is limited to 1MB.
	SetValue([]byte)
}

const MEMCACHE_LOCALHOST = "localhost:11211"


//ErrorNotFoundInStore indicates, well, that it was not found in the store. 
var ErrorNotFoundInStore = errors.New("Key not found in storage layer")

//ErrorAddFailedInStore indicates there was already a value present, so Add() call had
//no effect. 
var ErrorAddFailedInStore = errors.New("Add failed because value was already present")

//ErrorCASConflictInStore indicates, well, that that your CompareAndSwap operation failed
//because the value was changed before you made the switch. 
var ErrorCASConflictInStore = errors.New("Compare and swap conflict")

//NewStore creates an implementation of StoreImpl and returns it.  The underlying
//implementation is memcached. 
func NewStoreImpl(server string) StoreImpl {
	return &memcacheWrapper{memcache.New(server)}
}

//NewStoredItem creates an implementation of a StoredItem and returns it.  The underlying
//implementation is memcached's Item
func NewStoredItem() StoredItem {
	return &itemWrapper{&memcache.Item{}}
}

//wrapper around the memcached item object
type itemWrapper struct {
	i *memcache.Item
}

//Key returns the key of this item
func (self *itemWrapper) Key() string {
	return self.i.Key
}

//Value returns the value of this item
func (self *itemWrapper) Value() []byte {
	return self.i.Value
}

//SetKey sets the key for this item
func (self *itemWrapper) SetKey(k string) {
	if len([]byte(k)) > 250 {
		panic(fmt.Sprintf("key %s is too large", k))
	}
	self.i.Key = k
}

//SetValue sets the value for this item
func (self *itemWrapper) SetValue(v []byte) {
	if len(v) > 1e6 {
		panic("value is too large")
	}
	self.i.Value = v
}

//memcacheWrapper is a tiny layer around a memcache client that allows it to play nicely
//with the StoreImpl construct.  This wouldn't be necessary except that the error return
//values need to be adjusted in a couple of cases and the parameter type changed to
//StoredItem
type memcacheWrapper struct {
	c *memcache.Client
}

//Get is an read (lookup).
func (self *memcacheWrapper) Get(key string) (StoredItem, error) {
	i, err := self.c.Get(key)
	if err == memcache.ErrCacheMiss {
		return &itemWrapper{i}, ErrorNotFoundInStore
	}
	return &itemWrapper{i}, err
}

//GetMulti is a read of many items, order not guaranteed
func (self *memcacheWrapper) GetMulti(keys []string) (map[string]StoredItem, error) {
	var result map[string]StoredItem

	m, err := self.c.GetMulti(keys)
	if m != nil {
		result = make(map[string]StoredItem)
		for k, v := range m {
			result[k] = &itemWrapper{v}
		}
	}
	if err == memcache.ErrCacheMiss {
		return result, ErrorNotFoundInStore
	}
	return result, err
}

//Set is the unconditional write operation
func (self *memcacheWrapper) Set(item StoredItem) error {
	wrapper := item.(*itemWrapper)
	err := self.c.Set(wrapper.i)
	if err == memcache.ErrCacheMiss {
		return ErrorNotFoundInStore
	}
	return err
}

//CompareAndSwap is the write part of the CAS operation.  The same item must have been used
//in a previous call to Get() or GetMulti() because it holds state needed to decide if a
//CAS violation has occured.
func (self *memcacheWrapper) CompareAndSwap(item StoredItem) error {
	wrapper := item.(*itemWrapper)

	err := self.c.CompareAndSwap(wrapper.i)
	if err == memcache.ErrCacheMiss {
		return ErrorNotFoundInStore
	}
	if err == memcache.ErrCASConflict {
		return ErrorCASConflictInStore
	}
	return err
}

//Delete is a straight up gangsta mack cap poppin.
func (self *memcacheWrapper) Delete(key string) error {
	err := self.c.Delete(key)
	if err == memcache.ErrCacheMiss {
		return ErrorNotFoundInStore
	}
	return err
}

//Increment is an atomic update of a key value.  delta is the amount to increment the 
//counter by.  
func (self *memcacheWrapper) Increment(key string, delta uint64) (uint64, error) {
	ret, err := self.c.Increment(key, delta)
	if err == memcache.ErrCacheMiss {
		return ret, ErrorNotFoundInStore
	}
	return ret, nil
}

//Add is really "only write if nothing is there now" 
func (self *memcacheWrapper) Add(item StoredItem) error {
	wrapper := item.(*itemWrapper)

	err := self.c.Add(wrapper.i)
	if err == memcache.ErrCacheMiss { //seems REALLY unlikely
		return ErrorNotFoundInStore
	}
	if err == memcache.ErrNotStored {
		return ErrorAddFailedInStore
	}
	return err
}

//DestroyAll will delete all data from the hosts (or from localhost on 11211 if no hosts have)
//have been set.  This call is "out of band" can be executed whether or not there is a connected
//client active.
func (self *memcacheWrapper) DestroyAll(host string) error {
	//fmt.Fprintf(os.Stderr, "Warning: clearing memcache....\n")

	conn, err := net.Dial("tcp", host)
	if err != nil {
		return err
	}
	conn.Write([]byte("flush_all\r\n"))
	waitForResponseBuffer := make([]byte, 1, 1)
	_,err = conn.Read(waitForResponseBuffer)
	if err != nil {
		return err
	}

	err = conn.Close()
	if err != nil {
		return err
	}

	return nil
}
