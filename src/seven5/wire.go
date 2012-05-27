package seven5

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"seven5/plugin"
	"seven5/util"
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
	request *http.Request, logger util.SimpleLogger) *http.Response {
	var response *http.Response
	req := util.MarshalRequest(request, logger)

	switch cmd {
	case plugin.VALIDATE_PROJECT:
		response = self.validateProject(&req, logger) 
	case plugin.ECHO:
		response = self.echo(&req, writer, logger) 
	default:
		logger.Error("don't understand command! can't send to seven5! your groupies must be broken!")
		response = nil
	}
	return response
}

//echo does the marshal/unmarshal pir for echo plugin
func (self *Wire) echo(req *util.BrowserRequest, writer http.ResponseWriter, logger util.SimpleLogger) *http.Response {
	var jsonBlob string
	var err error

	eArgs := plugin.EchoArgs{}
	cmd, args, r := self.marshal("", plugin.ECHO, &eArgs, req, logger)
	if jsonBlob = self.runProcess(plugin.ECHO, cmd, args, r, logger); jsonBlob == "" {
		return nil
	}

	// 
	// RESULTS
	//
	result := plugin.EchoResult{}
	dec := json.NewDecoder(strings.NewReader(jsonBlob))
	dec.Decode(&result)
	if result.Error {
		logger.Debug("here is marshalled result from seven5:")
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

//validateProject does the marshal/unmarshal pir for checking a project
func (self *Wire) validateProject(req *util.BrowserRequest, 
	logger util.SimpleLogger) *http.Response {
	var jsonBlob string
	cmd, args, r := self.marshal("", plugin.VALIDATE_PROJECT,
		&plugin.ValidateProjectArgs{}, req, logger)
	if jsonBlob = self.runProcess(plugin.VALIDATE_PROJECT, cmd, args, r, logger); jsonBlob == "" {
		return nil
	}

	// 
	// RESULTS
	//
	result := plugin.ValidateProjectResult{}
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
func (self *Wire) runProcess(name string, cmd string, args string,
	browserReq string, logger util.SimpleLogger) string {
	shellCommand := exec.Command(self.path, cmd, args, browserReq)
	out, err := shellCommand.CombinedOutput()
	allOutput := string(out)
	index := strings.Index(allOutput, plugin.MARKER)
	if index < 0 {
		logger.Error("got some bad out putfrom seven5:")
		logger.DumpTerminal(allOutput)
		return ""
	}
	pos := len(plugin.MARKER) + index
	logMsg := allOutput[pos:]
	marshalledText := allOutput[:index]

	if err != nil {
		logger.Debug("Error running seven5! here are the arguments we sent:")
		logger.DumpJson(cmd)
		logger.DumpJson(args)
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
func (self *Wire) marshal(dir string, name string, args interface{}, 
	req *util.BrowserRequest, logger util.SimpleLogger) (string, string, string) {
	var cmdBuffer bytes.Buffer
	var argBuffer bytes.Buffer
	var reqBuffer bytes.Buffer
	var err error

	if dir == "" {
		if dir, err = os.Getwd(); err != nil {
			logger.Panic("unable to get the current working dir: %s", err)
		}
	}

	cmdEnc := json.NewEncoder(&cmdBuffer)
	argEnc := json.NewEncoder(&argBuffer)
	reqEnc := json.NewEncoder(&reqBuffer)
	cmd := &plugin.Command{
		AppDirectory: dir,
		Name:         name,
	}
	argEnc.Encode(args)
	cmdEnc.Encode(cmd)
	reqEnc.Encode(req)

	return cmdBuffer.String(), argBuffer.String(), reqBuffer.String()
}
