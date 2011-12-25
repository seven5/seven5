//store is a subpackage of seven5 for storing and retrieving data.  The interface store.T
//is the abstraction that (with some luck) will be sufficient to support multiple underlying
//storage strategies.  Different storage implementations may place different restrictions
//on what values can be stored, how they must be related, etc.  The Write() and
//Find...() interfaces take a pointer to a structure and this should not be changed by
//the store implementation--although the store implementation may place restrictions on
//the structure itself.  The annotation seven5key:"true" should be used on struct fields
//that should be considered keys; this is not necessary on the Id field as it is always
//a key.
//Any type of value can be used an "extra" key via the annotation mechanism but it is the
//implementors responsibility to make sure that the value correctly flattens to a "clean" key
//when printed using the String() method.  It is normally necessary to take steps to
//remove special characters or convert to a base64 representation.
package store

import (
	"errors"
	"reflect"
)

type T interface {
	//DestroyAll should completely clear the store.  It is used primarily by tests.
	DestroyAll(params...string) error 
	//Write should take a pointer to a structure and store it.  The pointer may need to have
	//its Id field filled in and this is indicated by a 0 value in the field.  Multiple calls
	//to Write must be idempotent.
	Write(interface{}) error
	//FindById fills in a pointer to a structure passed as an argument based on the id passed
	//as a parameter.  It returns NO_SUCH_KEY if the id could not be located.
	FindById(interface{}, uint64) error
	//FindByKey fills in a pointer to a structure based on a keyName and a value. If the key
	//cannot be found with the correct value, NO_SUCH_KEY is returned.
	FindByKey(someStruct interface{}, keyName string, keyValue string) error}

var (
	BAD_STRUCT = errors.New("stored values are always read and written as pointers to structs")
	BAD_ID = errors.New("stored structs must have a field named 'Id' that is type uint64")
	NO_SUCH_KEY = errors.New("key not found")
)


//VerifyStructPointerFields makes sure that the object passed is a pointer to a structure and
//that it has a field named Id that is of type uint64.  If the check fails, it return non-nil
//error value, otherwise the current value of the Id field and the name of the structure type
//is return (not the name of the type passed as a parameter, as this is a POINTER to a structure)
func VerifyStructPointerFields(s interface{}) (uint64,string,error) {
	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Ptr {
		return uint64(0),"",BAD_STRUCT
	}
	str := v.Elem()
	if str.Kind() != reflect.Struct {
		return uint64(0),"",BAD_STRUCT
	}
	typeName := str.Type().String()
	id := str.FieldByName("Id")
	if (id == reflect.Value{}) {
		return uint64(0),"",BAD_ID
	}
	if id.Kind() != reflect.Uint64 {
		return uint64(0),"",BAD_ID
	}
	return id.Uint(),typeName,nil
}

//GetStructKeys returns a slice of strings that are the names of the fields that are also to
//be considered keys for this structure type in memcached.  It assumes that one has already
//validated the struct with verifyStructPointerFields.
func GetStructKeys(s interface{}) []string {
	str:=reflect.ValueOf(s).Elem()
	result:=[]string{}
	
	numFields:=str.NumField()
	for i:=0; i<numFields; i++ {
		f:=str.Type().Field(i)
		if f.Name=="Id" {
			continue
		}
		tag:=f.Tag.Get("seven5key")
		if tag=="" {
			continue
		}
		if tag!="true" {
			panic("cannot understand seven5key tag value!")
		}
		result=append(result,f.Name)
	}
	return result
}