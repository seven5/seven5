//store is a subpackage of seven5 for storing and retrieving data against a key-value store
//The interface store.T is the abstraction that (with some luck) will be sufficient to 
//support multiple key-value impls.  Different storage implementations may place 
//different restrictions on what values can be stored, how they must be related, etc.  
//Details of the expectations of a store implementation are visible on the type
//StoreImpl.
//
//The Write() and Find...() interfaces take a pointer to a structure and this should 
//not be changed by the store implementation--although the store implementation may 
//place restrictions on the structure itself.  The annotation 
//seven5key:"<keyname>[,<keyname>...]" should 
//be used on struct fields that should be considered keys.  For a simple key, set the
//keyname value to the name of the field. Any other value must be a method on the
//struct (not the pointer to it!) that will be called to compute a value for the field.
//This allows construction of some types of aggregrates and some simple joins.
//
//Any type of value can be used an "extra" key via the annotation mechanism but it is the
//implementors responsibility to make sure that the value correctly flattens to a "clean" key
//when printed using the String() method and that this value is not excessively large.  
//Some stores have key length limits so lengths should be probably be less than 100
//characters.  It is normally necessary to take steps to
//remove special characters or convert to a base64 representation.
//
//Fields marked as keys should be careful to make sure that the values are well "spread" or
//reading and writing the index can become a serious bottleneck.  An example is that
//the FindAll() method must maintain a complete list of all the items of a particular type
//and this can be slow.  You can use the struct tag seven5order:"none" turn this off (only
//for the Id field) and the FindAll method will always return an error.
//
package store

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

const (
	//order inserted is order returned (expensive)
	fifo_order = iota  
	//order inserted is opposite returned (expensive)
	lifo_order = fifo_order+1  
	//order returned is unrelated to order inserted (cheap)
	unspecified_order = fifo_order+2  
)

type T interface {
	//Write should take a pointer to a structure and store it.  The pointer may need to have
	//its Id field filled in and this is indicated by a 0 value in the field.  Multiple calls
	//to Write must be idempotent.
	Write(interface{}) error
	//FindById fills in a pointer to a structure passed as an argument based on the id passed
	//as a parameter.  FindById should not allocate the storage for the pointer.
	FindById(interface{}, uint64) error
	//FindByKey fills in the slice values in result based on the keyName and keyValue provided.
	//The caller must initialize the slice and the number of values returned will be the number
	//empty slots in the slice.  Because slices are passed by value, you must pass the address
	//of the created slice.  If you pass a slice with no empty slots, no results are returned.
	//FindByKey allocates the storage needed for all the new items it returns. If you pass 
	//a userId parameter, it will be used to retrieve only values who have the Owner field
	//in the structure set to that value (use zero if no owner field).  If you have an
	//Owner field in your struct, this does a "cross owner" search which is both more
	//expensive and cannot preserve the order of insertion of the keys.
	FindByKey(result interface{}, keyName string, keyValue string, userId uint64) error
	//DeleteById deletes an item from store so it cannot be found again with either FindById
	//or FindByKey.  The first parameter is not touched, but must be a pointer to a structure
	//of the appropriate type being deleted.  The Id and (if present) Owner field of the structure
	//should be set to the values needed for deletion.
	Delete(example interface{}) error
	//Init takes an example pointer and creates the necessary entries in the store to prepare
	//for storing this type.  This is not necessary if you do a Write() before your first
	//Find..., but is needed if you want FindByKey() to work properly with an empty set of objects.
	Init(example interface{}) error
	//FindAll fills in the slice pointed to by the first item with all the known items of 
	//the particular type denoted by the slice.  Note that the slice will be filled with only
	//as many items as there is capacity to hold in the slice.  Because it can be 
	//expensive to keep an index of all keys, this can be turned off with the structure
	//tag seven5order:"none" on the Id field of the structure.  If you pass a userId
	//only structures that have that value in the Owner field will be used. For structures
	//without such a field, pass 0.  If on such structures have been stored, this returns
	//a zero length result (nil)
	FindAll(result interface{}, userId uint64) error
	//BulkWrite does the same logical operation as write but does not write any intermediate 
	//states so it is much more appropriate for loading a datastore with data. The first
	//parameter should be a slice, where each entry is a pointer to the appropriate structure
	//type.  Note that if you want to assign ids to each of these items as you go, you must
	//insure that their Id field is uint64(0) to force the creation of an id.  If the 
	//items to be written have an Owner field, it must be set to a valid value and it must
	//be the same for all values written.
	BulkWrite(sliceOfPtrs interface{}) error
	//UniqueKeyValues returns the set of unique keys that have been written to the store for
	//a given type and key.  The returned result is pairs which are unique keys and which 
	//user "owns" that value.  Note that the store may not allow access to some of the values
	//if they are owned by other users.  The first parameter must be a pointer to a valid
	//structure with appropriate seven5key annotations, but it is unchanged.  The caller
	//should pass a ptr to slice with empty spaces in it (cap!=len) and these will be filled until
	//no more spots remain or all the values have been elucidated.  This function will
	//allocate new objects to hold the newly returned values.  If the final parameter is
	//not zero, then only values owned by that id will be returned in the result.
	UniqueKeyValues(example interface{}, keyName string, result *[]ValueInfo, ownerFilter uint64) error
}

//ValueInfo is a structure that is used when returning the set of unique keys. Note that
//values have to flatten nicely, so the value portion here is always a string.
type ValueInfo struct {
	Value string
	Owner uint64
	Stored time.Time
}

//lessValueInfoValue is needed because we keep a LLRB tree underneath and need a comparison function
//for that to work.  This compares the values so we can keep a tree that allows fast decision
//to see if a given value has been used before.
func lessValueInfoForValue(a,b interface{}) bool {
	x := a.(ValueInfo)
	y := b.(ValueInfo)
	
	if x.Value==y.Value {
		return x.Stored.Before(y.Stored)
	}
	return x.Value < y.Value
}

//Lesser is a shim to allow us to do sorting with exactly knowing the types.  We delegate the
//comparison of order to the true storage class.  This is used automatically to sort results
//of FindByKey() and FindAll() if it is present.  Note that a result set is sorted *after*
//the results are found, so if you only alocate 10 places in the result slice but there are
//100 items available you might not get what you expect.  Some control over this can be
//had with the use of the seven5order annotation.
type Lesser interface {
	Less(reflect.Value, reflect.Value) bool
}

var (
	BAD_STRUCT       = errors.New("stored values are always read and written as pointers to structs")
	BAD_SLICE        = errors.New("slices should have element type that is pointer to structs")
	BAD_SLICE_PTR    = errors.New("slices are passed by value, so you must pass a pointer to a slice so it can be 'filled in' by seven")
	BAD_ID           = errors.New("stored structs must have a field named 'Id' that is type uint64")
	NO_STRING_METHOD = errors.New("any value used in a key field must have a String() method")
	INDEX_MISS       = errors.New("can't find that item in the index")
	ZeroValue        = reflect.Value{}
)

//these are the key name and key value for the "all" index, if used
const (
	MAGIC_KEY   = "--"
	MAGIC_VALUE = "--"
)

//getIdValueAndStructureName makes sure that the object passed is a pointer to a structure and
//that it has a field named Id that is of type uint64.  If the check fails, it return non-nil
//error value, otherwise the current value of the Id field and the name of the structure type
//is return (not the name of the type passed as a parameter, as this is a POINTER to a structure).
func getIdValueAndStructureName(s interface{}) (uint64, string, error) {
	return getIdValueAndStructureNameFromValue(reflect.ValueOf(s))
}

//getIdValueAndStructureNameFromValue is identical to getIdValueAndStructureName except it operates
//on a reflect.Value not an interface{}.
func getIdValueAndStructureNameFromValue(v reflect.Value) (uint64, string, error) {
	if err := verifyStructPointerFieldTypes(v.Type()); err != nil {
		return uint64(0), "", err
	}
	str := v.Elem()
	typeName := str.Type().String()
	id := str.FieldByName("Id")
	return id.Uint(), typeName, nil
}

//verifyStructPointerFieldType is used to check that the object described by v is a pointer to
//structure that has an appropriate Id field.
func verifyStructPointerFieldTypes(v reflect.Type) error {
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

type keyOrder int

//MethodPlusName is used to allow us to return the string of the methods name along with an
//object that points to that method plus the value of the receiver.  Without this pair,
//the method name cannot be determined from the reflect.Value.
type methodPlusName struct {
	Name   string
	Meth   reflect.Value
	IsFifo keyOrder
}

//FieldPlusName is used to allow us to return name of a field
//plus a value object that is the value of the field named.
type fieldPlusName struct {
	Name   string
	Value  reflect.Value
	IsFifo keyOrder
}

//GetStructKeys returns two slices that are the names of the fields that are keys
// (first slice) or methods that generate keys for this structure type in memcached.  
//It assumes that one has already validated the struct with verifyStructPointerFields.
func getStructKeys(s interface{}) ([]fieldPlusName, []methodPlusName) {
	str := reflect.ValueOf(s).Elem()
	t:=str.Type()
	resultFields := []fieldPlusName{}
	resultMethods := []methodPlusName{}
	numFields := t.NumField()

	for i := 0; i < numFields; i++ {
		isFifo := unspecified_order
		none := false
		f := t.Field(i)

		tag := f.Tag.Get("seven5order")
		if tag != "" {
			//lifo and none are possible other cases
			if tag == "lifo" {
				isFifo = lifo_order
			} else if tag == "none" {
				if f.Name != "Id" {
					panic(fmt.Sprintf("setting seven5order to none makes no sense on key %s", f.Name))
				}
				none = true
			} else if tag == "fifo" {
				isFifo=fifo_order
			} else {
				panic(fmt.Sprintf("invalid value of seven5order on key %s [%s]", f.Name, tag))
			}
		}
		//Id is a special case, because you can turn off the findAll()
		if f.Name == "Id" {
			if !none {
				resultFields = append(resultFields, fieldPlusName{MAGIC_KEY, ZeroValue, keyOrder(isFifo)})
			}
			continue
		}
		//check for the other keys on other fields, possibly comma separated
		if f.Tag.Get("seven5key") == "" {
			if isFifo!=unspecified_order {
				panic(fmt.Sprintf("You specified an order, but no key(s) with that order (%s)",f.Name))
			}
			continue
		}
		tagList := strings.Split(f.Tag.Get("seven5key"), ",")
		for _, tag := range tagList {
			if tag == f.Name {
				resultFields = append(resultFields, fieldPlusName{f.Name, str.Field(i), keyOrder(isFifo)})
				continue
			}
			m:=str.MethodByName(tag)
			if m==ZeroValue{
				panic(fmt.Sprintf("method %s does not exist on struct! (did you define it on the POINTER to the struct?)", tag))
			}
			resultMethods = append(resultMethods, methodPlusName{tag, m, keyOrder(isFifo)})
		}
	}

	return resultFields, resultMethods
}

