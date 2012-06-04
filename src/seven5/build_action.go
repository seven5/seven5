package seven5

import (
	"net/http"
	"seven5/util"
)

//Build Action allows other programs (roadie) to do things and segregates 
//these actions from each other.  It holds a reference to the wire which
//is nil if we have not successfully built seven5.
type BuildAction struct {
	Wire *Wire
}

//buildSeven5 does the work of building the boostrap seven5 instance. It
//does not dispatch anything, but creates the wire that other actions
//need to do a dispatch to the seven5 instance.
func (self *BuildAction) BuildSeven5(writer http.ResponseWriter,
	request *http.Request, logger util.SimpleLogger) bool {
	currentSeven5 := Bootstrap(writer, request, logger)
	if currentSeven5 == "" {
		self.Wire = nil
	} else {
		self.Wire = NewWire(currentSeven5)
	}
	return self.Wire != nil
}

//buildUserLib tries to build a user library and returns any error that
//happened in the process
func (self *BuildAction) BuildUserLib(cwd string, writer http.ResponseWriter,
	request *http.Request, logger util.SimpleLogger) error {
	_, err:=self.dispatchAny(BUILDUSERLIB, cwd, writer, request, 
		nil, logger)
	return err
}

//Destroy all the generated files in the user's app directory.
func (self *BuildAction) DestroyGeneratedFile(cwd string, writer http.ResponseWriter,
	request *http.Request, logger util.SimpleLogger) error {
	_, err:=self.dispatchAny(DESTROYGENERATEDFILE, cwd, writer, request, 
		nil, logger)
	return err
}

//validateProject checks that the structure of thee project is ok.  Returns
//an error if occurred
func (self *BuildAction) ValidateProject(cwd string, writer http.ResponseWriter,
	request *http.Request, logger util.SimpleLogger) error {
	_,err:=self.dispatchAny(VALIDATEPROJECT, cwd, writer, request, nil, logger);
	return err
}

//Explode types returns the information about the types in a user library.
func (self *BuildAction) ExplodeType(cwd string, writer http.ResponseWriter,
	request *http.Request, arg *ExplodeTypeArg, 
	logger util.SimpleLogger) (*ExplodeTypeResult,error) {
	raw, err:=self.dispatchAny(EXPLODETYPE, cwd, writer, request, arg, logger);
		if err!=nil {
		return nil,err
	}
	result:=raw.(*ExplodeTypeResult)
	return result,nil
}

//Explode types returns the information about the types in a user library.
func (self *BuildAction) ProcessVocab(cwd string, writer http.ResponseWriter,
	request *http.Request, arg *ProcessVocabArg, 
	logger util.SimpleLogger) (*ProcessVocabResult,error) {
	raw, err:=self.dispatchAny(PROCESSVOCAB, cwd, writer, request, arg, logger);
	if err!=nil {
		return nil,err
	}
	result:=raw.(*ProcessVocabResult)
	return result,nil
}

//dispatchAny can send any of the messages to  it does this via
//the wire object inside this roadie action.
func (self *BuildAction) dispatchAny(name string, cwd string, writer http.ResponseWriter,
	request *http.Request, arg interface{}, logger util.SimpleLogger) (interface{}, error) {

	retVal, err := self.Wire.Dispatch(name, cwd, writer, request, arg, logger)
	switch err {
	case WIRE_SEMANTIC_ERROR, nil:
		/*already displayed the feedback*/
	case WIRE_PROCESS_ERROR:
		logger.Error("%s", err)
	default:
		logger.Error("Can't continue due to error in dispatch: %s", err)
	}
	return retVal, err
}
