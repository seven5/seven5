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
	"fmt"
	"reflect"
)

type T interface {
	//Write should take a pointer to a structure and store it.  The pointer may need to have
	//its Id field filled in and this is indicated by a 0 value in the field.  Multiple calls
	//to Write must be idempotent.
	Write(interface{}) error
	//FindById fills in a pointer to a structure passed as an argument based on the id passed
	//as a parameter.  
	FindById(interface{}, uint64) error
	//FindByKey fills in the slice values in result based on the keyName and keyValue provided.
	//The caller must initialize the slice and the number of values returned will be the number
	//empty slots in the slice.  Because slices are passed by value, you must pass the address
	//of the created slice.  If you pass a slice with no empty slots, no results are returned.
	FindByKey(result interface{}, keyName string, keyValue string) error
	//DeleteById deletes an item from store so it cannot be found again with either FindById
	//or FindByKey.  The first parameter is not touched, but must be a pointer to a structure
	//of the appropriate type being deleted.  The Id value of the first parameter is ignored.
	DeleteById(example interface{}, id uint64) error
}

var (
	BAD_STRUCT       = errors.New("stored values are always read and written as pointers to structs")
	BAD_SLICE        = errors.New("slices should have element type that is pointer to structs")
	BAD_SLICE_PTR    = errors.New("slices are passed by value, so you must pass a pointer to a slice so it can be 'filled in' by seven")
	BAD_ID           = errors.New("stored structs must have a field named 'Id' that is type uint64")
	NO_STRING_METHOD = errors.New("any value used in a key field must have a String() method")
	INDEX_MISS = errors.New("can't find that item in the index")
)

var ZeroValue = reflect.Value{}

//VerifyStructPointerFields makes sure that the object passed is a pointer to a structure and
//that it has a field named Id that is of type uint64.  If the check fails, it return non-nil
//error value, otherwise the current value of the Id field and the name of the structure type
//is return (not the name of the type passed as a parameter, as this is a POINTER to a structure)
func GetIdValueAndStructureName(s interface{}) (uint64, string, error) {
	if err := VerifyStructPointerFieldTypes(reflect.TypeOf(s)); err != nil {
		return uint64(0), "", err
	}
	v := reflect.ValueOf(s)
	str := v.Elem()
	typeName := str.Type().String()
	id := str.FieldByName("Id")
	return id.Uint(), typeName, nil
}

//VerifyStructPointerFieldType is used to check that the object described by v is a pointer to
//structure that has an appropriate Id field.
func VerifyStructPointerFieldTypes(v reflect.Type) error {
	if v.Kind() != reflect.Ptr {
		return BAD_STRUCT
	}
	str := v.Elem()
	if str.Kind() != reflect.Struct {
		return BAD_STRUCT
	}
	id, ok := str.FieldByName("Id")
	if !ok {
		return BAD_ID
	}
	if id.Type.Kind() != reflect.Uint64 {
		return BAD_ID
	}
	return nil
}

//MethodPlusName is used to allow us to return the string of the methods name along with an
//object that points to that method plus the value of the receiver.  Without this pair,
//the method name cannot be determined from the reflect.Value.
type MethodPlusName struct {
	Name string
	Meth reflect.Value
}
//FieldPlusName is used to allow us to return name of a field
//plus a value that is its value.
type FieldPlusName struct {
	Name string
	Value reflect.Value
}

//GetStructKeys returns two slices that are the names of the fields that are keys
// (first slice) or methods that generate keys for this structure type in memcached.  
//It assumes that one has already validated the struct with verifyStructPointerFields.
func GetStructKeys(s interface{}) ([]FieldPlusName, []MethodPlusName) {
	str := reflect.ValueOf(s).Elem()
	resultFields := []FieldPlusName{}
	resultMethods := []MethodPlusName{}

	numFields := str.NumField()
	for i := 0; i < numFields; i++ {
		f := str.Type().Field(i)
		if f.Name == "Id" {
			continue
		}
		tag := f.Tag.Get("seven5key")
		if tag == "" {
			continue
		}
		if tag == f.Name {
			resultFields = append(resultFields, FieldPlusName{tag,str.Field(i)})
			continue
		}
		m := str.MethodByName(tag)
		if m == ZeroValue {
			panic(fmt.Sprintf("method %s does not exist on struct! (did you define it on the POINTER to the struct?)", tag))
		}
		resultMethods = append(resultMethods, MethodPlusName{tag,m})
	}
	return resultFields, resultMethods
}
