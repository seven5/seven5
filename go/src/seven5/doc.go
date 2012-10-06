package seven5

import (
	"fmt"
	_ "github.com/russross/blackfriday"
	"net/http"
	"reflect"
	"strings"
)

//ResourceDoc is used by the api/documentation generation code. It maintains a list of all the
//interfaces that this resource supports and what is known about it.  This is public because
//it is likely any implementation of Handler will need this to generated api/doc.
type ResourceDoc struct {
	ResType     interface{}
	GETSingular string
	GETPlural   string
	Index       Indexer
	Find        Finder
}

//Field description gives information about a particular field and this is part of what is passed
//over the wire to describe a resource.  This describes the type that is being used for a particular
//resource.  If Array or Struct fields are not nil, then there is a nested structure or array
//inside the struct.
type FieldDescription struct {
	Name string
	GoType string
	DartType string
	SQLType string
	Array *FieldDescription
	Struct *[]FieldDescription
}

//ResourceDescription is the type passed over the wire to describe how a particular can be called
//and what fields the objects have that it manipulates.
type ResourceDescription struct {
	Name string
	Index bool
	Find bool
	GETSingular string
	GETPlural string
	CollectionURL string
	CollectionDoc []string
	ResourceURL string
	ResourceDoc[] string
	Fields *[]FieldDescription
}

//NewResourceDoc is called to create a new ResourceDoc instance from a given type.  The passed
//value should be a struct (not a pointer to a struct) that describes the json exchanged between
//the client and server.  If the passed type isn't a struct, this function panics.
func NewResourceDoc(r interface{}) *ResourceDoc {
	if r == nil || reflect.TypeOf(r).Kind() != reflect.Struct {
		panic(fmt.Sprintf("Can't create a resource from an object that is not a struct (%T)!", r))
	}
	return &ResourceDoc{ResType: r}
}

//isLiveDocRequest tests a request to see if this is a request for live documentation.  Only used
//with a real network.
func isLiveDocRequest(req *http.Request) bool {
	qparams := toSimpleMap(map[string][]string(req.URL.Query()))
	if len(qparams) != 1 {
		return false
	}
	someBool, ok := qparams["livedoc"]
	if !ok {
		return false
	}
	return strings.ToLower(strings.Trim(someBool, " ")) == "true" || strings.Trim(someBool, " ") == "1"
}

//GenerateDoc walks through the registered resources to find the one requested and the compute
//the description of it.  Note that iisLiveDocRequest(req *http.Request)t will compute the same result for singular or plural
//since these really correspond to the same logical entity.
func (self *SimpleHandler) GenerateDoc(uriPath string) *ResourceDescription {
	result := &ResourceDescription{}
	path, _, _ := self.resolve(uriPath)
	
	//no such path?
	if path=="" {
		return nil
	}
	doc := self.doc[path]
	result.Name = reflect.TypeOf(doc.ResType).Name()
	
	result.Fields = walkJsonType(reflect.TypeOf(doc.ResType))
	if doc.Find!= nil{
		result.Find = true
		result.ResourceURL = fmt.Sprintf("/%s/123",doc.GETSingular)
		result.ResourceDoc = doc.Find.FindDoc()
		result.GETSingular = doc.GETSingular
	}
	if doc.Index!= nil{
		result.Index = true
		result.CollectionURL = fmt.Sprintf("/%s/",doc.GETPlural)
		result.CollectionDoc = doc.Index.IndexDoc()
		result.GETSingular = doc.GETPlural
	}
	return result
}

//walkJsonType recursively walks the type structure provided.  This is called recursively 
//if there is a nested struct or array. The top level type must be a struct, not a simple
//type or array. 
//walkJsonType will panic if it finds something that cannot be correctly flattened to json,
//converted to Dart, and converted to SQL.
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
			g, dart, sql := toMultipleTypeNames(k)
			result = append(result, FieldDescription{Name: name, GoType: g, DartType: dart, 
				SQLType: sql})
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
func toMultipleTypeNames(k reflect.Kind) (string, string, string){
	switch k {
	case reflect.Bool:
		return "bool", "bool", "int"
	case reflect.String:
		return "string", "String", "varchar(255)"
	case reflect.Int64:
		return "int64", "int", "integer"
	case reflect.Float64:
		return "float64", "double", "real"
	}
	panic(fmt.Sprintf("don't know how to convert %s to a dart and sql type!",k.String()))
}
