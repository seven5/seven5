package store

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/bradfitz/gomemcache"
	"net"
	//"os"
	"reflect"
	//	"strconv"
)

type MemcacheGobStore struct {
	*memcache.Client
}

const (
	LOCALHOST = "localhost:11211"
	IDKEY     = "%s-idcounter"
	RECKEY    = "%s-%d"
	EXTRAKEY  = "%s-key-%s"
)

//DestroyAll will delete all data from the hosts (or from localhost on 11211 if no hosts have)
//have been set.  This call is "out of band" can be executed whether or not there is a connected
//client active.
func (self *MemcacheGobStore) DestroyAll(host ...string) error {
	//fmt.Fprintf(os.Stderr, "Warning: clearing memcache....\n")

	for _, h := range host {
		conn, err := net.Dial("tcp", h)
		if err != nil {
			return err
		}
		conn.Write([]byte("flush_all\r\n"))
		waitForResponseBuffer := make([]byte, 1, 1)
		conn.Read(waitForResponseBuffer)
		conn.Close()

	}
	return nil
}

//GetNextId will return the next id for the type X.  It will create the necessary structures in
//the memcache if that is needed.
func (self *MemcacheGobStore) GetNextId(typeName string) (uint64, error) {
	key := fmt.Sprintf(IDKEY, typeName)
	newValue, err := self.Client.Increment(key, uint64(1))
	if err == nil {
		return newValue, nil
	}
	if err != memcache.ErrCacheMiss {
		return uint64(0), err
	}

	item := &memcache.Item{Key: key, Value: []byte("0")}
	err = self.Client.Add(item)
	if err == memcache.ErrNotStored {
		//try a 2nd time, maybe race condition
		newValue, err = self.Client.Increment(key, uint64(1))
		if err != nil {
			return uint64(0), err
		}
		return newValue, nil
	}
	if err != nil {
		return uint64(0), err
	}
	//try to do increment again since we set it successfully
	result, err := self.Client.Increment(key, uint64(1))

	if err != nil {
		return uint64(0), err
	}
	return result, nil
}

//Write takes a structure and sends it to memcache.  If the id field is not set yet, it creates
//an Id for the item before writing it to memcache.  The value passed must be a pointer to a 
//struct and the struct must have an Id field that is uint64.
func (self *MemcacheGobStore) Write(s interface{}) error {
	var value uint64
	var typeName string
	var err error

	if value, typeName, err = GetIdValueAndStructureName(s); err != nil {
		return err
	}
	if value == uint64(0) {
		newId, err := self.GetNextId(typeName)
		if err != nil {
			return err
		}
		id := reflect.ValueOf(s).Elem().FieldByName("Id")
		id.SetUint(newId)
		value = newId
	}
	//at this point value holds the number of the record
	key := fmt.Sprintf(RECKEY, typeName, value)
	buffer := new(bytes.Buffer)
	enc := gob.NewEncoder(buffer)
	if err := enc.Encode(s); err != nil {
		return err
	}
	//stuff the bytes into the store
	item := &memcache.Item{Key: key, Value: buffer.Bytes()}
	err = self.Client.Set(item)
	if err != nil {
		return err
	}

	//update all the extra keys using writeKey
	err = self.walkAllExtraKeys(s, func(n string, v string) error {
		return self.writeKey(s, n, typeName, v, value)
	})

	return nil
}

//fieldFunc is the type of function you have to supply to walkAllExtraKeys which will get called
//for each key of the structure (or methods, if declared in the structure defn)
type fieldFunc func(fieldOrMethodName string, flattenedValue string) error

//walkAllExtraKeys is a utility route to run a function for every key mentioned in the structure
//declaration of s.  This function assumes that s has already been checked an is properly
//formed.
func (self *MemcacheGobStore) walkAllExtraKeys(s interface{}, fn fieldFunc) error{
	//write the extra indexes
	fields, methods := GetStructKeys(s)
	var mapKey string
	var err error

	//try the fields
	for _, k := range fields {
		mapKey = toStringMethodOrSprintf(k.Value)
		err = fn(k.Name, mapKey)
		if err != nil {
			return err
		}
	}
	//now the methods
	for _, m := range methods {
		out := m.Meth.Call([]reflect.Value{})
		mapKey = toStringMethodOrSprintf(out[0])
		err = fn(m.Name, mapKey)
		if err != nil {
			return err
		}
	}
	return nil
}

//toStringMethodOrSprintf needs a lot of improvement for the basic types! It's crap right now.
//TODO
func toStringMethodOrSprintf(v reflect.Value) string {
	method := v.MethodByName("String")
	if method == ZeroValue {
		//why do I need to elucidate all the types here?
		if v.Kind() == reflect.Int {
			return fmt.Sprintf("%d", v.Int())
		}

		return fmt.Sprintf("%v", v)
	}
	return method.Call(nil)[0].String()
}

//FindById is the reverse of Write and reads a structure from memcached for a given type and Id.
//The contents of the first parameter will be overwritten by this call and the previous contents
//ignored (although its type must be a pointer to a struct of the right type)
func (self *MemcacheGobStore) FindById(s interface{}, id uint64) error {
	var typeName string
	var err error

	if _, typeName, err = GetIdValueAndStructureName(s); err != nil {
		return err
	}
	key := fmt.Sprintf(RECKEY, typeName, id)
	var item *memcache.Item
	if item, err = self.Client.Get(key); err != nil {
		return err
	}
	buffer := bytes.NewBuffer(item.Value)
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(s)
}

//writeKey assumes that the pointer to struct has already been checked and is ok.
//writeKey tries to find a map that it can use to index the records by the value
//provided.  it creates the map if necessary.  The map is per keyName with the keys
//of the map being values and the values are ids.
func (self *MemcacheGobStore) writeKey(s interface{}, keyName string, typeName string, mapKey string, id uint64) error {
	var index map[string][]uint64
	var item *memcache.Item
	if err := self.readIndex(&index, &item, typeName, keyName, true); err != nil {
		return err
	}

	//ok, if we get here, we are ok and have the index loaded (or created)...
	index[mapKey] = append(index[mapKey], id)

	return self.writeIndex(index, item, typeName, keyName)
}

//writeIndex puts an index in memcached by serializing it with gobs.  it needs to be passed 
//back the same item value that was returned from readIndex so we can correctly detect 
//concurrency problems. if item is nil it assumes that this is a creation (and no concurrency check needed)
func (self *MemcacheGobStore) writeIndex(index map[string][]uint64, item *memcache.Item, typeName string, keyName string) error {
	//serialize to gob the index
	buffer := new(bytes.Buffer)
	enc := gob.NewEncoder(buffer)
	if err := enc.Encode(index); err != nil {
		return err
	}

	//did we have anything in the index before? if so, try to swap in new value but abort
	//if that fails
	if item != nil {
		item.Value = buffer.Bytes()
		return self.Client.CompareAndSwap(item)
	}
	//this is a brand new index, write it out to disk
	newItem := &memcache.Item{Key: fmt.Sprintf(EXTRAKEY, typeName, keyName), Value: buffer.Bytes()}
	return self.Client.Set(newItem)
}

//readIndex pulls an index from memcached and sets the first parameter to it, if there was no error.
//the create flag indicates if not finding the item in memcached is an error or it should be
//created (true).  The item parameter is set to the retrieved Item object for use later
//with compareAndSwap
func (self *MemcacheGobStore) readIndex(result *map[string][]uint64, item **memcache.Item, typeName string, keyName string, create bool) error {
	//compute memcached key
	key := fmt.Sprintf(EXTRAKEY, typeName, keyName)
	var err error

	*item, err = self.Client.Get(key)

	//if not there, create the map from scratch, based on create param
	if err == memcache.ErrCacheMiss {
		if create {
			*result = make(map[string][]uint64)
			*item = nil
		} else {
			return err
		}
	} else if err == nil {
		//no error, read the map
		buffer := bytes.NewBuffer((*item).Value)
		decoder := gob.NewDecoder(buffer)
		err = decoder.Decode(result)
		if err != nil {
			return err
		}
	} else {
		//some other error, give up
		return err
	}
	//if we reach here we have the result and item updated properly
	return nil
}

//deleteKey does the work to update an index when an item is deleted
func (self *MemcacheGobStore) deleteKey(s interface{}, keyName string, typeName string, mapKey string, id uint64) error {
	var index map[string][]uint64
	var item *memcache.Item
	
	if err := self.readIndex(&index, &item, typeName, keyName, true); err != nil {
		return err
	}
	slice := index[mapKey]
	ok := false
	for i := 0; i < len(slice); i++ {
		if slice[i] == id {
			slice[i] = slice[len(slice)-1]
			index[mapKey] = slice[0 : len(slice)-1]
			ok = true
			break
		}
	}
	if !ok {
		return INDEX_MISS
	}
	return self.writeIndex(index, item, typeName, keyName)
}

//FindByKey looks up a value in the memcache by a field _other_ than the Id field.  You have
//to supply the name of the field.  Further, that field must exist in the structure, be
//exported (uppercase).  The value field must match exactly the flattened version of the value
//when String() is called on it (at the time it is written). 
//This call results in n+1 roundtrips to the memcache server for n values because
//it first retrieves the ids of the objects that are stored under the value provided and
//then calls FindById on each one.  The result is placed the slice pointed to by the first
//element--and only as many results are returned as available places in the slice.
func (self *MemcacheGobStore) FindByKey(ptrToResult interface{}, keyName string, value string) (errReturn error) {
	var err error

	if reflect.TypeOf(ptrToResult).Kind() != reflect.Ptr {
		return BAD_SLICE_PTR
	}

	result := reflect.ValueOf(ptrToResult).Elem()

	if result.Type().Kind() != reflect.Slice {
		return BAD_SLICE
	}

	s := result.Type().Elem()
	if err = VerifyStructPointerFieldTypes(s); err != nil {
		return err
	}

	typeName := s.Elem().String()

	var item *memcache.Item
	var index map[string][]uint64
	if err=self.readIndex(&index,&item,typeName,keyName,false); err!=nil {
		return err
	}
	
	//get the item we are interested in
	slice, ok := index[value]
	if !ok || len(slice)==0 {
		//just tell the caller there is nothing with that value
		return nil
	}
	
	//loop over all the ids in the index, until we run out of ids or run out of space in
	//in the result slice
	for _, v := range index[value] {
		if result.Len() == result.Cap() {
			break
		}
		e := result.Type().Elem().Elem()
		obj := reflect.New(e)
		selfv := reflect.ValueOf(self)
		out := selfv.MethodByName("FindById").Call([]reflect.Value{obj, reflect.ValueOf(v)})
		if !out[0].IsNil() {
			rtn := reflect.ValueOf(errReturn)
			rtn.Set(out[0])
			return
		}
		l := result.Len()
		result.SetLen(l + 1)
		result.Index(l).Set(obj)
	}
	return nil
}

//DeleteById deletes an item from the store by its unique id.  This does the work to update all
//the index
func (self *MemcacheGobStore) DeleteById(s interface{}, id uint64) error {
	var typeName string
	var err,memcache_err error

	if _, typeName, err = GetIdValueAndStructureName(s); err != nil {
		return err
	}
	//update all the extra keys using writeKey
	err= self.walkAllExtraKeys(s, func(n string, v string) error {
		return self.deleteKey(s, n, typeName, v, id)
	})
	
	if err!=nil && err!=INDEX_MISS {
		return err
	}
	key := fmt.Sprintf(RECKEY, typeName, id)
	memcache_err=self.Client.Delete(key)
	
	if memcache_err == memcache.ErrCacheMiss && err==INDEX_MISS {
		return memcache.ErrCacheMiss
	}
	
	if err==INDEX_MISS {
		panic(fmt.Sprintf("indexes are out of sync with data: %d not found",id))
	}
	
	return memcache_err
	
	
}
