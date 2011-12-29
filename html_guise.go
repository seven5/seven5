package seven5

import (
	"fmt"
//	"os"
	"mongrel2"
	"seven5/dsl"
	"strings"
)

var doc = make(map[string]string)

type HtmlGuise struct {
	//we need the implementation of the default HTTP machinery 
	*HttpRunnerDefault
}

func (self *HtmlGuise) Name() string {
	return "HtmlGuise" //used to generate the UniqueId so don't change this
}

func (self *HtmlGuise) IsJson() bool {
	return false
}

func (self *HtmlGuise) Pattern() string {
	return "/(\\a.*html)"
}

func (self *HtmlGuise) AppStarting(config *ProjectConfig) error {
	return nil
}

//create a new one... but only one should be needed in any program
func NewHtmlGuise() *HtmlGuise {
	return &HtmlGuise{&HttpRunnerDefault{mongrel2.HttpHandlerDefault: &mongrel2.HttpHandlerDefault{new (mongrel2.RawHandlerDefault)}}}
}

//Handle a single request of the HTTP level of mongrel. 
func (self *HtmlGuise) ProcessRequest(req *mongrel2.HttpRequest) *mongrel2.HttpResponse {
	path:=req.Header["PATH"]
	path=path[len("/"):]
	
	resp:=new (mongrel2.HttpResponse)
	resp.ServerId= req.ServerId
	resp.ClientId = []int{req.ClientId}
	
	s,ok := doc[path]
	if !ok{
		resp.StatusCode=404
		resp.StatusMsg="No such document."
		return resp
	}
	resp.ContentLength=len(s)
	resp.Body = strings.NewReader(s)
	return resp
}

//RegisterDocument creates a mapping between a name (at the HTTP level a resource) and an
//dsl.Document object.
func RegisterDocument(s dsl.Document) {
	doc[s.Name]=fmt.Sprintf("%s",s)
}