package seven5

import (
	"fmt"
//	"os"
	"mongrel2"
	"seven5/dsl"
	"strings"
)

var sheet = make(map[string]string)

type CssGuise struct {
	//we need the implementation of the default HTTP machinery 
	*HttpRunnerDefault
}

func (self *CssGuise) Name() string {
	return "CssGuise" //used to generate the UniqueId so don't change this
}

func (self *CssGuise) IsJson() bool {
	return false
}

func (self *CssGuise) Pattern() string {
	return "/css/(\\a.*css)"
}

func (self *CssGuise) AppStarting(config *ProjectConfig) error {
	return nil
}

//create a new one... but only one should be needed in any program
func NewCssGuise() *CssGuise {
	return &CssGuise{&HttpRunnerDefault{mongrel2.HttpHandlerDefault: &mongrel2.HttpHandlerDefault{new (mongrel2.RawHandlerDefault)}}}
}

//Handle a single request of the HTTP level of mongrel.  
func (self *CssGuise) ProcessRequest(req *mongrel2.HttpRequest) *mongrel2.HttpResponse {
	path:=req.Header["PATH"]
	path=path[len("/css/"):]
	
	resp:=new (mongrel2.HttpResponse)
	resp.ServerId= req.ServerId
	resp.ClientId = []int{req.ClientId}
	
	s,ok := sheet[path]
	if !ok{
		resp.StatusCode=404
		resp.StatusMsg="No such stylesheet."
		return resp
	}
	resp.ContentLength=len(s)
	resp.Body = strings.NewReader(s)
	return resp
}

//RegisterStyleSheet creates a mapping between an HTTP level resource and a dsl.StyleSheet
//object. 
func RegisterStylesheet(s dsl.StyleSheet) {
	sheet[s.Name]=fmt.Sprintf("%v",s)
}