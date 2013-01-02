package seven5

import (
	"fmt"
	//	_ "github.com/russross/blackfriday"
	"reflect"
	"strings"
)

//Dispatch is used by the SimpleHandler to know what resources are currently loaded. 
//The SimpleHandler maintains a list of all the interfaces that are currently in place
//and what is known about each.  This is public because it is likely any implementation 
//of Handler will need this to generate api/doc.
type Dispatch struct {
	ResType		interface{}
	Field		*FieldDescription
	ResourceName	string
	Index		Indexer
	Find		Finder
	Post		Poster
	Put		Puter
	Delete		Deleter
}

//Field description gives information about a particular field and this is part of what 
//is passed over the wire to describe a resource.  This describes the type that is 
//being used for a particular resource.  If Array or Struct fields are not nil, 
//then there is a nested structure or array inside the struct and the type name should
//be ignored.
type FieldDescription struct {
	//name is required
	Name	string
	//type name must a simple type (from seven5) 
	TypeName	string
	//arrays are composed of some number of a single _type_ ... if there is an array
	//there should NOT be a TypeName or a Struct defn
	Array	*FieldDescription
	//struct name is separate from TypeName so we can disambiguate a struct 'Floating' from a
	//base type of the same name
	StructName	string
	//structs are zero or more different fields
	Struct	[]*FieldDescription
}

//APIDoc is the full type passed over the wire to describe how a particular resource
//can be called and what fields the objects have that it manipulates.
type APIDoc struct {
	Name		string
	Index		bool
	Find		bool
	Post		bool
	Put		bool
	Delete		bool
	ResourceName	string
	FindDoc		*BaseDocSet
	IndexDoc	*BaseDocSet
	PostDoc		*BodyDocSet
	PutDoc		*BodyDocSet
	DeleteDoc	*BaseDocSet
	Field		*FieldDescription
}

//NewDispatch is called to create a new Dispatch instance from a given resource
//struct example (usually the zero value version).  This will panic if the type
//is not a struct or pointer to a struct.  The top-most level struct must have a
//field named "Id" of type "seven5.Id" or this code panics.  It does not check
//nested structs for this property because we allow nested structs that are 
//_not_ resources and if resources are nested the inner struct will be checked
//when this function is called on it (as it is added to the URL mapping, typically).
//The REST interface implementations should be passed as the later parameters, and
//these can be nil.
func NewDispatch(singularName string, r interface{}, i Indexer, f Finder, p Poster, put Puter, d Deleter) *Dispatch {



	t := reflect.TypeOf(r)



	fieldDescription := WalkJsonType(t)




	if !fieldDescription.HasId() {



		panic(fmt.Sprintf("Resources such as %s must contain an Id field of type seven5.Id",
			singularName))
	}




	if t.Kind() == reflect.Struct {



		fieldDescription.Name = t.Name()
	} else {



		fieldDescription.Name = t.Elem().Name()
	}




	return &Dispatch{ResType: r, ResourceName: singularName, Field: fieldDescription, Index: i, Find: f, Post: p, Put: put, Delete: d}
}

//WalkJsonType is the recursive machine that creates a FieldDescription from 
//a go type.  Given a type it returns a pointer to a FieldDescription struct whose
//name is not filled in.  This is public because it's likely to be useful to
//others.
func WalkJsonType(t reflect.Type) *FieldDescription {




	if strings.HasSuffix(t.PkgPath(), "seven5") {



		switch t.Name() {
		case "Floating", "String255", "Textblob", "Integer", "Id", "Boolean":



			return &FieldDescription{Name: "Not_Known_Yet", TypeName: t.Name()}
		}
	}




	if t.Kind() == reflect.Slice {



		nested := WalkJsonType(t.Elem())



		return &FieldDescription{Name: "Not_Known_Yet", Array: nested}
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




			nested := WalkJsonType(f.Type)



			nested.Name = f.Name



			fieldCollection = append(fieldCollection, nested)
		}




		return &FieldDescription{Name:	"Not Known Yet", StructName:	structType.Name(),
			Struct:	fieldCollection}
	}




	switch t.Kind() {
	case reflect.Bool, reflect.Float32, reflect.Float64, reflect.Int, reflect.Int32,
		reflect.Int64, reflect.Int8, reflect.String, reflect.Uint, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uint8:



		panic(fmt.Sprintf("Please use seven5 basic types (instead of %v in pkg %s) so it is not ambiguous"+
			" how to translate them to Json, Dart, and SQL", t.Kind(), t.PkgPath()))
	}



	panic(fmt.Sprintf("Unable to understand type definition %s for conversion "+
		"to a Json/Dart/Sql format", t.Name()))
}
