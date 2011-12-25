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
)

type MemcacheGobStore struct {
	*memcache.Client
}

const (
	LOCALHOST = "localhost:11211"
	IDKEY     = "%s-idcounter"
	RECKEY    = "%s-%d"
	EXTRAKEY    = "%s-%s-%s"
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
	
	if value, typeName, err= VerifyStructPointerFields(s); err!=nil {
		return err
	}
	if value == uint64(0) {
		newId, err := self.GetNextId(typeName)
		if err != nil {
			return err
		}
		id:= reflect.ValueOf(s).Elem().FieldByName("Id")
		id.SetUint(newId)
		value = newId
	}
	//at this point value holds the number of the record
	key := fmt.Sprintf(RECKEY, typeName, value)
	buffer := new(bytes.Buffer)
	enc:=gob.NewEncoder(buffer)
	if err:=enc.Encode(s); err!=nil {
		return err
	}
	item:=&memcache.Item{Key:key,Value:buffer.Bytes()}
	err=self.Client.Set(item)
	if err!=nil {
		return err
	}
	for _,k:=range(GetStructKeys(s)) {
		err=self.writeKey(s,k,typeName,value)
		if err!=nil {
			return err
		}
	}
	return nil
}

//FindById is the reverse of Write and reads a structure from memcached for a given type and Id.
//The first parameter should be a point to a zero-valued struct.
func (self *MemcacheGobStore) FindById(s interface{}, id uint64) error {
	var value uint64
	var typeName string
	var err error
	
	if value, typeName, err= VerifyStructPointerFields(s); err!=nil {
		return err
	}
	if value!=uint64(0) && value!=id {
		panic("disagreement on id of the value to read!")
	}
	key:=fmt.Sprintf(RECKEY, typeName, id)
	var item *memcache.Item
	if item,err=self.Client.Get(key); err!=nil {
		return err
	}
	buffer:=bytes.NewBuffer(item.Value)
	decoder:=gob.NewDecoder(buffer)
	return decoder.Decode(s)
}


//writeKey assumes that the pointer to struct has already been checked and is ok.
//writeKey writes the keyName field of the struct pointed to by s into the
//memcache.  It puts under that key the id needed to find the real value of the object.
func (self *MemcacheGobStore) writeKey(s interface{},keyName string, typeName string, id uint64) error {
	str:=reflect.ValueOf(s).Elem()
	f:=str.FieldByName(keyName)
	value:=fmt.Sprintf("%s",f)
	key:=fmt.Sprintf(EXTRAKEY,typeName,keyName,value)
	item:=&memcache.Item{Key:key, Value:[]byte(fmt.Sprintf("%d",id))}
	return self.Client.Set(item)
}

//FindByKey looks up a value in the memcache by a field _other_ than the Id field.  You have
//to supply the name of the field.  Further, that field must exist in the structure, be
//exported (uppercase), and must flatten (via String()) to a value that is acceptable to be
//used as a key in memcache.  This code (and its opposite that writes values by a non id key) 
//is careful to convert the value of the new key to bytes via "Sprintf" so only an implementation
//of String() is needed.  This call results in two roundtrips to the memcache server because
//it first retrieves the id of the object that is stored under the key value provided and
//then calls FindById.
func (self *MemcacheGobStore) FindByKey(s interface{}, keyName string, value string) error {
	var typeName string
	var err error
	
	if _, typeName, err= VerifyStructPointerFields(s); err!=nil {
		return err
	}

	var item *memcache.Item

	key:=fmt.Sprintf(EXTRAKEY,typeName,keyName,value)
	if item,err=self.Client.Get(key); err!=nil {
		if err==memcache.ErrCacheMiss {
			return NO_SUCH_KEY
		}
		return err
	}
	id,err:=strconv.ParseInt(string(item.Value),  10, 64)
	if err!=nil {
		return err
	}
	return self.FindById(s,uint64(id))
}