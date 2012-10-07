package seven5

import (
	"fmt"
	_ "github.com/russross/blackfriday"
	"reflect"
)

//Dispatch is used by the SimpleHandler to know what resources are currently loaded. 
//The SimpleHandler maintains a list of all the interfaces that are currently in place
//and what is known about each.  This is public because it is likely any implementation 
//of Handler will need this to generated api/doc.
type Dispatch struct {
	ResType     interface{}
	GETSingular string
	GETPlural   string
	Index       Indexer
	Find        Finder
}

//Field description gives information about a particular field and this is part of what 
//is passed over the wire to describe a resource.  This describes the type that is 
//being used for a particular resource.  If Array or Struct fields are not nil, 
//then there is a nested structure or array inside the struct and the type name should
//be ignored.
type FieldDescription struct {
	Name string
	GoType, DartType, SqlType string
	Array *FieldDescription
	Struct *[]FieldDescription
}

//ResourceDescription is the full type passed over the wire to describe how a particular 
//can be called and what fields the objects have that it manipulates.
type ResourceDescription struct {
	Name string
	Index bool
	Find bool
	GETSingular string
	GETPlural string
	CollectionDoc []string
	ResourceDoc[] string
	Fields *[]FieldDescription
}

//NewDispatch is called to create a new Dispatch instance from a given type.  The passed
//value should be a struct (not a pointer to a struct) that describes the json exchanged 
//between the client and server.  If the passed type isn't a struct, this function panics.
//This code does not fill anything in the result other than the resource type.
func NewDispatch(r interface{}) *Dispatch {
	if r == nil || reflect.TypeOf(r).Kind() != reflect.Struct {
		panic(fmt.Sprintf("Can't create a resource from an object that is not a struct (%T)!", r))
	}
	return &Dispatch{ResType: r}
}



//walkJsonType recursively walks the type structure provided.  This is called recursively 
//if there is a nested struct or array. The top level type must be a struct, not a simple
//type or array.  walkJsonType will panic if it finds something that cannot be correctly 
//flattened to json, converted to Dart, and converted to SQL.
func walkJsonType(t reflect.Type) *[]FieldDescription {
	var result []FieldDescription
	
	if t.Kind() != reflect.Struct &&t.Kind() != reflect.Array {
		panic(fmt.Sprintf("can't understand type %v as a json, top-level object", t))
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		name:= field.Name
		k := field.Type.Kind()
		switch (k) {
		case reflect.Bool, reflect.String, reflect.Float64, reflect.Int64:
			g, d, s := toMultipleTypeNames(field.Type)
			result = append(result, FieldDescription{Name: name, GoType: g, DartType:d, SqlType:s})
		case reflect.Struct:
			nested:=walkJsonType(t.Field(i).Type)
			result = append(result, FieldDescription{Name: name, Struct: nested})
		case reflect.Slice:
			elem:=t.Field(i).Type.Elem()
			elemDesc:=walkJsonType(elem)
			nested:=FieldDescription{Name: elem.Name(), Struct:elemDesc}
			result = append(result, FieldDescription{Name: name, Array: &nested})
		default:
			panic(fmt.Sprintf("can't understand type field %s of type %s as part of a json struct", 
				name, k.String()))
		}
	}
	return &result
}

//given a go type, compute the human readable type name for various systems
func toMultipleTypeNames(t reflect.Type) (string, string, string){
	if t.PkgPath()!="seven5" {
		panic(fmt.Sprintf("Should not be trying to serialize a structure that is not "+
			" completely composed of seven5 types.  Type %s.%s is not from seven5.",
			t.PkgPath(),t.Name()))
	}
	switch t.Name() {
	case "Boolean":
		return "bool", "bool", "int"
	case "String255":
		return "string", "String", "varchar(255)"
	case "Integer", "Id":
		return "int64", "int", "integer"
	case "Floating":
		return "float64", "double", "real"
	}
	panic(fmt.Sprintf("don't know how to convert %s.%s to a dart and sql type!",
		t.PkgPath(),t.Name()))
}
