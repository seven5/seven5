package store

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/bradfitz/gomemcache"
	"net"
	//"os"
	"reflect"
	"strconv"
	"sort"
)

type MemcacheGobStore struct {
	*memcache.Client
}

type pointerToObjSlice struct {
	S          Lesser
	Mem        *MemcacheGobStore
	SliceValue reflect.Value
}

const (
	LOCALHOST = "localhost:11211"
	IDKEY     = "%s-idcounter"
	RECKEY    = "%s-%d"
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
	var id uint64
	var typeName string
	var err error

	if id, typeName, err = GetIdValueAndStructureName(s); err != nil {
		return err
	}

	if err = self.writeItemData(s, id, typeName); err != nil {
		return err
	}

	newValueOfId := reflect.ValueOf(s).Elem().FieldByName("Id")
	userId := uint64(0)
	if self.updateUserIdWithOwner(s, &userId) == true {
		if userId == uint64(0) {
			panic("You have an Owner field but tried to write a value without setting Owner!")
		}
	}	
	//horrible hack to get out of value space and into an actual uint via string
	stringified:=fmt.Sprintf("%d",newValueOfId.Uint())
	hacked,err:=strconv.ParseUint(stringified, 10, 64)
	if err!=nil {
		panic(fmt.Sprintf("here, can't parse:%d %T %d",newValueOfId,newValueOfId,newValueOfId.Uint()))
		return err
	}
	//update all the extra keys using writeKey
	err = self.walkAllExtraKeys(s, func(n string, v string, isFifo keyOrder) error {
		return self.writeKey(s, n, typeName, v, hacked, nil, isFifo, userId)
	})

	return nil
}

//fieldFunc is the type of function you have to supply to walkAllExtraKeys which will get called
//for each key of the structure (or methods, if declared in the structure defn)
type fieldFunc func(fieldOrMethodName string, flattenedValue string, isFifo keyOrder) error

//walkAllExtraKeys is a utility route to run a function for every key mentioned in the structure
//declaration of s.  This function assumes that s has already been checked an is properly
//formed.
func (self *MemcacheGobStore) walkAllExtraKeys(s interface{}, fn fieldFunc) error {
	//write the extra indexes
	fields, methods := GetStructKeys(s)
	var mapKey string
	var err error

	//try the fields
	for _, k := range fields {
		if k.Name == MAGIC_KEY {
			mapKey = MAGIC_VALUE
		} else {
			mapKey = toStringMethodOrSprintf(k.Value)
		}
		err = fn(k.Name, mapKey, k.IsFifo)
		if err != nil {
			return err
		}
	}
	//now the methods
	for _, m := range methods {
		out := m.Meth.Call([]reflect.Value{})
		mapKey = toStringMethodOrSprintf(out[0])
		err = fn(m.Name, mapKey, m.IsFifo)
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
		switch v.Kind() {
		case reflect.Int64, reflect.Int16, reflect.Int32, reflect.Int8, reflect.Int:
			return fmt.Sprintf("%d",v.Int())
		case reflect.Uint64, reflect.Uint16, reflect.Uint32, reflect.Uint8, reflect.Uint:
			return fmt.Sprintf("%d",v.Uint())
		case reflect.String:
			return fmt.Sprintf("%s",v.String())
		default:
			panic(fmt.Sprintf("not implemented flattener for that type yet (toStringMethodOrSprintf): %v",v.Kind()))
		}
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
	return self.DecodeItemBytes(s, item)
}

//decodeItemBytes returns a structured object from the gob blob of bytes. It assumes that s
//is a pointer a structure that has already been checked.
func (self *MemcacheGobStore) DecodeItemBytes(s interface{}, item *memcache.Item) error {
	buffer := bytes.NewBuffer(item.Value)
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(s)
}

//writeKey assumes that the pointer to struct has already been checked and is ok.
//writeKey tries to find a map that it can use to index the records by the value
//provided.  it creates the map if necessary.  The map is per keyName with the keys
//of the map being values and the values are ids.
func (self *MemcacheGobStore) writeKey(s interface{}, keyName string, typeName string, mapKey string, id uint64, fullIndex []uint64, isFifo keyOrder, userId uint64) error {
	var index []uint64
	var item *memcache.Item
	var err error

	if err = self.readIndex(mapKey, &index, &item, typeName, keyName, true, userId); err != nil {
		return err
	}

	//ok, if we get here, we are ok and have the index loaded (or created)...
	if len(fullIndex) == 0 {
		//we are adding a single item
		if isFifo != keyOrder(LIFO_ORDER) {
			index = append(index, id)
		} else {
			index = append([]uint64{id}, index...)
		}
		err = self.writeIndex(mapKey, index, item, typeName, keyName, userId)
	} else {
		err = self.writeIndex(mapKey, fullIndex, item, typeName, keyName, userId)
	}

	return err
}

//writeIndex puts an index in memcached by serializing it with gobs.  it needs to be passed 
//back the same item value that was returned from readIndex so we can correctly detect 
//concurrency problems. if item is nil it assumes that this is a creation (and no concurrency check needed)
func (self *MemcacheGobStore) writeIndex(keyValue string, indexValue []uint64, item *memcache.Item, typeName string, keyName string, userId uint64) error {
	//serialize to gob the index
	buffer := new(bytes.Buffer)
	enc := gob.NewEncoder(buffer)
	if err := enc.Encode(indexValue); err != nil {
		return err
	}

	//did we have anything in the index before? if so, try to swap in new value but abort
	//if that fails
	if item != nil {
		item.Value = buffer.Bytes()
		return self.Client.CompareAndSwap(item)
	}
	//this is a brand new index, write it out to disk
	memcacheKey := self.getKeyNameForRecord(typeName, keyName, keyValue, userId)
	newItem := &memcache.Item{Key: memcacheKey, Value: buffer.Bytes()}
	return self.Client.Set(newItem)
}

//readIndex pulls an index from memcached and sets the first parameter to it, if there was no error.
//the create flag indicates if not finding the item in memcached is an error or it should be
//created (true).  The item parameter is set to the retrieved Item object for use later
//with compareAndSwap
func (self *MemcacheGobStore) readIndex(keyValue string, result *[]uint64, item **memcache.Item, typeName string, keyName string, create bool, userId uint64) error {
	//compute memcached key
	key := self.getKeyNameForRecord(typeName, keyName, keyValue, userId)
	var err error

	*item, err = self.Client.Get(key)

	//if not there, create the map from scratch, based on create param
	if err == memcache.ErrCacheMiss {
		if create {
			*result = []uint64{}
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

//deleteKey does the work to update an index when an item is deleted and preserve the
//lifo or fifo ordering (by not changing it)
func (self *MemcacheGobStore) deleteKey(s interface{}, keyName string, typeName string, mapKey string, id uint64, userId uint64) error {
	var slice []uint64
	var item *memcache.Item

	if err := self.readIndex(mapKey, &slice, &item, typeName, keyName, true, userId); err != nil {
		return err
	}
	ok := false
	for i := 0; i < len(slice); i++ {
		if slice[i] == id {
			if i == len(slice)-1 {
				slice = slice[0:i]
			} else {
				slice = append(slice[0:i], slice[i+1:]...)
			}
			ok = true
			break
		}
	}
	if !ok {
		return INDEX_MISS
	}
	return self.writeIndex(mapKey, slice, item, typeName, keyName, userId)
}

//FindByKey looks up a value in the memcache by a field _other_ than the Id field.  You have
//to supply the name of the field.  Further, that field must exist in the structure, be
//exported (uppercase).  The value field must match exactly the flattened version of the value
//when String() is called on it (at the time it is written). 
//This call results in n+1 roundtrips to the memcache server for n values because
//it first retrieves the ids of the objects that are stored under the value provided and
//then calls FindById on each one.  The result is placed the slice pointed to by the first
//element--and only as many results are returned as available places in the slice.
func (self *MemcacheGobStore) FindByKey(ptrToResult interface{}, keyName string, value string, userId uint64) (errReturn error) {
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
	var slice []uint64
	if err = self.readIndex(value, &slice, &item, typeName, keyName, false, userId); err != nil {
		//no key with that value at all?
		if err == memcache.ErrCacheMiss {
			return nil
		}
		return err
	}
	//had it before, but not now?
	if len(slice) == 0 {
		//just tell the caller there is nothing with that value
		return nil
	}

	//walk through the type structure and generate an exemplar so we can examine it for
	//the fields and structure tages
	e := result.Type().Elem().Elem()
	example := reflect.New(e).Interface()

	ok := false
	//we need to see if they specified any order for this key
	f, m := GetStructKeys(example)
	order := keyOrder(UNSPECIFIED_ORDER)
	for _, fld := range f {
		if fld.Name == keyName {
			order = fld.IsFifo
			ok = true
			break
		}
	}
	if !ok {
		for _, meth := range m {
			if meth.Name == keyName {
				order = meth.IsFifo
				ok = true
				break
			}
		}
	}
	if !ok {
		panic(fmt.Sprintf("unable to find key that is being used in FindByKey (%s)", keyName))
	}

	//can we do this with readMulti?
	if order == keyOrder(UNSPECIFIED_ORDER) {
		self.readMulti(example, slice, result)
	} else {
		ct := result.Len()
		//we have to read them in order to make sure we preserve the order in the slice
		for _, id := range slice {
			if ct == result.Cap() {
				break
			}
			newObject := reflect.New(e)
			ptr := newObject.Interface()
			if err = self.FindById(ptr, id); err != nil {
				return err
			}
			result.SetLen(ct + 1)
			result.Index(ct).Set(newObject)
			ct++
		}
	}
	//try to sort the result, if there is not the proper sort function it has no effect
	self.sort(example, result)

	return nil
}

//readMulti is only used when we have found a key and need to load the actual objects that
//are stored (as Ids, not values) under that key.
func (self *MemcacheGobStore) readMulti(s interface{}, ids []uint64, result reflect.Value) (errReturn error) {
	min := len(ids)
	avail := result.Cap() - result.Len()
	if avail < min {
		min = avail
	}
	key := make([]string, min, min)
	var err error
	var item map[string]*memcache.Item
	var typeName string

	if _, typeName, err = GetIdValueAndStructureName(s); err != nil {
		return err
	}
	for i, v := range ids {
		if i == min {
			break
		}
		key[i] = fmt.Sprintf(RECKEY, typeName, v)
	}
	//we are now sure that the key slice is the right size to fit the result
	item, err = self.Client.GetMulti(key)
	if err != nil {
		return err
	}
	ct := result.Len()
	method := reflect.ValueOf(self).MethodByName("DecodeItemBytes")

	for _, v := range item {
		x := reflect.New(result.Type().Elem().Elem())
		in := []reflect.Value{x, reflect.ValueOf(v)}
		out := method.Call(in)
		if !out[0].IsNil() {
			rtn := reflect.ValueOf(&errReturn)
			reflect.Indirect(rtn).Set(out[0])
			return
		}
		result.SetLen(ct + 1)
		result.Index(ct).Set(x)
		ct++
	}
	return nil
}

//DeleteById deletes an item from the store by its unique id.  This does the work to update all
//the index.  The only fields that are examined are the Id field and (optionally) the
//Owner field.  If there *is* an Owner field, it must be set.
func (self *MemcacheGobStore) Delete(s interface{}) error {
	var typeName string
	var err, memcache_err error
	var id, userId uint64

	if id, typeName, err = GetIdValueAndStructureName(s); err != nil {
		return err
	}

	if self.updateUserIdWithOwner(s, &userId) == true {
		if userId == uint64(0) {
			panic("tried to delete a record, but the Owner field is set to zero!")
		}
	}

	//update all the extra keys using writeKey
	err = self.walkAllExtraKeys(s, func(n string, v string, IGNORED keyOrder) error {
		return self.deleteKey(s, n, typeName, v, id, userId)
	})

	if err != nil && err != INDEX_MISS {
		return err
	}
	key := fmt.Sprintf(RECKEY, typeName, id)
	memcache_err = self.Client.Delete(key)

	if memcache_err == memcache.ErrCacheMiss && err == INDEX_MISS {
		return memcache.ErrCacheMiss
	}

	if err == INDEX_MISS {
		panic(fmt.Sprintf("indexes are out of sync with data: %d not found", id))
	}

	return memcache_err
}

//updateUserIdWithOwner will change the uint64 passed as 2nd arg to the value of the 
//Owner field in the structure pointed to by first arg, if the Owner field is present.
func (self *MemcacheGobStore) updateUserIdWithOwner(s interface{}, userId *uint64) bool {
	return self.updateUserIdWithOwnerValue(reflect.ValueOf(s), userId)
}

//updateUserIdWithOwnerValue will change the uint64 passed as 2nd arg to the value of the 
//Owner field in the structure pointed to by first arg, if the Owner field is present.
func (self *MemcacheGobStore) updateUserIdWithOwnerValue(s reflect.Value, userId *uint64) bool {
	ownerValue := s.Elem().FieldByName("Owner")
	if ownerValue != ZeroValue {
		*userId = ownerValue.Uint()
		return true
	}
	return false
}

//Init sets up the store to be ready to receive objects of this type.  This is useful if you
//want to allow reads() before you have had any writes.  Note that you must fill in the owner
//field if you plan to use it!
func (self *MemcacheGobStore) Init(s interface{}) error {
	var typeName string
	var err error

	if _, typeName, err = GetIdValueAndStructureName(s); err != nil {
		return err
	}

	item := &memcache.Item{Key: fmt.Sprintf(IDKEY, typeName), Value: []byte("0")}
	if err = self.Client.Set(item); err != nil {
		return err
	}

	userId := uint64(0)
	if self.updateUserIdWithOwner(s, &userId) == true {
		if userId == uint64(0) {
			panic("You have an Owner field, but called it with the field set to zero!")
		}
	}
	return self.walkAllExtraKeys(s, func(n string, v string, IGNORED keyOrder) error {
		empty := []uint64{}
		var item *memcache.Item
		if e := self.readIndex(v, &empty, &item, typeName, n, true, userId); e != nil {
			return e
		}
		return self.writeIndex(v, empty, item, typeName, n, userId)
	})
}

//FindAll returns all the items of the particular type that is pointed to by the first parameter
//in a slice, up to the capacity of the slice. This returns a nil (a length 0 result) if you have 
//chose to turn off this feature with seven5order:"none"
func (self *MemcacheGobStore) FindAll(s interface{}, userId uint64) error {
	return self.FindByKey(s, MAGIC_KEY, MAGIC_VALUE, userId)
}

//sort can take a slice and sort it (second param) if the first parameter has the method
//Less() and the types of the pointers are the same between the first parameter and the
//slice of pointers that is the second one.
func (self *MemcacheGobStore) sort(x interface{}, sliceValue reflect.Value) error {
	if reflect.TypeOf(x) != sliceValue.Index(0).Type() {
		panic(fmt.Sprintf("types are not the same between %v and %v", reflect.TypeOf(x), sliceValue.Index(0).Type()))
	}
	s, ok := x.(Lesser)
	if !ok {
		return nil
	}

	sortable := &pointerToObjSlice{s, self, sliceValue}
	sort.Sort(sortable)
	return nil
}

//BulkWrite repeatedly writes the raw structure object to memcached but does not update the
//indexes until the end.  Note that if you have an Owner field it must be set and that it
//must be the same for all values written together in a single call to this method.
func (self *MemcacheGobStore) BulkWrite(sliceOfPtrs interface{}) error {
	var id uint64
	var typeName string
	var err error
	var userId uint64
	var s interface{}

	master := make(map[string]map[string][]uint64)

	v := reflect.ValueOf(sliceOfPtrs)
	if v.Kind() != reflect.Slice {
		return BAD_SLICE
	}
	//nothing to do
	if v.Len() == 0 {
		return nil
	}
	// ok, don't repeat that work
	for i := 0; i < v.Len(); i++ {
		valueOfItem := v.Index(i)
		s = valueOfItem.Interface()
		if id, typeName, err = GetIdValueAndStructureNameFromValue(valueOfItem); err != nil {
			return err
		}
		o := uint64(0)
		if self.updateUserIdWithOwnerValue(valueOfItem, &o) == true {
			if o == uint64(0) {
				panic("Structure has an Owner field but it is not set!")
			}
			if userId != uint64(0) && o != userId {
				panic("Structure has an Owner field but different owners inside a bulk write!")
			}
			userId = o
		}
		if err = self.writeItemData(s, id, typeName); err != nil {
			return err
		}
		newValueOfId := valueOfItem.Elem().FieldByName("Id").Uint()
		err = self.walkAllExtraKeys(s, func(n string, v string, isFifo keyOrder) error {
			index := master[n]
			if index == nil {
				index = make(map[string][]uint64)
				master[n] = index
			}
			if isFifo != LIFO_ORDER {
				index[v] = append(index[v], newValueOfId)
			} else {
				index[v] = append([]uint64{newValueOfId}, index[v]...)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	if typeName == "" || userId == uint64(0) || s == nil {
		panic("did not do any bulk writing! did not set typeName or userId!")
	}

	for keyName, maps := range master {
		for mapKey, index := range maps {
			order := keyOrder(UNSPECIFIED_ORDER) /*already dealt with this issue in the loops above*/
			if err := self.writeKey(s, keyName, typeName, mapKey, 0, index, order, userId); err != nil {
				return err
			}
		}
	}
	return nil
}

//writeItemData knows how to write the fields of a structure to memcache, assigning
//a new Id number if necessary.  this is a primitive for other types of write.
func (self *MemcacheGobStore) writeItemData(s interface{}, id uint64, typeName string) error {
	var err error

	if id == uint64(0) {
		nextId, err := self.GetNextId(typeName)
		if err != nil {
			return err
		}
		newId := reflect.ValueOf(s).Elem().FieldByName("Id")
		newId.SetUint(nextId)
		id = nextId
	}
	//at this point id holds the number of the record
	key := fmt.Sprintf(RECKEY, typeName, id)
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
	return nil
}

//getRecordKey returns a key string for a given set of information about the key.  This is
//used to create different keys if you supply a userId (!=0) to implement a bit of segregation
//by owner.
func (self *MemcacheGobStore) getKeyNameForRecord(typeName string, keyName string, keyValue string, userId uint64) string {
	if fmt.Sprintf("%s",keyValue)=="<uint64 Value>" {
		panic("here")
	}
	if userId == uint64(0) {
		return fmt.Sprintf("%s-key-%s-val-%s", typeName, keyName, keyValue)
	}
	return fmt.Sprintf("%s-key-%s-user-%d-val-%s", typeName, keyName, userId, keyValue)
}

//
// SORTING CRUFT
//

//Len returns the size of the slice
func (self *pointerToObjSlice) Len() int {
	return self.SliceValue.Len()
}

//Swap exchanges two elements of the slice
func (self *pointerToObjSlice) Swap(i, j int) {
	tmp := reflect.New(self.SliceValue.Index(i).Type())
	reflect.Indirect(tmp).Set(self.SliceValue.Index(i))
	self.SliceValue.Index(i).Set(self.SliceValue.Index(j))
	self.SliceValue.Index(j).Set(reflect.Indirect(tmp))
	self.S.Less(self.SliceValue.Index(i), self.SliceValue.Index(j))
}

//Less uses delegation to underlying type
func (self *pointerToObjSlice) Less(i, j int) bool {
	return self.S.Less(self.SliceValue.Index(i), self.SliceValue.Index(j))
}
