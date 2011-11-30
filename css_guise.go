package seven5

import (
	"fmt"
	"os"
	"mongrel2"
)
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
	fmt.Fprintf(os.Stderr,"css guise process: %s\n",path)
	return nil
}