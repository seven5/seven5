package seven5

import (
	"net/http"
	"seven5/util"
)

//ProcessControllerResult is the result type of a call on the ProcessController plugin.
//Must be public for json encoding.
type ProcessControllerResult struct {
	Error bool
}

// Default echo plugin just prints unformatted version of what you sent
type DefaultProcessController struct {
}

func (self *DefaultProcessController) GetArg() interface{} {
	return nil
}


func (self *DefaultProcessController) Exec(name string, dir string,
	config *ApplicationConfig, request *http.Request, ignored interface{},
	log util.SimpleLogger) interface{} {

	return &ProcessControllerResult{Error: true}
}
