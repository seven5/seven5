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
	//write the extra indexes
	fields, methods := GetStructKeys(s)
	var mapKey string

	//try the fields
	for _, k := range fields {
		//does the FIELD have a string method?
		mapKey = toStringMethodOrSprintf(k.Value)
		err = self.writeKey(s, k.Name, typeName, mapKey, value)
		if err != nil {
			return err
		}
	}
	for _, m := range methods {
		out := m.Meth.Call([]reflect.Value{})
		mapKey = toStringMethodOrSprintf(out[0])
		//hacky way to get this to be a "real" string not a value
		err = self.writeKey(s, m.Name, typeName, mapKey, value)
	}
	return nil
}

func toStringMethodOrSprintf(v reflect.Value) string {
	method := v.MethodByName("String")
	if method == ZeroValue {
		//why do I need to elucidate all the types here?
		if v.Kind()==reflect.Int {
			return fmt.Sprintf("%d",v.Int())
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
	//compute memcached key
	key := fmt.Sprintf(EXTRAKEY, typeName, keyName)

	fmt.Printf("key is %s [%s] '%d'\n",key,mapKey,id)

	item, err := self.Client.Get(key)

	//map from values to ids of objects
	var index map[string][]uint64

	//if not there, create the map from scratch
	if err == memcache.ErrCacheMiss {
		index = make(map[string][]uint64)
		item = nil
	} else if err == nil {
		//no error, read the map
		buffer := bytes.NewBuffer(item.Value)
		decoder := gob.NewDecoder(buffer)
		err = decoder.Decode(&index)
		if err != nil {
			return err
		}
	} else {
		//some other error, give up
		return err
	}
	//ok, if we get here, we are ok and have the index loaded (or created)...
	index[mapKey] = append(index[mapKey], id)

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
	item = &memcache.Item{Key: key, Value: buffer.Bytes()}
	return self.Client.Set(item)
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

	//read index
	key := fmt.Sprintf(EXTRAKEY, typeName, keyName)
	if item, err = self.Client.Get(key); err != nil {
		return err
	}

	var index map[string][]uint64

	//no error, read the index
	buffer := bytes.NewBuffer(item.Value)
	decoder := gob.NewDecoder(buffer)
	err = decoder.Decode(&index)
	if err != nil {
		return err
	}

	//get the item we are interested in
	_, ok := index[value]
	if !ok {
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

func (self *MemcacheGobStore) DeleteById(s interface{},id uint64) error {
	var typeName string
	var err error
	
	if _, typeName, err = GetIdValueAndStructureName(s); err != nil {
		return err
	}
	key:=fmt.Sprintf(RECKEY,typeName,id)
	return self.Client.Delete(key)
}