package seven5

import (
	"fmt"
	"reflect"
	"strings"
)

//TypeHolder is a utility for holding the descriptions of a collection of wire types.  Typically,
//dispatchers use this if they want to provide a way to allow code generators to easily connect to
//their REST resources.  Typically, callers should be aware that calling Add() on a type that does
//not meet the seven5 requirements for all it's fields will cause a panic.
type TypeHolder interface {
	Add(name string, wireType interface{})
	All() []*FieldDescription
}

//SimpleTypeHolder is a basic implementation of TypeHolder.
type SimpleTypeHolder struct {
	all []*FieldDescription
}

//NewSimpleTypeHolder returns a new, initialized ptr to a SimpleTypeHolder.
func NewSimpleTypeHolder() *SimpleTypeHolder {
	return &SimpleTypeHolder{nil}
}

//All returns the full list of wire types known to this TypeHolder.
func (self *SimpleTypeHolder) All() []*FieldDescription {
	return self.all
}

//Add takes the type supplied and adds it to the list of known resources.  If this type is not
//a valid wire type it will panic.  It does not check to see if the type has been
//added previously.
func (self *SimpleTypeHolder) Add(name string, i interface{}) {
	t := reflect.TypeOf(i)
	d := WalkWireType(name, t)
	self.all = append(self.all, d)
}

//Field description gives information about a particular field and this is part of what
//is passed over the wire to describe a resource.  This describes the type that is
//being used for a particular resource.  If Array or Struct fields are not nil,
//then there is a nested structure or array inside the struct and the type name should
//be ignored.
type FieldDescription struct {
	//name is required
	Name string
	//type name must a simple type
	TypeName string
	//Ptr to allow nil values
	IsPtr bool
	//arrays are composed of some number of a single _type_ ... if there is an array
	//there should NOT be a TypeName or a Struct defn
	Array *FieldDescription
	//struct name is separate from TypeName so we can disambiguate a struct 'Floating' from a
	//base type of the same name
	StructName string
	//structs are zero or more different fields
	Struct []*FieldDescription
}

//WalkWireType is the recursive machine that creates a FieldDescription from
//a go type.  Given a type it returns a pointer to a FieldDescription struct.
//This is public because it's likely to be useful to others.
func WalkWireType(name string, t reflect.Type) *FieldDescription {
	if t.Kind() == reflect.Slice {
		nested := WalkWireType("slice", t.Elem())
		return &FieldDescription{Name: name, Array: nested}
	}
	if t.Kind() == reflect.Struct || (t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct) {
		var structType reflect.Type
		if t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct {
			structType = t.Elem()
		} else {
			structType = t
		}
		fieldCollection := []*FieldDescription{}
		for i := 0; i < structType.NumField(); i++ {
			f := structType.Field(i)
			tagValue := f.Tag.Get("seven5")
			values := strings.Split(tagValue, ",")
			ignore := false
			for _, v := range values {
				if v == "wireignore" {
					ignore = true
					break
				}
			}
			if ignore {
				continue
			}
			nested := WalkWireType(f.Name, f.Type)
			fieldCollection = append(fieldCollection, nested)
		}
		return &FieldDescription{Name: name, StructName: structType.Name(),
			Struct: fieldCollection}
	}
	k := t.Kind() //we change this if this a ptr
	isPtr := false
	if t.Kind() == reflect.Ptr {
		k = t.Elem().Kind()
		isPtr = true
	}
	switch k {
	case reflect.Bool, reflect.Float64, reflect.Int64, reflect.String:
		return &FieldDescription{Name: name, IsPtr: isPtr, TypeName: t.Name()}
	case reflect.Float32, reflect.Int, reflect.Int32, reflect.Int8:
		panic(fmt.Sprintf("Use of (%v) prohibited in wire types.  "+
			"Use float64 or int64 to avoid ambiguity when json marshalling", t.Kind()))
	case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint8:
		panic(fmt.Sprintf("Use of unsigned types like %v prohibited."+
			" Use int64 to prevent json marshalling ambiguity.", t.Kind()))
	}
	panic(fmt.Sprintf("Unable to understand type definition %s for conversion to a json format", t.Name()))
}
