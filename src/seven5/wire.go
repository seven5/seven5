package seven5

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"seven5/groupie"
	"seven5/util"
	"strings"
)

//Wire represents the way to talk to a seven5 program. It handles marshalling
//calling commands on the other side (invoking seven5), and unmarshalling.
type Wire struct {
	path string
}

//wireStubs are the proxies for various commands
type wireStub func(http.ResponseWriter, string, 
	util.SimpleLogger) *http.Response

//simulate const map
func STUBS() map[string]wireStub {
	return map[string]wireStub{
		groupie.ECHO: echoStub,
		groupie.VALIDATEPROJECT: validateProjectStub,
		groupie.PROCESSCONTROLLER: processControllerStub,
	}
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
	request *http.Request, logger util.SimpleLogger) *http.Response {
	var cfg *groupie.ApplicationConfig
	var jsonBlob string
	//marshalling for requset is special, not using standard json path
	req := util.MarshalRequest(request, logger)
	
	//calculate response, if any
	cfgBlob, reqBlob := self.marshal(cfg, req, logger)
	if jsonBlob = self.runSeven5(cmd, "", cfgBlob, reqBlob, logger); jsonBlob == "" {
		return nil
	}
	
	//need to handle the custom results
	stub := STUBS()[cmd]
	response := stub(writer, jsonBlob, logger)

	return response
}

//echo does the marshal/unmarshal pir for echo plugin
func echoStub(writer http.ResponseWriter,jsonBlob string, logger util.SimpleLogger) *http.Response {
	var err error
	// 
	// RESULTS
	//
	result := groupie.EchoResult{}
	dec := json.NewDecoder(strings.NewReader(jsonBlob))
	dec.Decode(&result)
	if result.Error {
		logger.Debug("here is marshalled result from seven5:")
		logger.DumpJson(jsonBlob)
	} else {
		_, err = writer.Write([]byte(result.Body))
		if err != nil {
			//not sure what to do here
			fmt.Fprintf(os.Stdout, "Problems with writing to HTTP output: %s\n",
				err.Error())
		}
	}
	return &result.Response

}

//processController gets the response to a rpocess controller
func processControllerStub(writer http.ResponseWriter,
	jsonBlob string, logger util.SimpleLogger) *http.Response {
	return nil
}

//validateProject handles the response to to validate proj
func validateProjectStub(writer http.ResponseWriter,
	jsonBlob string, logger util.SimpleLogger) *http.Response {

	// 
	// RESULTS
	//
	result := groupie.ValidateProjectResult{}
	dec := json.NewDecoder(strings.NewReader(jsonBlob))
	dec.Decode(&result)
	if result.Error {
		logger.Debug("here is marshalled result from seven5:")
		logger.DumpJson(jsonBlob)
	}
	return nil
}

//runProcess is called to take the two bunches of json arguments and invoke
//the process of Seven5 with those. It captures the output of the process
//and returns the marshalled result that it got from stdout.
func (self *Wire) runSeven5(name string, dir string, config string,
	browserReq string, logger util.SimpleLogger) string {
	var err error

	if dir == "" {
		if dir, err = os.Getwd(); err != nil {
			logger.Panic("unable to get the current working dir: %s", err)
		}
	}
	
	shellCommand := exec.Command(self.path, name, dir, config, browserReq)
	out, err := shellCommand.CombinedOutput()
	allOutput := string(out)
	index := strings.Index(allOutput, groupie.MARKER)
	if index < 0 {
		logger.Error("got some bad out putfrom seven5:")
		logger.DumpTerminal(allOutput)
		return ""
	}
	pos := len(groupie.MARKER) + index
	logMsg := allOutput[pos:]
	marshalledText := allOutput[:index]

	if err != nil {
		logger.Debug("Error running seven5! here are the arguments we sent after '%s' and '%s'",
			name,dir)
		logger.DumpJson(config)
		logger.DumpJson(browserReq)
		logger.Raw(logMsg)
		logger.Error("error running command %s was %s", name, err.Error())
		return ""
	}

	logger.Raw(logMsg)
	return marshalledText

}

//HaveSeven5 returns true if it currently has a seven5 binary to talk to.
//This is false when there is a problem building seven5 itself.
func (self *Wire) HaveSeven5() bool {
	return self.path != ""
}

//marshal is called to create the two bunches of json that we need to
//talk to seven5
func (self *Wire) marshal(cfg *groupie.ApplicationConfig,
	req *util.BrowserRequest, logger util.SimpleLogger) (string, string) {
	var cfgBuffer bytes.Buffer
	var reqBuffer bytes.Buffer

	cfgEnc := json.NewEncoder(&cfgBuffer)
	reqEnc := json.NewEncoder(&reqBuffer)
	cfgEnc.Encode(cfg)
	reqEnc.Encode(req)

	return cfgBuffer.String(), reqBuffer.String()
}
