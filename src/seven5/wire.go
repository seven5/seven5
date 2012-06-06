package seven5

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"seven5/util"
	"strings"
	"time"
	"seven5/cmd"
)

//WireStrategy is used to encode the particular marshalling/unmarshalling
//strategy used by the system.  It is only interesting if you want to
//implement new ways to talk to commands.
type WireStrategy interface {
	//Call does the work of calling the command (fn) with the arguments
	//supplied. All the params have been encoded by MarshalArgument. 
	//The first return value is log data from the invoked code.  The second
	//is the (encoded) result of running the command.
	Call(fn func(util.SimpleLogger,...interface{}) interface{},
		log util.SimpleLogger, param ... interface{}) (string, interface{})
	//MarshalArgument does an encoding the particluar argument
	//supplied (arg) and returning the encoded value.
	MarshalArgument(arg *cmd.CommandArgPair, commandName string, count int, 
		log util.SimpleLogger, cl cmd.ClientSideCapability) (interface{}, error)
	//Unmarshal result does decoding of the result of a command and returns
	//the result.  
	UnmarshalResult(rez *cmd.CommandReturn, log util.SimpleLogger) (interface{}, error)
}
	
//Wire represents the way to talk to a seven5 program.  It is, in some sense,
//the client side of the RPC system. It handles marshalling
//calling commands on the other side (invoking seven5 as a program), and 
//unmarshalling the result. It holds a reference to the current Seven5 executable.
type Wire struct {
	path string
}

var (
	//WIRE_SEMANTIC_ERROR means the other side ran to completion, but returned
	//an error on the semantic processing.  The details should be in the logs.
	WIRE_SEMANTIC_ERROR = errors.New("Semantic error in seven5 command")
	//WIRE_PROCESS_ERROR means that the other side could not be contacted
	//successfully.
	WIRE_PROCESS_ERROR  = errors.New("Unable to run the seven5 command")
)

//NewWire creates a new Wire instance.  Don't call this unless you have
//a viable executable to run as the first param.
func NewWire(path string) *Wire {
	return &Wire{path}
}

//Dispatch is called to invoke a command.  It is called to generate arguments,
//marshall them up, invoke the server, then decode the result by unmarshalling
//the string result. It returns the decoded result object and an error. There
//will only be a valid result value if error is nil.  It can be the case
//that the error is WIRE_SEMANTIC_ERROR meaning that the other side ran ok,
//but the return value said that there was an error.
func (self *Wire) Dispatch(commandName string, dir string, writer http.ResponseWriter,
	request *http.Request, log util.SimpleLogger) (interface{}, error) {
	start := time.Now()

	defer func() {
		diff := time.Since(start)
		log.Info("Command '%s' took %s", commandName, diff)
	}()

	//first two args arge required
	toPass:=[]string{}
	toPass = append(toPass,log.GetLogLevel())
	toPass = append(toPass,commandName)
	log.DumpTerminal(util.DEBUG, 
		fmt.Sprintf("fixed args (0=log level, 1=command name) to %s", commandName),
		log.GetLogLevel()+"\n"+commandName);

	
	//determine the rest of the args	
	count := 2;
	command:=Seven5app[commandName]
	argDefn := command.Arg
	cap := cmd.NewDefaultClientCapability(request)
	
	for _,arg:=range argDefn {
		encodedArg := ""
		var err error
		encodedArg, err = marshalUserArgument(arg, commandName, count, log, cap)
		//check for marshalling error
		if err!=nil {
			return nil, err
		}
		toPass = append(toPass, encodedArg)
		count++
	}
	
	
	//
	// Execution of cammand
	//
	var jsonBlob string
	if jsonBlob = self.runSeven5(log, toPass...); jsonBlob == "" {
		return nil, WIRE_PROCESS_ERROR
	}
	
	//
	// Decode result
	//
	retDefn := command.Ret
	result := retDefn.Unmarshalled()
	decoder:=json.NewDecoder(strings.NewReader(jsonBlob))
	if err:=decoder.Decode(result); err!=nil {
		log.Error("Unable to decode result of %s, %+v: %s", commandName, result, err)
		return nil, err
	}
	if retDefn.ErrorTest(result) {
		return nil, WIRE_SEMANTIC_ERROR
	}
	//might need to copy the body to the browser
	if retDefn.GetBody!=nil {
		size, err := writer.Write([]byte(retDefn.GetBody(result)))
		log.Debug("Size of body written to browser was %d bytes",size)
		if err!=nil {
			log.Error("Unable to write to the browser for command: %s:%s",
				commandName, err)
			return nil, err
		}
	}
		
	return result, nil
}

//runSeven5 is called to take the three bunches of json arguments (and a couple
//of miscellaneous pramaters) and invoke
//the process of Seven5 with those. It captures the output of the process
//and returns the marshalled result that it got from stdout.
func (self *Wire) runSeven5(log util.SimpleLogger, param ... string) string {
	var err error

	//run the shell command and collect output
	shellCommand := exec.Command(self.path, param...)
	out, err := shellCommand.CombinedOutput()
	allOutput := string(out)
	
	//look for the magic marker in the stdout string
	index := strings.Index(allOutput, MARKER)
	if index < 0 {
		log.DumpTerminal(util.ERROR, "Bad output received from seven5", allOutput)
		return ""
	}
	
	//log errors that are at our level (not semantics) of process handling
	pos := len(MARKER) + index
	logMsg := allOutput[pos:]
	marshalledText := allOutput[:index]
	if err != nil || badMarshalledText(marshalledText) {
		log.Raw(logMsg)
		if badMarshalledText(marshalledText) {
			log.Error("internal seven5 error running command %s (no result from command)", 
				param[1])
		} else {
			log.Error("error running command %s was %s", param[1], err.Error())
		}
		return ""
	}
	
	//copy the log value from the other side into our log so the user can see it
	log.Raw(logMsg)
	return marshalledText

}

//marshalUserArgument does the work of figuring out how to take a particular
//argument and convert it to a json string suitable for sending to the 
//Seven5 application.  The commandName and count parameters are just for
//producing nicer error messages.  The arg has the necessary information about
//the parameter and the cl has functions that may be needed to create the
//_value_ of the parameter to be encoded.
func marshalUserArgument(arg *cmd.CommandArgPair, commandName string, count int, 
	log util.SimpleLogger, cl cmd.ClientSideCapability) (string, error){
		
	generated, err:= arg.Generator(cl,log)
	if err!=nil {
		log.Error("Error trying to generate arg %d for command %s",count,commandName)
		return "", err
	}
	encodedArg, ok := generated.(string)
	if ok && arg.Unmarshalled!=nil {
		log.Error("Mismatch on argument %d of %s, string value provided but not expected",
			count,commandName)
		return "", err
	}
	if !ok && arg.Unmarshalled==nil {
		log.Error("Mismatch on argument %d of %s, encoding needed but not expected",
			count,commandName)
		return "", err
	}
	if ok {
		log.DumpTerminal(util.DEBUG, fmt.Sprintf("raw arg %d to %s", count, commandName), 
			encodedArg)
	} else {
		//the object needs to be marshalled
		var jsonBuffer bytes.Buffer
		encoder:=json.NewEncoder(&jsonBuffer)
		err:=encoder.Encode(generated)
		if err!=nil {
			log.Error("Could not encode argument %d of %s, %+v: %s",
				count, commandName, generated, err)
			return "", err
		}
		encodedArg = jsonBuffer.String()
		log.DumpJson(util.DEBUG, fmt.Sprintf("json arg %d to %s", count, commandName), encodedArg)
	}
	return encodedArg,nil
}

//badMarshalledText returns true if the result value from Seven5 as "marshalled"
//value doesn't make sense
func badMarshalledText(s string) bool {
	blob := strings.Trim(s, " \t\n")
	return blob == "" || blob == "{}"
}
