package seven5

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"seven5/util"
	"strings"
	"time"
)

//Wire represents the way to talk to a seven5 program. It handles marshalling
//calling commands on the other side (invoking seven5), and unmarshalling.
type Wire struct {
	path string
}

//wireStubs are the proxies for various commands that will run in the other
//process
type wireStub struct {
	result         func() interface{}
	resultHasError func(interface{}) bool
	resultBody     func(interface{}) string
}

//stubs is a constant map that just does some type conversions.  This is here
//primarily to isolate the one necessary checked conversion for each type of
//result.
var (
	STUBS = map[string]*wireStub{
		ECHO: &wireStub{
			func() interface{} { return &EchoResult{} },
			func(v interface{}) bool { return v.(*EchoResult).Error },
			func(v interface{}) string { return v.(*EchoResult).Body }},
		VALIDATEPROJECT: &wireStub{
			func() interface{} { return &ValidateProjectResult{} },
			func(v interface{}) bool { return v.(*ValidateProjectResult).Error },
			nil},
		PROCESSCONTROLLER: &wireStub{
			func() interface{} { return &ProcessControllerResult{} },
			func(v interface{}) bool { return v.(*ProcessControllerResult).Error },
			nil},
		PROCESSVOCAB: &wireStub{
			func() interface{} { return &ProcessVocabResult{} },
			func(v interface{}) bool { return v.(*ProcessVocabResult).Error },
			nil},
		BUILDUSERLIB: &wireStub{
			func() interface{} { return &BuildUserLibResult{} },
			func(v interface{}) bool { return v.(*BuildUserLibResult).Error },
			nil},
		EXPLODETYPE: &wireStub{
			func() interface{} { return &ExplodeTypeResult{} },
			func(v interface{}) bool { return v.(*ExplodeTypeResult).Error},
			nil},
		DESTROYGENERATEDFILE: &wireStub{
			func() interface{} { return &DestroyGeneratedFileResult{} },
			func(v interface{}) bool { return v.(*DestroyGeneratedFileResult).Error},
			nil},
	}
	WIRE_SEMANTIC_ERROR = errors.New("Semantic error in seven5 command")
	WIRE_PROCESS_ERROR  = errors.New("Unable to run the seven5 command")
)

//NewWire creates a new Wire instance.  It is ok to pass it "" as the path
//to the seven5 executable because it understands how to generate errors in
//that case.
func NewWire(path string) *Wire {
	return &Wire{path}
}

//Dispatch is called to invoke a command.  It is called to run a stub
//then decode the results. It returns the decoded object and an error. There
//will only be a valide result value if error is nil.  It can be the case
//that the error is WIRE_SEMANTIC_ERROR meaning that the other side ran ok,
//but the return value said that there was an error.
func (self *Wire) Dispatch(cmd string, dir string, writer http.ResponseWriter,
	request *http.Request, arg interface{}, logger util.SimpleLogger) (interface{}, error) {
	var cfg *ApplicationConfig
	var req *util.BrowserRequest
	var jsonBlob string
	var err error

	start := time.Now()

	//figure out app config... don't really need this but it is better for
	//safety to check it again here
	cfg, err = decodeAppConfig(dir)

	//marshalling for requset is special, not using standard json path
	if req, err = util.MarshalRequest(request, logger); err != nil {
		return nil, err
	}

	//marshal the BrowserRequest and config to json
	cfgBlob, reqBlob, argBlob := self.marshal(cfg, req, arg, logger)

	msg := fmt.Sprintf("Configuration blob for '%s'", cmd)
	logger.DumpJson(util.DEBUG, msg, cfgBlob)
	msg = fmt.Sprintf("Request blob for '%s'", cmd)
	logger.DumpJson(util.DEBUG, msg, reqBlob)
	if argBlob!="" {
		msg = fmt.Sprintf("Arg blob for '%s'", cmd)
		logger.DumpJson(util.DEBUG, msg, argBlob)
	} else {
		logger.Debug("No argument passed to command '%s' (empty)",cmd)
	}
	logger.Debug("Working directory for '%s': %s", cmd, dir)

	defer func() {
		diff := time.Since(start)
		logger.Info("Command '%s' took %s", cmd, diff)
	}()

	if jsonBlob = self.runSeven5(cmd, dir, cfgBlob, reqBlob, argBlob, logger); jsonBlob == "" {
		return nil, WIRE_PROCESS_ERROR
	}
	//need to handle the custom results
	result := STUBS[cmd].result()
	dec := json.NewDecoder(strings.NewReader(jsonBlob))
	if err = dec.Decode(result); err != nil {
		msg = fmt.Sprintf("unable to understand json from seven5:%s", err)
		logger.DumpTerminal(util.ERROR, msg, jsonBlob)
		return nil, err
	}
	logger.DumpJson(util.DEBUG, "Json result from seven5", jsonBlob)
	if STUBS[cmd].resultHasError(result) {
		return nil, WIRE_SEMANTIC_ERROR
	}
	if STUBS[cmd].resultBody != nil {
		body := STUBS[cmd].resultBody(result)
		_, err = writer.Write([]byte(body))
		if err != nil {
			//not sure what to do here
			fmt.Fprintf(os.Stdout, "Problems with writing to HTTP output: %s\n",
				err.Error())
			return nil, err
		}
	}
	return result, nil
}

//runSeven5 is called to take the three bunches of json arguments (and a couple
//of miscellaneous pramaters) and invoke
//the process of Seven5 with those. It captures the output of the process
//and returns the marshalled result that it got from stdout.
func (self *Wire) runSeven5(name string, dir string, config string,
	browserReq string, arg string, logger util.SimpleLogger) string {
	var err error

	shellCommand := exec.Command(self.path, name, dir, logger.GetLogLevel(),
		config, browserReq, arg)
	out, err := shellCommand.CombinedOutput()
	allOutput := string(out)
	index := strings.Index(allOutput, MARKER)
	if index < 0 {
		logger.DumpTerminal(util.ERROR, "Bad output received from seven5", allOutput)
		return ""
	}
	pos := len(MARKER) + index
	logMsg := allOutput[pos:]
	marshalledText := allOutput[:index]
	if err != nil || badMarshalledText(marshalledText) {
		logger.Raw(logMsg)
		if badMarshalledText(marshalledText) {
			logger.Error("internal seven5 error running command %s (no result from command)", name)
		} else {
			logger.Error("error running command %s was %s", name, err.Error())
		}
		return ""
	}
	logger.Raw(logMsg)

	return marshalledText

}

func badMarshalledText(s string) bool {
	blob := strings.Trim(s, " \t\n")
	return blob == "" || blob == "{}"
}

//marshal is called to create the three bunches of json that we need to
//talk to seven5
func (self *Wire) marshal(cfg *ApplicationConfig, req *util.BrowserRequest,
	arg interface{}, logger util.SimpleLogger) (string, string, string) {
	var cfgBuffer bytes.Buffer
	var reqBuffer bytes.Buffer
	var argBuffer bytes.Buffer

	cfgEnc := json.NewEncoder(&cfgBuffer)
	reqEnc := json.NewEncoder(&reqBuffer)
	argEnc := json.NewEncoder(&argBuffer)

	cfgEnc.Encode(cfg)
	reqEnc.Encode(req)
	if arg == nil {
		argBuffer.WriteString("")
	} else {
		argEnc.Encode(arg)
	}
	return cfgBuffer.String(), reqBuffer.String(), argBuffer.String()
}
