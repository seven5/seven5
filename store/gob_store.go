package store

import (
	"bytes"
	"encoding/gob"
	"fmt"
	//"os"
	"github.com/petar/GoLLRB/llrb"
	"reflect"
	"sort"
	"strconv"
	"time"
)

//init function registers our type with the gobs code so we can use gobs with LLRB
func init() {
	//so we can use gobs with the llrb tree nodes
	var y ValueInfo
	gob.Register(y)
}

//GobStore is the implementation of the Store.T API that stores everything with enc/gob.  This
//object does not care what the underlying key-value implementation is.
type GobStore struct {
	impl StoreImpl
}

//NewGobStore returns a GobStore pointed at the given key-value store (StoreImpl)
func NewGobStore(storeImpl StoreImpl) *GobStore {
	return &GobStore{storeImpl}
}

//pointerToObjSlice is used by the sorting routines that sort the result of a query.
type pointerToObjSlice struct {
	S          Lesser
	Mem        *GobStore
	SliceValue reflect.Value
}

const (
	IDKEY     = "%s-idcounter"
	RECKEY    = "%s-%d"
	VALUEKEY  = "%s-value-%s"
)

//LessUint64 is a comparison function used to compare uint64s for the GoLLRB datastructure
//to keep the nodes of it's RB tree in order.
func LessUint64(a,b interface{}) bool {
	x:=a.(uint64)
	y:=b.(uint64)
	return x<y
}


//GetNextId will return the next id for the type X.  It will create the necessary structures in
//the store if that is needed.
func (self *GobStore) GetNextId(typeName string) (uint64, error) {
	key := fmt.Sprintf(IDKEY, typeName)
	newValue, err := self.impl.Increment(key, uint64(1))
	if err == nil {
		return newValue, nil
	}
	if err != ErrorNotFoundInStore {
		return uint64(0), err
	}

	item := NewStoredItem()
	item.SetKey(key)
	item.SetValue([]byte("0"))
	
	err = self.impl.Add(item)
	if err == ErrorNotFoundInStore {
		//try a 2nd time, maybe race condition
		newValue, err = self.impl.Increment(key, uint64(1))
		if err != nil {
			return uint64(0), err
		}
		return newValue, nil
	}
	if err != nil {
		return uint64(0), err
	}
	//try to do increment again since we set it successfully
	result, err := self.impl.Increment(key, uint64(1))

	if err != nil {
		return uint64(0), err
	}
	return result, nil
}

//Write takes a structure and sends it to store  If the id field is not set yet, it creates
//an Id for the item before writing it to sotre.  The value passed must be a pointer to a 
//struct and the struct must have an Id field that is uint64.
func (self *GobStore) Write(s interface{}) error {
	var id uint64
	var typeName string
	var err error

	if id, typeName, err = getIdValueAndStructureName(s); err != nil {
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
	stringified := fmt.Sprintf("%d", newValueOfId.Uint())
	hacked, err := strconv.ParseUint(stringified, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("here, can't parse:%d %T %d", newValueOfId, newValueOfId, newValueOfId.Uint()))
		return err
	}
	//update all the extra keys using writeKey
	err = self.walkAllExtraKeys(s, func(n string, v string, isFifo keyOrder) error {
		if e := self.writeKey(s, n, typeName, v, hacked, nil, isFifo, userId); e != nil {
			return e
		}
		return self.addUniqueKey(typeName, n, v, userId, isFifo)
	})

	return nil
}

//fieldFunc is the type of function you have to supply to walkAllExtraKeys which will get called
//for each key of the structure (or methods, if declared in the structure defn)
type fieldFunc func(fieldOrMethodName string, flattenedValue string, isFifo keyOrder) error

//walkAllExtraKeys is a utility route to run a function for every key mentioned in the structure
//declaration of s.  This function assumes that s has already been checked an is properly
//formed.
func (self *GobStore) walkAllExtraKeys(s interface{}, fn fieldFunc) error {
	//write the extra indexes
	fields, methods := getStructKeys(s)
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
			return fmt.Sprintf("%d", v.Int())
		case reflect.Uint64, reflect.Uint16, reflect.Uint32, reflect.Uint8, reflect.Uint:
			return fmt.Sprintf("%d", v.Uint())
		case reflect.String:
			return fmt.Sprintf("%s", v.String())
		default:
			panic(fmt.Sprintf("not implemented flattener for that type yet (toStringMethodOrSprintf): %v", v.Kind()))
		}
	}
	return method.Call(nil)[0].String()
}

//FindById is the reverse of Write and reads a structure from memcached for a given type and Id.
//The contents of the first parameter will be overwritten by this call and the previous contents
//ignored (although its type must be a pointer to a struct of the right type)
func (self *GobStore) FindById(s interface{}, id uint64) error {
	var typeName string
	var err error

	if _, typeName, err = getIdValueAndStructureName(s); err != nil {
		return err
	}
	key := fmt.Sprintf(RECKEY, typeName, id)
	var item StoredItem
	if item, err = self.impl.Get(key); err != nil {
		return err
	}
	return self.DecodeItemBytes(s, item)
}

//DecodeItemBytes returns a structured object from the gob blob of bytes. It assumes that s
//is a pointer a structure that has already been checked.  This has to be public for now,
//as it is called via the reflection interface which refuses to invoke non-exported functions.
func (self *GobStore) DecodeItemBytes(s interface{}, item StoredItem) error {
	buffer := bytes.NewBuffer(item.Value())
	decoder := gob.NewDecoder(buffer)
	return decoder.Decode(s)
}

//writeKey assumes that the pointer to struct has already been checked and is ok.
//writeKey tries to find a map that it can use to index the records by the value
//provided.  it creates the map if necessary.  The map is per keyName with the keys
//of the map being values and the values are ids.
func (self *GobStore) writeKey(s interface{}, keyName string, typeName string, mapKey string, id uint64, fullIndex []uint64, isFifo keyOrder, userId uint64) error {
	var index []uint64
	var item StoredItem
	var err error

	if err = self.readIndex(mapKey, &index, &item, typeName, keyName, true, userId); err != nil {
		return err
	}

	//ok, if we get here, we are ok and have the index loaded (or created)...
	if len(fullIndex) == 0 {
		//we are adding a single item
		if isFifo != keyOrder(lifo_order) {
			index = append(index, id)
		} else {
			index = append([]uint64{id}, index...)
		}
		err = self.writeIndex(mapKey, index, item, typeName, keyName, userId)
	} else {
		//this is the case of a bulk write, order is handled elsewhere
		err = self.writeIndex(mapKey, fullIndex, item, typeName, keyName, userId)
	}

	return err
}

//writeIndex puts an index in memcached by serializing it with gobs.  it needs to be passed 
//back the same item value that was returned from readIndex so we can correctly detect 
//concurrency problems. if item is nil it assumes that this is a creation (and no concurrency check needed)
func (self *GobStore) writeIndex(keyValue string, indexValue []uint64, item StoredItem, typeName string, keyName string, userId uint64) error {
	//serialize to gob the index
	buffer := new(bytes.Buffer)
	enc := gob.NewEncoder(buffer)
	if err := enc.Encode(indexValue); err != nil {
		return err
	}

	//did we have anything in the index before? if so, try to swap in new value but abort
	//if that fails
	if item != nil {
		item.SetValue(buffer.Bytes())
		return self.impl.CompareAndSwap(item)
	}
	
	//this is a brand new index, write it out to disk
	memcacheKey := self.getKeyNameForRecord(typeName, keyName, keyValue, userId)
	newItem := NewStoredItem()
	newItem.SetKey(memcacheKey)
	newItem.SetValue(buffer.Bytes())
	return self.impl.Set(newItem)
}

//readIndex pulls an index from memcached and sets the first parameter to it, if there was no error.
//the create flag indicates if not finding the item in memcached is an error or it should be
//created (true).  The item parameter is set to the retrieved Item object for use later
//with compareAndSwap
func (self *GobStore) readIndex(keyValue string, result *[]uint64, item *StoredItem, typeName string, keyName string, create bool, userId uint64) error {
	//compute memcached key
	key := self.getKeyNameForRecord(typeName, keyName, keyValue, userId)
	var err error

	*item, err = self.impl.Get(key)

	//if not there, create the map from scratch, based on create param
	if err == ErrorNotFoundInStore {
		if create {
			*result = []uint64{}
			*item = nil
		} else {
			return err
		}
	} else if err == nil {
		//no error, read the map
		buffer := bytes.NewBuffer((*item).Value())
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
func (self *GobStore) deleteKey(s interface{}, keyName string, typeName string, mapKey string, id uint64, userId uint64, isFifo keyOrder) error {
	var slice []uint64
	var item StoredItem

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
			if len(slice)==0 {
				if e:=self.deleteUniqueKey(typeName, keyName, mapKey, userId, isFifo); e!=nil {
					return e
				}
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
//when String() is called on it (which happened at the time it is written) or it must be a basic type. 
//
// The result is placed the slice pointed to by the first
//element--and only as many results are returned as available places in the slice.
func (self *GobStore) FindByKey(ptrToResult interface{}, keyName string, value string, userId uint64) (errReturn error) {
	var err error
	if reflect.TypeOf(ptrToResult).Kind() != reflect.Ptr {
		return BAD_SLICE_PTR
	}
	result := reflect.ValueOf(ptrToResult).Elem()

	if result.Type().Kind() != reflect.Slice {
		return BAD_SLICE
	}

	s := result.Type().Elem()
	if err = verifyStructPointerFieldTypes(s); err != nil {
		return err
	}
	typeName := s.Elem().String()

	//walk through the type structure and generate an exemplar so we can examine it for
	//the fields and structure tages
	e := result.Type().Elem().Elem()
	example := reflect.New(e).Interface()

	//nothing tricky with the userId?
	_,hasOwner := e.FieldByName("Owner")
	if !hasOwner || userId!=uint64(0) { 
		return self.findByKeyInternal(ptrToResult,e,typeName,keyName,value,userId)
	}
	
	//load up to 256 key values
	pairs:=make([]ValueInfo,0,256)
	if err=self.UniqueKeyValues(example,keyName, &pairs, uint64(0)); err!=nil {
		return err
	}
	
	mergedResult := make([]uint64,0,256)
	tree:=llrb.New(LessUint64)
	
	var ignored StoredItem

	for _, v:=range pairs {
		
		if v.Value!=value {
			continue
		}	
		ids:=make([]uint64,0,256)	
		if err=self.readIndex(v.Value,&ids,&ignored,typeName,keyName,false,v.Owner);err!=nil {
			return err
		}
		for _, id := range ids {
			if tree.Get(id)==nil {
				mergedResult = append(mergedResult,id)
				tree.InsertNoReplace(id)
			}
		}
	}
	return self.fillResult(result,  mergedResult, e)
}

//findByKeyInternal does the work of a single search of the store.  It assumes that the validity
//checking of the types (such as first parameter) has already been done by the caller.
func (self *GobStore) findByKeyInternal(ptrToResult interface{}, resultItemType reflect.Type, typeName string, keyName string, value string, userId uint64) (errReturn error) {
	var item StoredItem
	var slice []uint64
	var err error
	result := reflect.ValueOf(ptrToResult).Elem()
	
	if err = self.readIndex(value, &slice, &item, typeName, keyName, false, userId); err != nil {
		//no key with that value at all?
		if err == ErrorNotFoundInStore {
			return nil
		}
		return err
	}
	//had it before, but not now?
	if len(slice) == 0 {
		//just tell the caller there is nothing with that value
		return nil
	}

	ok := false
	example:=reflect.New(resultItemType).Interface()
	//we need to see if they specified any order for this key
	f, m := getStructKeys(example)
	order := keyOrder(unspecified_order)
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
	if order == keyOrder(unspecified_order) {
		self.readMulti(example, slice, result)
	} else {
		/*
		ct := result.Len()
		//we have to read them in order to make sure we preserve the order in the slice
		for _, id := range slice {
			if ct == result.Cap() {
				break
			}
			newObject := reflect.New(resultItemType)
			ptr := newObject.Interface()
			if err = self.FindById(ptr, id); err != nil {
				return err
			}
			result.SetLen(ct + 1)
			result.Index(ct).Set(newObject)
			ct++
		}*/
		if err=self.fillResult(result,slice,resultItemType); err!=nil {
			return err
		}
	}
	//try to sort the result, if there is not the proper sort function it has no effect
	self.sort(example, result)

	return nil
}

//fillResult will fill in a result with the values of the ids provided, up to the limit of
//of the first parameter's capacity.  It assumes that result is a pointer to a slice that
//has already been checked too if the items in the slice (pointed to) are ok.  The second
//parameter is a slice of the ids to load.
func (self *GobStore) fillResult(result reflect.Value, slice []uint64, resultItemType reflect.Type) error {
	var err error
	ct := result.Len()
	//we have to read them in order to make sure we preserve the order in the slice
	for _, id := range slice {
		if ct == result.Cap() {
			break
		}
		newObject := reflect.New(resultItemType)
		ptr := newObject.Interface()
		if err = self.FindById(ptr, id); err != nil {
			return err
		}
		result.SetLen(ct + 1)
		result.Index(ct).Set(newObject)
		ct++
	}
	return nil
}

//readMulti is only used when we have found a key and need to load the actual objects that
//are stored (as Ids, not values) under that key.
func (self *GobStore) readMulti(s interface{}, ids []uint64, result reflect.Value) (errReturn error) {
	min := len(ids)
	avail := result.Cap() - result.Len()
	if avail < min {
		min = avail
	}
	key := make([]string, min, min)
	var err error
	var item map[string]StoredItem
	var typeName string


	if _, typeName, err = getIdValueAndStructureName(s); err != nil {
		return err
	}
	for i, v := range ids {
		if i == min {
			break
		}
		key[i] = fmt.Sprintf(RECKEY, typeName, v)
	}
	//we are now sure that the key slice is the right size to fit the result
	item, err = self.impl.GetMulti(key)
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
func (self *GobStore) Delete(s interface{}) error {
	var typeName string
	var err, memcache_err error
	var id, userId uint64

	if id, typeName, err = getIdValueAndStructureName(s); err != nil {
		return err
	}

	if self.updateUserIdWithOwner(s, &userId) == true {
		if userId == uint64(0) {
			panic("tried to delete a record, but the Owner field is set to zero!")
		}
	}

	//update all the extra keys using deleteKey
	err = self.walkAllExtraKeys(s, func(n string, v string, isFifo keyOrder) error {
		return self.deleteKey(s, n, typeName, v, id, userId, isFifo)
	})

	if err != nil && err != INDEX_MISS {
		return err
	}
	key := fmt.Sprintf(RECKEY, typeName, id)
	memcache_err = self.impl.Delete(key)

	if memcache_err == ErrorNotFoundInStore && err == INDEX_MISS {
		return ErrorNotFoundInStore
	}

	if err == INDEX_MISS {
		panic(fmt.Sprintf("indexes are out of sync with data: %d not found", id))
	}

	return memcache_err
}

//updateUserIdWithOwner will change the uint64 passed as 2nd arg to the value of the 
//Owner field in the structure pointed to by first arg, if the Owner field is present.
func (self *GobStore) updateUserIdWithOwner(s interface{}, userId *uint64) bool {
	return self.updateUserIdWithOwnerValue(reflect.ValueOf(s), userId)
}

//updateUserIdWithOwnerValue will change the uint64 passed as 2nd arg to the value of the 
//Owner field in the structure pointed to by first arg, if the Owner field is present.
func (self *GobStore) updateUserIdWithOwnerValue(s reflect.Value, userId *uint64) bool {
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
func (self *GobStore) Init(s interface{}) error {
	var typeName string
	var err error

	if _, typeName, err = getIdValueAndStructureName(s); err != nil {
		return err
	}

	item := NewStoredItem()
	item.SetKey(fmt.Sprintf(IDKEY, typeName))
	item.SetValue([]byte("0"))
	if err = self.impl.Set(item); err != nil {
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
		var item StoredItem
		if e := self.readIndex(v, &empty, &item, typeName, n, true, userId); e != nil {
			return e
		}
		if e:=self.writeIndex(v, empty, item, typeName, n, userId); e!=nil {
			return e
		}
		return self.writeEmptyTree(typeName,n)
	})
}

//FindAll returns all the items of the particular type that is pointed to by the first parameter
//in a slice, up to the capacity of the slice. This returns a nil (a length 0 result) if you have 
//chose to turn off this feature with seven5order:"none"
func (self *GobStore) FindAll(s interface{}, userId uint64) error {
	return self.FindByKey(s, MAGIC_KEY, MAGIC_VALUE, userId)
}

//sort can take a slice and sort it (second param) if the first parameter has the method
//Less() and the types of the pointers are the same between the first parameter and the
//slice of pointers that is the second one.
func (self *GobStore) sort(x interface{}, sliceValue reflect.Value) error {
	if sliceValue.Len()==0 {
		return nil //no sorting to do!
	}
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
func (self *GobStore) BulkWrite(sliceOfPtrs interface{}) error {
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
		if id, typeName, err = getIdValueAndStructureNameFromValue(valueOfItem); err != nil {
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
			if isFifo != lifo_order {
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

	if typeName == "" ||  s == nil {
		panic(fmt.Sprintf("did not do any bulk writing! did not set typeName[%v] or s[%v]!",typeName,userId,s))
	}

	for keyName, maps := range master {
		unique:=[]ValueInfo{}
		for mapKey, index := range maps {
			order := keyOrder(unspecified_order) /*already dealt with this issue in the loops above*/
			if err := self.writeKey(s, keyName, typeName, mapKey, 0, index, order, userId); err != nil {
				return err
			}
			q:=ValueInfo{mapKey,userId,time.Time{}}
			unique=append(unique,q)
		}
		if err:=self.writeTreeBatch(typeName,keyName,unique); err!=nil {
			return err
		}
	}
	return nil
}

//writeItemData knows how to write the fields of a structure to memcache, assigning
//a new Id number if necessary.  this is a primitive for other types of write.
func (self *GobStore) writeItemData(s interface{}, id uint64, typeName string) error {
	var err error

	mightBeBadId:=(id!=uint64(0))

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
	
	if mightBeBadId {
		key := fmt.Sprintf(RECKEY, typeName, id)
		item,err:=self.impl.Get(key);
		if err!=nil {
			return err
		}
		item.SetValue(buffer.Bytes())
		return self.impl.CompareAndSwap(item)
	}
	
	//stuff the bytes into the store
	item := NewStoredItem()
	item.SetKey(key)
	item.SetValue(buffer.Bytes())
	err = self.impl.Set(item)
	if err != nil {
		return err
	}
	return nil
}

//getRecordKey returns a key string for a given set of information about the key.  This is
//used to create different keys if you supply a userId (!=0) to implement a bit of segregation
//by owner.
func (self *GobStore) getKeyNameForRecord(typeName string, keyName string, keyValue string, userId uint64) string {
	if fmt.Sprintf("%s", keyValue) == "<uint64 Value>" {
		panic("somebody passed a VALUE object to getKeyNameForRecord")
	}
	if userId == uint64(0) {
		return fmt.Sprintf("%s-key-%s-val-%s", typeName, keyName, keyValue)
	}
	return fmt.Sprintf("%s-key-%s-user-%d-val-%s", typeName, keyName, userId, keyValue)
}

//UniqueKeyValues should return all the unique keys (plus the owners) for a given key and type.
//This should filter for the owner==last parameter if the vaule is non-zero.
func (self *GobStore) UniqueKeyValues(example interface{}, keyName string, result *[]ValueInfo, ownerFilter uint64) error {
	_, typeName, err := getIdValueAndStructureName(example)
	if err!=nil {
		return err
	}
	_, tree, err:=self.loadTree(typeName,keyName)
	if err!=nil {
		return err
	}
	
	ch:=tree.IterAscend()
	
	for {
		item:= <- ch
		if item==nil {
			 break
		}
		v:=item.(ValueInfo)
		if ownerFilter!=uint64(0) && v.Owner!=ownerFilter {
			continue
		}
		if len(*result)==cap(*result) {
			break
		}
		*result = append(*result,v)
	}
	return nil
}

//loadTree brings in a tree from the store, creating it if necessary.  It returns the item it read
//so it can be used with CAS ops later.  It returns null for both values if there is an error.
func (self *GobStore) loadTree(typeName string, keyName string) (StoredItem,*llrb.Tree,error) {
	key:=fmt.Sprintf(VALUEKEY,typeName,keyName)
	var err error
	var item StoredItem
	if item, err = self.impl.Get(key); err != nil {
		if err!=ErrorNotFoundInStore{
			return nil,nil,err
		}
	}
	var root *llrb.Node
	
	if err==ErrorNotFoundInStore{
		tree:=llrb.New(lessValueInfoForValue)
		return nil, tree, nil
	} 
	buffer := bytes.NewBuffer(item.Value())
	decoder := gob.NewDecoder(buffer)
	err = decoder.Decode(&root)
	if err!=nil {
		return nil,nil, err
	}
	tree:=llrb.New(lessValueInfoForValue)
	tree.SetRoot(root)
	
	return item,tree,nil
}

//saveTree takes a tree and writes it back to storage.  It assumes the passed memache.Item
//is the one returned from loadTree().  The typeName and keyName are only used if the
//item is null (because this is a fresh create).
func (self *GobStore) saveTree(typeName string, keyName string, item StoredItem, tree *llrb.Tree) error {
	//encode
	buffer := new(bytes.Buffer)
	enc := gob.NewEncoder(buffer)
	if err := enc.Encode(tree.Root()); err != nil {
		return err
	}
	
	//write it back
	if item!=nil {
		item.SetValue(buffer.Bytes())
		return self.impl.CompareAndSwap(item)
	}
	//unconditional write because we have never seen it before
	item=NewStoredItem()
	item.SetKey(fmt.Sprintf(VALUEKEY,typeName,keyName))
	item.SetValue(buffer.Bytes())
	return self.impl.Set(item)
}


//addUniqueKey does the work to check and see if the key's value has been written before and
//if not to update it.
func (self *GobStore) addUniqueKey(typeName string, n string, v string, userId uint64, isFifo keyOrder) error {

	//fmt.Printf("addUniqueKey: %s [%s] '%s'\n",typeName,n,v)

	item, tree, err:=self.loadTree(typeName, n)
	if err!=nil {
		return err
	}
	pair:=ValueInfo{v,userId,time.Time{}}
	before:=tree.Len()
	tree.ReplaceOrInsert(pair)
	
	//nothing changed, don't bother with the rest
	if tree.Len()==before {
		return nil
	}
	
	return self.saveTree(typeName,n,item,tree)

}

//deleteUnique key updates our tree that maintains what keys we have seen before
func (self *GobStore) deleteUniqueKey(typeName string, n string, v string, userId uint64, isFifo keyOrder) error {

	//fmt.Printf("deleteUniqueKey: %s [%s] '%s'\n",typeName,n,v)


	item, tree, err:=self.loadTree(typeName, n)
	if err!=nil {
		return err
	}
	pair:=ValueInfo{v,userId,time.Time{}}

	if tree.Delete(pair)==nil {
		panic(fmt.Sprintf("deleted an item from the tree but could not find it! %+v\n",pair))
	}
	
	return self.saveTree(typeName,n,item,tree)
	
}

//writeEmptyTree is used to init the data structure if you want to read  before
//you have written anything.  
func (self *GobStore) writeEmptyTree(typeName string, n string) error {
	return nil
}
//writeTreeBatch is used when you have many writes to do all at once.
func (self *GobStore) writeTreeBatch(typeName string, keyName string, possiblyUnique []ValueInfo) error {
	item, tree, err:=self.loadTree(typeName, keyName)
	if err!=nil {
		return err
	}
	for _,u:=range possiblyUnique {
		tree.InsertNoReplace(u)
	}
	return self.saveTree(typeName,keyName,item,tree)
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
