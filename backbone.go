package seven5

import (
	"bytes"
	"fmt"
	"mongrel2"
	"os"
	"text/template"
	"strings"
	"seven5/store"
	"reflect"
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

type RestfulOp int

//Restful is the key interface for a restful service implementor.  This should be used to make the semantic
//decisions necessary for the particular type implementor such as what user can read a particular
//record or what parameters are required to create a new record.
type Restful interface {
	//Create is called when /api/PLURALNAME is posted to.  The values POSTed are passed to this function
	//in the second parameter.  The currently logged in user (who supplied the params) is the last 
	//parameter and this can be nil if there is no logged in user.
	Create(store store.T, ptrToValues interface{},session *Session) error
	//Read is called when a url like /api/PLURALNAME/72 is accessed (GET) for the object with id=72.
	//Implementations should implement the correct read semantics, such as security.
	Read(store store.T, ptrToObject interface{}, id uint64,session *Session) error 
	//Update is called when a url like /api/PLURALNAME/72 is accessed (PUT) for the object with id=72.
	//Implementations should implement the correct write semantics, such as security.  Note that this
	//call is supposed to be idempotent, unlike Create.
	Update(store store.T, ptrToNewValues interface{},id uint64,session *Session) error
	//Update is called when a url like /api/PLURALNAME/72 is accessed (DELETE) for the object with id=72.
	//Implementations should implement the correct delete semantics, such as security. 
	Delete(store store.T, id uint64,session *Session) error
	//FindByKey is called when a url like /api/PLURALNAME is accessed (GET) with query parameters
	//that indicate the key to be searched for and the value sought.  The session can be used in
	//cases where there is a need to differentiate based on ownership of the object.  The max
	//number of objects to return is supplied in the last parameter.
	FindByKey(store store.T,key string, value string, session *Session, max int) (interface{}, error)
	//Validate is called BEFORE any other method in this set (except Make).  It can be used to
	//centralize repeated validation checks.  If the validation passes, the receiver should return
	//nil, otherwise a map indicating the field name (key) where the problem was detected and the
	//error message to display to the user as the value.  This will be sent back to the client and
	//the intentded method is NOT called.
	Validate(store store.T, ptrToValues interface{}, id uint64, op RestfulOp,session *Session) map[string]string
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
//at the client).
func BackboneModel(name string, ptrToStruct interface{}) {
	fields:=[]string{};
	
	v:=reflect.ValueOf(ptrToStruct);
	if v.Kind()!=reflect.Ptr {
		panic("backbone models must be a pointer to a struct");
	}
	s:=v.Elem();
	if s.Kind()!=reflect.Struct {
		panic("backbone models must be a pointer to a struct");
	}
	t:=s.Type();
	for i:=0; i<t.NumField();i++ {
		f:=t.Field(i);
		name:=f.Name;
		tag:=f.Tag.Get("json");
		if tag!="" {
			if tag=="-" {
				continue;
			}
		}
		fields=append(fields,name)
	}
	models[name]=fields;
}

//modelGuise is responsible for shipping backbone models to the client.  It takes in models (structures
//in go) and spits out Javascript that allow the client-side to have a model of the same structure
//with the same field names. The exception is the Id field in go, which is 'id' (lowercase) in
//Javascript.
type modelGuise struct {
	//we need the implementation of the default HTTP machinery 
	*HttpRunnerDefault
}

//Name returns "ModelGuise"
func (self *modelGuise) Name() string {
	return "ModelGuise" //used to generate the UniqueId so don't change this
}

//Pattern returns "/api/seven5/models" because this guise is part of the seven5 api.  This is not a 
//rest API.
func (self *modelGuise) Pattern() string {
	return "/api/seven5/models"
}

//AppStarting is called by the infrastructure to tell the guise that the application is starting.
//Unused for now.
func (self *modelGuise) AppStarting(config *projectConfig) error {
	return nil
}

//NewModelGuise creates a new ModelGuise.. but only one should be needed in any program.  This is created
//by the infrastructure and user-level code should never need to call this.
func newModelGuise() *modelGuise {
	return &modelGuise{&HttpRunnerDefault{mongrel2.HttpHandlerDefault: &mongrel2.HttpHandlerDefault{new(mongrel2.RawHandlerDefault)}}}
}

func (self *modelGuise) ProcessRequest(req *mongrel2.HttpRequest) *mongrel2.HttpResponse {
	resp := new(mongrel2.HttpResponse)
	resp.ServerId = req.ServerId
	resp.ClientId = []int{req.ClientId}

	buffer := new(bytes.Buffer)

	t:=template.Must(template.New("js").Parse(MODEL_TEMPLATE))
	for model, fields := range models {
		data:=make(map[string]interface{})
		data["modelName"]=model
		data["modelNamePlural"]=plural(model)
		data["fields"]=fields
		if err := t.Execute(buffer, data); err != nil {
			fmt.Fprintf(os.Stderr,"error writing model:%s\n",err)
			resp.StatusCode=500
			resp.StatusMsg=err.Error()
			return resp
		}
	}
	resp.ContentLength = buffer.Len()
	resp.Body = buffer
	return resp
}

func plural(n string) string {
	return strings.ToLower(n[0:1])+n[1:]+"s"
}


const MODEL_TEMPLATE = `
window.{{.modelName}} = Backbone.Model.extend({
	{{range .fields}} {{.}} : null,
	{{end}}
defaults: function(){
	this.urlRoot="/{{.modelNamePlural}}"
}
});`
