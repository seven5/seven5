package main

import (
	"net/http"
	"seven5"
	"seven5/util"
)

//roadie action allows the roadie to do things and segregates these actions
//from each other.
type roadieAction struct {
	wire *seven5.Wire
}

//buildSeven5 does the work of building the boostrap seven5 instance. It
//does not dispatch anything, but creates the wire that other actions
//need to do a dispatch to the seven5 instance.
func (self *roadieAction) buildSeven5(writer http.ResponseWriter,
	request *http.Request, logger util.SimpleLogger) bool {
	currentSeven5 := seven5.Bootstrap(writer, request, logger)
	if currentSeven5 == "" {
		self.wire = nil
	} else {
		self.wire = seven5.NewWire(currentSeven5)
	}
	return self.wire != nil
}

//buildUserLib tries to build a user library and returns any error that
//happened in the process
func (self *roadieAction) buildUserLib(cwd string, writer http.ResponseWriter,
	request *http.Request, logger util.SimpleLogger) error {
	return self.dispatchAny(seven5.BUILDUSERLIB, cwd, writer, request, logger)
}

//validateProject checks that the structure of thee project is ok.  Returns
//an error if occurred
func (self *roadieAction) validateProject(cwd string, writer http.ResponseWriter,
	request *http.Request, logger util.SimpleLogger) error {
	return self.dispatchAny(seven5.VALIDATEPROJECT, cwd, writer, request, logger);
}

//dispatchAny can send any of the messages to seven5. it does this via
//the wire object inside this roadie action.
func (self *roadieAction) dispatchAny(name string, cwd string, writer http.ResponseWriter,
	request *http.Request, logger util.SimpleLogger) error {

	err := self.wire.Dispatch(name, cwd, writer, request, logger)
	switch err {
	case seven5.WIRE_SEMANTIC_ERROR, nil:
		/*already displayed the feedback*/
	case seven5.WIRE_PROCESS_ERROR:
		logger.Error("%s", err)
	default:
		logger.Error("Can't continue due to error in dispatch: %s", err)
	}
	return err
}
