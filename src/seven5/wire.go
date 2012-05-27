package seven5

import (
	"net/http"
	"seven5/util"
	"seven5/plugin"
)

//Wire represents the way to talk to a seven5 program. It handles marshalling
//calling commands on the other side (invoking seven5), and unmarshalling.
type Wire struct {
	path string
}

//NewWire creates a new Wire instance.  It is ok to pass it "" as the path
//to the seven5 executable because it understands how to generate errors in
//that case.
func NewWire(path string) *Wire {
	return &Wire{path}
}

//Dispatch is called to invoke a command.  It returns nothing because it
//puts errors in the log connected to this response writer.
func (self *Wire) Dispatch(cmd string, writer http.ResponseWriter, 
	request *http.Request) {

	logger := util.NewHtmlLogger(util.DEBUG, true, writer)
	switch cmd {
	case plugin.VALIDATE_PROJECT:
		self.validateProject(logger) //doesn't need the request
	default:
		logger.Error("don't understand command! can't send to seven5! your groupies must be broken!")
	}
	return
}

func (self *Wire) validateProject(logger util.SimpleLogger) {
}