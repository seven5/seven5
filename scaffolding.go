package seven5

import (
	"errors"
	"fmt"
	"go/ast"
	"reflect"
	"seven5/store"
	"strings"
)

var internalErrorBadType = errors.New("Unexpected type encountered in scaffolding implementation of Restful service!")
var internalErrorNoField = errors.New("Unexpected type encountered ... expected to find field but did not!")

//ScaffoldRestService gives you a simple implementation of a rest service that will allow you
//to begin work on a client side application immediately.  You must supply a slice of your 
//type (usually size 0) as the first parameter, typicially with make([]*MyType,0,0) or a
//literal equivalent.  The second and third parameters represent some basic security checks.
//If the second parameter is true, any user can read from the store even without being logged in;
//if it false you must be logged in to read values.  The third parameter is similar, except if
//false any logged in user can modify values in the store (including delete, create, etc) and
//if true only a superuser may do those operations. These are not intended to provide perfect
//control but just enough that work on the client side can progress with 3 possible cases,
//not logged in, logged in, and logged in as superuser.
func ScaffoldRestService(singularNounExampleSlice interface{}, allowAllToRead bool, allowOnlySuperUserToChange bool) *RestHandlerDefault {
	sliceType := reflect.TypeOf(singularNounExampleSlice)
	if sliceType.Kind() != reflect.Slice {
		panic(fmt.Sprintf("Scaffold Rest Service expects slice (of size 0) with element type of a pointer to an exported structure! You supplied %v\n", sliceType))
	}
	exampleType := sliceType.Elem()
	if exampleType.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("Scaffold Rest Service expects a pointer to an exported structure! You supplied %v\n", exampleType))
	}

	structType := exampleType.Elem()
	if structType.Kind() != reflect.Struct {
		panic(fmt.Sprintf("Scaffold Rest Service expects a pointer to an exported structure! You supplied a pointer to %v\n", structType))
	}

	singular := strings.ToLower(structType.String())
	if idx := strings.Index(singular, "."); idx != -1 && idx != len(singular) {
		singular = singular[idx+1:]
	}

	if !ast.IsExported(structType.Name()) {
		panic(fmt.Sprintf("Scaffold Rest Service expects a pointer to an *exported* structure! Your structure is not exported (doesn't start with uppercase letter)\n", structType))
	}

	_, hasOwner := structType.FieldByName("Owner")
	svc := &scaffoldedRestService{sliceType, exampleType, structType, hasOwner, allowAllToRead, allowOnlySuperUserToChange}
	return NewRestHandlerDefault(svc, singular)

}

//scaffoldedRestService is the implementation class used to implement the scaffolded services. It
//implements enough of Restful to be passed as a parawer to NewRestHandlerDefault.
type scaffoldedRestService struct {
	sliceType                  reflect.Type
	exampleType                reflect.Type
	structType                 reflect.Type
	hasOwner                   bool
	allowAllToRead             bool
	allowOnlySuperuserToChange bool
}

//getId returns the id field of the ptr to struct.
func (self *scaffoldedRestService) getId(ptrToValues interface{}) uint64 {
	var zero reflect.Value
	value := reflect.ValueOf(ptrToValues).Elem()
	f := value.FieldByName("Id")
	if f == zero {
		panic(internalErrorNoField)
	}
	return f.Uint()
}

//setUint64 sets the given fieldName in the poiner to a structure to a new value. Useful
//for setting the Id field.
func (self *scaffoldedRestService) setUint64Field(fieldName string, ptrToValues interface{}, newValue uint64) error {
	var zero reflect.Value

	value := reflect.ValueOf(ptrToValues).Elem()
	//fmt.Printf("valueof %v --- field wanted %s\n",value,fieldName)
	f := value.FieldByName(fieldName)
	if f == zero {
		panic(internalErrorNoField)
	}
	if !f.CanSet() {
		if fieldName == "Id" {
			fmt.Printf("----maybe you forgot to include an Id field in your structure??--- ")
		}
		panic("cant set value of f inside set field!")
	}
	f.SetUint(newValue)
	return nil
}

//setOwner sets the Owner field of a structure.  It checks first to see if the struct has
//an owner and if not, it does nothing.
func (self *scaffoldedRestService) setOwner(ptrToValues interface{}, session *Session) error {
	if !self.hasOwner || session == nil {
		return nil
	}
	return self.setUint64Field("Owner", ptrToValues, session.User.Id)
}

//checkValueType checks that the (alleged) ptrToStruct that is provided is the type we
//were expecting to see for this scaffolded service.  Mismatches are very bad here
//so we want to know about this internal error anytime it occurs.
func (self *scaffoldedRestService) checkValueType(ptrToValues interface{}) error {
	if reflect.TypeOf(ptrToValues) != self.exampleType {
		return internalErrorBadType
	}
	return nil
}

//Create in the scaffolded version assumes that any object supplied is ok to write.  It has
//no notion of required fields.    It returns the object created to the client.
func (self *scaffoldedRestService) Create(store store.T, ptrToValues interface{}, session *Session) error {
	if err := self.checkValueType(ptrToValues); err != nil {
		return err
	}
	//prevent errors and possible dodgy decisions by the user
	if err := self.setOwner(ptrToValues, session); err != nil {
		return err
	}
	//just to be sure
	if err := self.setUint64Field("Id", ptrToValues, 0); err != nil {
		return err
	}

	//we don't need to see if we need to read it back because we know that there is nobody
	//else adding fields and such
	err := store.Write(ptrToValues)
	return err

}

//Read an object from the store.  Return the value read to the client.
func (self *scaffoldedRestService) Read(store store.T, ptrToObject interface{}, session *Session) error {
	if err := self.checkValueType(ptrToObject); err != nil {
		return err
	}
	return store.FindById(ptrToObject, self.getId(ptrToObject))
}

//Update an object from the story.  This update will refuse to modify (ignore) attempts to
//change the Id or Owner fields.  The update policy is that the values provided from the
//client must be different from the existing value for this item (Id) in the store AND 
//must not be the zero value for the appropriate type.  This is not ideal as it means that
//you can reset an integer to zero or a string to "", for example, via the update method. 
//However, normally a zero value means that the client did not supply the value in the
//payload.  It returns the new state of the object.
func (self *scaffoldedRestService) Update(store store.T, ptrToNewValues interface{}, session *Session) error {
	//better be the same type
	if err := self.checkValueType(ptrToNewValues); err != nil {
		return err
	}

	//make a fresh new one based on id supplied
	id := self.getId(ptrToNewValues)
	other := self.Make(id)

	//read in current value
	if err := store.FindById(other, id); err != nil {
		return err
	}

	//need structures, not pointers to them
	otherStruct := reflect.ValueOf(other).Elem()
	newValuesStruct := reflect.ValueOf(ptrToNewValues).Elem()
	typeOther := otherStruct.Type()

	//walk all fields
	for i := 0; i < otherStruct.NumField(); i++ {
		otherField := otherStruct.Field(i)
		newValueField := newValuesStruct.Field(i)

		//get the type of the field and check to see if it's a special one
		fieldType := typeOther.FieldByIndex([]int{i})
		fieldName := fieldType.Name
		if fieldName == "Id" || fieldName == "Owner" {
			continue
		}
		//see if the newly supplied value is a zero value (meaning didn't come from client)
		isZero := false
		zero := reflect.Zero(fieldType.Type)
		if reflect.DeepEqual(zero.Interface(), newValueField.Interface()) {
			isZero = true
		}
		if !reflect.DeepEqual(otherField.Interface(), newValueField.Interface()) && !isZero {
			//choose new values wherever there is a difference and we're not setting to zero
			otherField.Set(newValueField)
		}
	}
	if err := store.Write(other); err != nil {
		return err
	}
	//reset the values they gave us
	elem := reflect.ValueOf(ptrToNewValues).Elem()
	valOther := reflect.ValueOf(other).Elem()
	elem.Set(valOther)
	return nil
}

//Delete removes an item from the store and returns the content of the object deleted to the
//client.
func (self *scaffoldedRestService) Delete(store store.T, ptrToValues interface{}, session *Session) error {
	if err := self.checkValueType(ptrToValues); err != nil {
		return err
	}
	id := self.getId(ptrToValues)
	other := self.Make(id)
	if err := store.FindById(other, id); err != nil {
		return err
	}
	//MUST BE OTHER! Cannot be ptrToValues (yet) because will fail with empty fields
	if err := store.Delete(other); err != nil {
		return err
	}
	elem := reflect.ValueOf(ptrToValues).Elem()
	valOther := reflect.ValueOf(other).Elem()
	elem.Set(valOther)
	return nil
}

//FindByKey searches he store for values equal to value for the key provided.  This code, for 
//now, does no checking that the key is actually a key field of the object nor does it check
//that the key and value are ok to be used with the store and don't contain special characters.
//It returns an array of items found with the key provided, or an array of 0 elements.
func (self *scaffoldedRestService) FindByKey(store store.T, key string, value string, session *Session, max uint16) (interface{}, error) {
	slice := reflect.MakeSlice(self.sliceType, 0, int(max))
	ptr := reflect.New(self.sliceType)
	ptr.Elem().Set(slice)
	owner := uint64(0)
	if self.hasOwner && session != nil {
		owner = session.User.Id
	}
	err := store.FindByKey(ptr.Interface(), key, value, owner)
	if err != nil {
		return nil, err
	}
	return ptr.Interface(), nil
}

//enforceWrite implements the write policy chosen with the 3rd parameter to ScaffoldRestService
func (self *scaffoldedRestService) enforceWrite(session *Session) (map[string]string, bool) {
	errMap := make(map[string]string)

	if session == nil {
		errMap["_"] = "insufficient privileges for operation (must be logged in)"
		return errMap, false
	}
	if !bool(session.User.IsSuperuser) && self.allowOnlySuperuserToChange {
		errMap["_"] = "insufficient privileges for operation (must be superuser)"
		return errMap, false
	}
	return nil, true
}

//enforceRead implements the read policy chosen with the 2nd parameter to ScaffoldRestService
func (self *scaffoldedRestService) enforceRead(session *Session) (map[string]string, bool) {
	errMap := make(map[string]string)

	if session == nil && !self.allowAllToRead {
		errMap["_"] = "insufficient privileges for operation (must be logged in)"
		return errMap, false
	}
	return nil, true
}

//Validate is called by the infrastructure.  For scaffolded services, the only checks that are
//made are the ones indicated by the last two parameters to ScaffoldRestService.
func (self *scaffoldedRestService) Validate(store store.T, ptrToValues interface{}, op RestfulOp, key string, value string, session *Session) map[string]string {

	if err := self.checkValueType(ptrToValues); err != nil {
		m := make(map[string]string)
		m["_"] = fmt.Sprintf("%v", err)
		return m
	}

	//fmt.Printf("Validate on scaffold: %v\n",op)

	switch op {
	case OP_CREATE:
		if errMap, ok := self.enforceWrite(session); !ok {
			return errMap
		}
	case OP_DELETE:
		if errMap, ok := self.enforceWrite(session); !ok {
			return errMap
		}
	case OP_UPDATE:
		if errMap, ok := self.enforceWrite(session); !ok {
			return errMap
		}
	case OP_READ:
		if errMap, ok := self.enforceRead(session); !ok {
			return errMap
		}
	case OP_SEARCH:
		if errMap, ok := self.enforceRead(session); !ok {
			return errMap
		}
	}
	return nil
}

//Make creates a new instance of the type specified in the first parameter to ScaffoldRestService.
func (self *scaffoldedRestService) Make(id uint64) interface{} {
	ptrToType := reflect.New(self.structType)
	self.setUint64Field("Id", ptrToType.Interface(), id)
	return ptrToType.Interface()
}
