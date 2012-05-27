package seven5

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"seven5/plugin"
	"seven5/util"
	"os/exec"
	"strings"
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
	request *http.Request, logger util.SimpleLogger) {

	switch cmd {
	case plugin.VALIDATE_PROJECT:
		self.validateProject("", logger) //doesn't need the request
	default:
		logger.Error("don't understand command! can't send to seven5! your groupies must be broken!")
	}
	return
}

//validateProject does the marshal/unmarshal pir for checking a project
func (self *Wire) validateProject(dir string, logger util.SimpleLogger) {
	var err error
	var jsonBlob string
	if dir == "" {
		if dir, err = os.Getwd(); err != nil {
			logger.Error("unable to get the current working dir: %s", err)
			return
		}
	}
	cmd, args := self.marshal(dir, plugin.VALIDATE_PROJECT,
		&plugin.ValidateProjectArgs{}, logger)
	if jsonBlob = self.runProcess(plugin.VALIDATE_PROJECT, cmd, args, logger); jsonBlob=="" {
		return 
	}
	
	// 
	// RESULTS
	//
	result := plugin.ValidateProjectResult{}
	dec := json.NewDecoder(strings.NewReader(jsonBlob))
	dec.Decode(&result)
	if (result.Error) {
		logger.Debug("here is marshalled result from seven5:")
		logger.DumpJson(jsonBlob)
	}
	return
}

//runProcess is called to take the two bunches of json arguments and invoke
//the process of Seven5 with those. It captures the output of the process
//and returns the marshalled result that it got from stdout.
func (self *Wire) runProcess(name string, cmd string, args string, 
	logger util.SimpleLogger) string {	
	shellCommand := exec.Command(self.path, cmd, args)	
	out, err := shellCommand.CombinedOutput()
	allOutput := string(out)
	index := strings.Index(allOutput, plugin.MARKER)
	if index<0 {
		logger.Error("got some bad out putfrom seven5:")
		logger.DumpTerminal(allOutput);
		return ""
	}
	pos := len(plugin.MARKER) + index
	logMsg := allOutput[pos:]
	marshalledText := allOutput[:index]
	
	if err!=nil {
		logger.Debug("Error running seven5! here are the arguments we sent:")
		logger.DumpJson(cmd)
		logger.DumpJson(args)
		logger.Raw(logMsg)
		logger.Error("error running command %s was %s",name,err.Error())
		return ""
	}
	
	logger.Raw(logMsg)
	return marshalledText
	
}

//HaveSeven5 returns true if it currently has a seven5 binary to talk to.
//This is false when there is a problem building seven5 itself.
func (self *Wire) HaveSeven5() bool {
	return self.path!=""
}

//marshal is called to create the two bunches of json that we need to
//talk to seven5
func (self *Wire) marshal(dir string, name string, args interface{}, logger util.SimpleLogger) (string,string){
	var cmdBuffer bytes.Buffer
	var argBuffer bytes.Buffer
	cmdEnc := json.NewEncoder(&cmdBuffer)
	argEnc := json.NewEncoder(&argBuffer)
	cmd := &plugin.Command{
		AppDirectory: dir,
		Name:         plugin.VALIDATE_PROJECT,
	}
	argEnc.Encode(args)
	cmdEnc.Encode(cmd)

	return cmdBuffer.String(), argBuffer.String()
}
