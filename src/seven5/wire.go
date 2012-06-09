package seven5

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"seven5/cmd"
	"seven5/util"
	"strings"
	"time"
)

//WireStrategy is used to encode the particular marshalling/unmarshalling
//strategy used by the system.  It is only interesting if you want to
//implement new ways to talk to commands.
type WireStrategy interface {
	//Call does the work of calling the command (fn) with the arguments
	//supplied. All the params have been encoded by MarshalArgument. 
	//The first return value is log data from the invoked code.  The second
	//is the (encoded) result of running the command.
	Call(commandName string, log util.SimpleLogger, param ...interface{}) (string, interface{})
	//MarshalArgument does an encoding the particluar argument
	//supplied (arg) and returning the encoded value.
	MarshalArgument(arg *cmd.CommandArgPair, commandName string, count int,
		log util.SimpleLogger, cl cmd.ClientSideCapability) (interface{}, error)
	//Unmarshal result does decoding of the result of a command and returns
	//the result.  
	UnmarshalResult(commandName string, rez *cmd.CommandReturn, 
		log util.SimpleLogger, retValue interface{}) (interface{}, error)
}

//Wire represents all the processing necessary for handling a command execution
//that is NOT specific to the particular encoding in use.
type Wire struct {
	strategy WireStrategy
}

var (
	//WIRE_SEMANTIC_ERROR means the other side ran to completion, but returned
	//an error on the semantic processing.  The details should be in the logs.
	WIRE_SEMANTIC_ERROR = errors.New("Semantic error in seven5 command")
	//WIRE_PROCESS_ERROR means that the other side could not be contacted
	//successfully.
	WIRE_PROCESS_ERROR = errors.New("Unable to run the seven5 command")
)

//NewWire creates a new Wire instance.  Don't call unless your WireStrategy
//is fully ready.
func NewWire(strategy WireStrategy) *Wire {
	return &Wire{strategy}
}

//Dispatch is called to invoke a command.  It is called to generate arguments,
//marshall them up, invoke the server, then decode the result by unmarshalling
//the string result. It returns the decoded result object and an error. There
//will only be a valid result value if error is nil.  It can be the case
//that the error is WIRE_SEMANTIC_ERROR meaning that the other side ran ok,
//but the return value said that there was an error.
func (self *Wire) Dispatch(commandName string, writer http.ResponseWriter,
	request *http.Request, log util.SimpleLogger, clientCap cmd.ClientSideCapability) (interface{}, error) {
	start := time.Now()

	defer func() {
		diff := time.Since(start)
		log.Info("Command '%s' took %s", commandName, diff)
	}()

	//first two args arge required
	toPass := []interface{}{}

	//determine the rest of the args	
	count := 0
	command := Seven5app[commandName]
	argDefn := command.Arg

	for _, arg := range argDefn {
		var encodedArg interface{}
		var err error
		encodedArg, err = self.strategy.MarshalArgument(arg, commandName, 
			count, log, clientCap)
		//check for marshalling error
		if err != nil {
			return nil, err
		}
		toPass = append(toPass, encodedArg)
		count++
	}

	logData, encodedResult := self.strategy.Call(commandName,log, toPass...)
	if encodedResult == nil{
		return nil,WIRE_PROCESS_ERROR
	}
	
	//show log data, since we have it now
	log.Raw(logData);
	
	//decode result
	result, err := self.strategy.UnmarshalResult(commandName, command.Ret, 
		log, encodedResult)
	if err!=nil {
		return nil,err
	}
	
	if command.Ret.ErrorTest(result) {
		return nil, WIRE_SEMANTIC_ERROR
	}
		//might need to copy the body to the browser
	if command.Ret.GetBody != nil {
		size, err := writer.Write([]byte(command.Ret.GetBody(result)))
		log.Debug("Size of body written to browser was %d bytes", size)
		if err != nil {
			log.Error("Unable to write to the browser for command: %s:%s",
				commandName, err)
			return nil, err
		}
	}
	
	return result, nil
}

//runSeven5 is called to take two fixed params (log level and command name)
//plus the encoded parameter arguments and pass them through to the program.
//It captures the output of the process and returns log data and marshalled
//value.
func (self *JsonStringExecStrategy) runSeven5(log util.SimpleLogger, param ...string) (string,string) {
	var err error

	//run the shell command and collect output
	shellCommand := exec.Command(self.pathToSeven5, param...)
	out, err := shellCommand.CombinedOutput()
	allOutput := string(out)

	//look for the magic marker in the stdout string
	index := strings.Index(allOutput, MARKER)
	if index < 0 {
		log.DumpTerminal(util.ERROR, "Bad output received from seven5", allOutput)
		return "",""
	}

	//log errors that are at our level (not semantics) of process handling
	pos := len(MARKER) + index
	logMsg := allOutput[pos:]
	marshalledText := allOutput[:index]
	if err != nil || self.badMarshalledText(marshalledText) {
		log.Raw(logMsg)
		if self.badMarshalledText(marshalledText) {
			log.Error("internal seven5 error running command %s (no result from command)",
				param[1])
		} else {
			log.Error("error running command %s was %s", param[1], err.Error())
		}
		return "",""
	}

	return logMsg, marshalledText

}

//MarshalUserArgument does the work of figuring out how to take a particular
//argument and convert it to a json string suitable for sending to the 
//Seven5 application.  The commandName and count parameters are just for
//producing nicer error messages.  The arg has the necessary information about
//the parameter and the cl has functions that may be needed to create the
//_value_ of the parameter to be encoded.
func (self *JsonStringExecStrategy) MarshalArgument(arg *cmd.CommandArgPair, 
	commandName string, count int, log util.SimpleLogger, cl cmd.ClientSideCapability) (interface{}, error) {

	generated, err := arg.Generator(cl, log)
	if err != nil {
		log.Error("Error trying to generate arg %d for command %s", count, commandName)
		return "", err
	}
	encodedArg, ok := generated.(string)
	if ok && arg.Unmarshalled != nil {
		log.Error("Mismatch on argument %d of %s, string value provided but not expected",
			count, commandName)
		return "", err
	}
	if !ok && arg.Unmarshalled == nil {
		log.Error("Mismatch on argument %d of %s, encoding needed but not expected",
			count, commandName)
		return "", err
	}
	if ok {
		log.DumpTerminal(util.DEBUG, fmt.Sprintf("raw arg %d to %s", count, commandName),
			encodedArg)
	} else {
		//the object needs to be marshalled
		var jsonBuffer bytes.Buffer
		encoder := json.NewEncoder(&jsonBuffer)
		err := encoder.Encode(generated)
		if err != nil {
			log.Error("Could not encode argument %d of %s, %+v: %s",
				count, commandName, generated, err)
			return "", err
		}
		encodedArg = jsonBuffer.String()
		log.DumpJson(util.DEBUG, fmt.Sprintf("json arg %d to %s", count, commandName), encodedArg)
	}
	return encodedArg, nil
}

//badMarshalledText returns true if the result value from Seven5 as "marshalled"
//value doesn't make sense
func (self *JsonStringExecStrategy) badMarshalledText(s string) bool {
	blob := strings.Trim(s, " \t\n")
	return blob == "" || blob == "{}"
}

//JsonStringExcStrategy is a wire encoding that uses json-encoded values (in 
//strings) as command line arguments to another (child) process.  The process
//runs and it's standard output is the result, including
//both log data for the parent as well as a json encoded result.Public
//because it's used by the roadie.
type JsonStringExecStrategy struct {
	pathToSeven5 string
}

//Create a JsonStringStrategy from a path to executable... public so it can be used
//by the roadie.
func NewJsonStringStrategy(path string) WireStrategy {
	result := &JsonStringExecStrategy{path}
	return result
}

//Convenience for users of NewJsonStringStrategy... public so it can be used
//by the roadie.  BuildSeven5 does the work of building the boostrap seven5 
//instance and returns the path to it or "" and an error.
func BuildSeven5(logger util.SimpleLogger) (string, error) {
	start := time.Now()

	b := &bootstrap{logger}
	config, err := b.configureSeven5("")
	if err != nil {
		return "", err
	}
	result, err := b.takeSeven5Pill(config)
	if err != nil {
		return "", err
	}

	delta := time.Since(start)
	logger.Info("Rebuilding seven5 took %s", delta.String())
	return result, nil
}

//Call does the work of calling the command  with the arguments
//supplied. All the params have been encoded by MarshalArgument. 
//The first return value is log data from the invoked code.  The second
//is the (encoded) result of running the command.  If something went
//wrong, the 2nd result will be nil.
func (self *JsonStringExecStrategy) Call(commandName string,
	log util.SimpleLogger, param ...interface{}) (string, interface{}) {
	
	//first two args args are required to tell the other side the log level
	//and the command name
	toPass := []string{}
	toPass = append(toPass, log.GetLogLevel())
	toPass = append(toPass, commandName)
	log.DumpTerminal(util.DEBUG,
		fmt.Sprintf("fixed args (0=log level, 1=command name) to %s", commandName),
		log.GetLogLevel()+"\n"+commandName)

	for _, p := range param {
		toPass = append(toPass, p.(string))
	}
	//
	// Execution of cammand
	//
	var jsonBlob string
	var logData string
	if logData, jsonBlob = self.runSeven5(log, toPass...); jsonBlob == "" {
		return "", nil
	}
	return logData, jsonBlob		
}

//Unmarshal result does decoding of the result of a command and returns
//the result.  
func (self *JsonStringExecStrategy) UnmarshalResult(commandName string, 
	retDefn *cmd.CommandReturn, log util.SimpleLogger, value interface{}) (interface{}, error) {

	jsonBlob := value.(string)

	result := retDefn.Unmarshalled()
	decoder := json.NewDecoder(strings.NewReader(jsonBlob))
	if err := decoder.Decode(result); err != nil {
		log.Error("Unable to decode result of %s, %+v: %s", commandName, result, err)
		return nil, err
	}
	
	return result, nil
}
