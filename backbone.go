package seven5

import (
	"bytes"
	"fmt"
	"mongrel2"
	"os"
	"text/template"
	"strings"
)

//0 is not a valid id, signal value
const ID_NONE = 0

type Restful interface {
	Create(values map[string]interface{}) int64
	Delete(id int64)
	Read(id int64) //can be ID_NONE for all items
	Update(id int64, values map[string]interface{}) bool
}

var models = make(map[string][]string)

func BackboneModel(name string, fields ...string) {
	models[name] = fields
}

func RESTService(name string, service Restful) {
}

//
// Guise that lets use squirt generated Javascript models to the client
//
type ModelGuise struct {
	//we need the implementation of the default HTTP machinery 
	*HttpRunnerDefault
}

func (self *ModelGuise) Name() string {
	return "ModelGuise" //used to generate the UniqueId so don't change this
}

func (self *ModelGuise) IsJson() bool {
	return false
}

func (self *ModelGuise) Pattern() string {
	return "/seven5/models"
}

func (self *ModelGuise) AppStarting(config *ProjectConfig) error {
	fmt.Fprintf(os.Stderr, "model guise working on %s\n", config.Path)
	return nil
}

//create a new one... but only one should be needed in any program
func NewModelGuise() *ModelGuise {
	return &ModelGuise{&HttpRunnerDefault{mongrel2.HttpHandlerDefault: &mongrel2.HttpHandlerDefault{new(mongrel2.RawHandlerDefault)}}}
}

func (self *ModelGuise) ProcessRequest(req *mongrel2.HttpRequest) *mongrel2.HttpResponse {
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
