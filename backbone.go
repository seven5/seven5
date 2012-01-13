package seven5

import (
	"bytes"
	"fmt"
	"log"
	"mongrel2"
	"os"
	"reflect"
	"seven5/store"
	"strings"
	"text/template"
)

const (
	//CREATE signals to the validation method that a create operation is in progress.
	OP_CREATE = iota
	//READ signals to the validation method that a read operation is in progress.
	OP_READ
	//UPDATE signals to the validation method that a update operation is in progress.
	OP_UPDATE
	//DELETE signals to the validation method that a delete operation is in progress.
	OP_DELETE
)

//RestfulOp is a type that indications the operation to be performed. It is passed to the Validate
//function of Restful before the operation actually occurs.
type RestfulOp int

//Restful is the key interface for a restful service implementor.  This should be used to make the semantic
//decisions necessary for the particular type implementor such as what user can read a particular
//record or what parameters are required to create a new record.
type Restful interface {
	//Create is called when /api/PLURALNAME is posted to.  The values POSTed are passed to this function
	//in the second parameter.  The currently logged in user (who supplied the params) is the last 
	//parameter and this can be nil if there is no logged in user.
	Create(store store.T, ptrToValues interface{}, session *Session) error
	//Read is called when a url like /api/PLURALNAME/72 is accessed (GET) for the object with id=72.
	//Implementations should implement the correct read semantics, such as security.
	Read(store store.T, ptrToObject interface{}, id uint64, session *Session) error
	//Update is called when a url like /api/PLURALNAME/72 is accessed (PUT) for the object with id=72.
	//Implementations should implement the correct write semantics, such as security.  Note that this
	//call is supposed to be idempotent, unlike Create.
	Update(store store.T, ptrToNewValues interface{}, id uint64, session *Session) error
	//Update is called when a url like /api/PLURALNAME/72 is accessed (DELETE) for the object with id=72.
	//Implementations should implement the correct delete semantics, such as security. 
	Delete(store store.T, id uint64, session *Session) error
	//FindByKey is called when a url like /api/PLURALNAME is accessed (GET) with query parameters
	//that indicate the key to be searched for and the value sought.  The session can be used in
	//cases where there is a need to differentiate based on ownership of the object.  The max
	//number of objects to return is supplied in the last parameter.
	FindByKey(store store.T, key string, value string, session *Session, max int) (interface{}, error)
	//Validate is called BEFORE any other method in this set (except Make).  It can be used to
	//centralize repeated validation checks.  If the validation passes, the receiver should return
	//nil, otherwise a map indicating the field name (key) where the problem was detected and the
	//error message to display to the user as the value.  This will be sent back to the client and
	//the intentded method is NOT called.
	Validate(store store.T, ptrToValues interface{}, id uint64, op RestfulOp, session *Session) map[string]string
	//Make is called with a given id to create a new instance of the appropriate type for use with
	//this interface.  The id may be zero.  This function is needed because the seven5 library does
	//not know the true type of the structures being stored/retrieved.
	Make(id uint64) interface{}
}

//models that have been found
var models = make(map[string][]string)

//BackboneModel is called by the "glue" code between a user-level package (application) and the seven5
//library to indicate that a given type is intended to be used on the client side.  Note that types
//that are to be sent to the client side will be marshalled/unmarshalled by the json library of
//Go and thus will obey structure tags such as json="-" (which prevents the field from arriving
//at the client).  The first parameter should be all lowercase.
func BackboneModel(singularName string, ptrToStruct interface{}) {
	fields := []string{}

	v := reflect.ValueOf(ptrToStruct)
	if v.Kind() != reflect.Ptr {
		panic("backbone models must be a pointer to a struct")
	}
	s := v.Elem()
	if s.Kind() != reflect.Struct {
		panic("backbone models must be a pointer to a struct")
	}
	t := s.Type()
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name := f.Name
		tag := f.Tag.Get("json")
		if tag != "" {
			if tag == "-" {
				continue
			}
		}
		fields = append(fields, name)
	}
	models[strings.ToLower(singularName)] = fields
}

//BackboneServiceis called by the "glue" code between a user-level package (application) and the seven5
//library to indicate to indicate the service that can implement storage and validation for the
//particular name.  Note that the actual URL will be /api/plural and the plural is computed via the
//Plural() function.  The signular name must be english and should be lower case.
func BackboneService(singularName string, svc Restful) {

}

//modelGuise is responsible for shipping backbone models to the client.  It takes in models (structures
//in go) and spits out Javascript that allow the client-side to have a model of the same structure
//with the same field names. The exception is the Id field in go, which is "id" (lowercase) in
//Javascript.
type modelGuise struct {
	//we need the implementation of the default HTTP machinery 
	*HttpRunnerDefault
}

//Name returns "ModelGuise"
func (self *modelGuise) Name() string {
	return "ModelGuise" //used to generate the UniqueId so don"t change this
}

//Pattern returns "/api/seven5/models" because this guise is part of the seven5 api.  This is not a 
//rest API.
func (self *modelGuise) Pattern() string {
	return "/api/seven5/models"
}

//AppStarting is called by the infrastructure to tell the guise that the application is starting.
//Unused for now.
func (self *modelGuise) AppStarting(log *log.Logger, store store.T) error {
	return nil
}

//newModelGuise creates a new ModelGuise.. but only one should be needed in any program.  This is created
//by the infrastructure and user-level code should never need to call this.
func newModelGuise() *modelGuise {
	return &modelGuise{&HttpRunnerDefault{mongrel2.HttpHandlerDefault: &mongrel2.HttpHandlerDefault{new(mongrel2.RawHandlerDefault)}}}
}

func (self *modelGuise) ProcessRequest(req *mongrel2.HttpRequest) *mongrel2.HttpResponse {
	resp := new(mongrel2.HttpResponse)
	resp.ServerId = req.ServerId
	resp.ClientId = []int{req.ClientId}

	buffer := new(bytes.Buffer)

	t := template.Must(template.New("js").Parse(modelTemplate))
	for model, fields := range models {
		data := make(map[string]interface{})
		data["modelName"] = model
		data["modelNamePlural"] = Pluralize(model)
		data["fields"] = fields
		if err := t.Execute(buffer, data); err != nil {
			fmt.Fprintf(os.Stderr, "error writing model:%s\n", err)
			resp.StatusCode = 500
			resp.StatusMsg = err.Error()
			return resp
		}
	}
	resp.ContentLength = buffer.Len()
	resp.Body = buffer
	return resp
}

//Plural takes (should take) a noun in the singular and returns the plural of the noun. 
//Based on http://code.activestate.com/recipes/577781.   Only understands english and lower case
//input.
func Pluralize(singular string) string {
	if singular == "" {
		return ""
	}
	aberrant, ok := aberrant_plural_map[singular]
	if ok {
		return aberrant
	}

	if len(singular) < 4 {
		return singular + "s"
	}

	root := singular
	suffix:=""

	switch {
	case negSlice(-1, root) == "y" && isVowel(negSlice(-2, root))==false:
		root = root[0 : len(root)-1]
		suffix = "ies"
	case negSlice(-1, singular) == "s":
		switch {
		case isVowel(negSlice(-2, singular)):
			if singular[len(singular)-3:] == "ius" {
				root = singular[0 : len(singular)-2]
				suffix = "i"
			} else {
				root = singular[0 : len(singular)-1]
				suffix = "ses"
			}
		default:
			suffix = "es"
		}
	case singular[len(singular)-2:] == "ch", singular[len(singular)-2:] == "sh":
		suffix = "es"
	default:
		suffix = "s"
	}

	return root + suffix
}

//aberrant_plural_map shows english is a weird and wonderful language
var aberrant_plural_map = map[string]string{
	"appendix":   "appendices",
	"barracks":   "barracks",
	"cactus":     "cacti",
	"child":      "children",
	"criterion":  "criteria",
	"deer":       "deer",
	"echo":       "echoes",
	"elf":        "elves",
	"embargo":    "embargoes",
	"focus":      "foci",
	"fungus":     "fungi",
	"goose":      "geese",
	"hero":       "heroes",
	"hoof":       "hooves",
	"index":      "indices",
	"knife":      "knives",
	"leaf":       "leaves",
	"life":       "lives",
	"man":        "men",
	"mouse":      "mice",
	"nucleus":    "nuclei",
	"person":     "people",
	"phenomenon": "phenomena",
	"potato":     "potatoes",
	"self":       "selves",
	"syllabus":   "syllabi",
	"tomato":     "tomatoes",
	"torpedo":    "torpedoes",
	"veto":       "vetoes",
	"woman":      "women",
}

//vowels is the set of vowels
var vowels = []string{"a", "e", "i", "o", "u"}

//negslice can compute a negative slice ala python
func negSlice(n int, s string) string {
	if n >= 0 {
		panic("bad negative slice index!")
	}
	if -n > len(s) {
		panic("negative slice index is too big")
	}
	i := len(s) + n //subtraction
	return s[i : i+1]
}

//isVowel returns true if a string of 1 char is a vowel
func isVowel(s string) bool {
	if len(s) != 1 {
		panic("bad call to isVowel")
	}
	for _,v := range vowels {
		if s == v {
			return true
		}
	}
	return false
}

const modelTemplate = `
window.{{.modelName}} = Backbone.Model.extend({
	{{range .fields}} {{.}} : null,
	{{end}}
defaults: function(){
	this.urlRoot="/{{.modelNamePlural}}"
}
});`
