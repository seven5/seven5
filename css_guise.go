package seven5

import (
	"fmt"
	"os"
	"mongrel2"
	"seven5/css"
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
	fmt.Fprintf(os.Stderr,"css guise working on %s\n",config.Path)
	return nil
}

//create a new one... but only one should be needed in any program
func NewCssGuise() *CssGuise {
	return &CssGuise{&HttpRunnerDefault{&mongrel2.HttpHandlerDefault{new (mongrel2.RawHandlerDefault)}}}
}

//Handle a single request of the HTTP level of mongrel.  This simply echos back to the 
//input information to the client and it will be displayed on any path starting with /echo.
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

//RegisterStyleSheet takes a css.StmtSeq and stores it in the CSS guise cache for the given
//name. The statement sequence is only evaluted once, at the time it is register.
func RegisterStylesheet(name string, stmt css.StmtSeq) {
	sheet[name]=fmt.Sprintf("%v",stmt)
}