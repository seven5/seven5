package plugin

import (
	"bytes"
	"encoding/json"
	"seven5/util"
	"strings"
	"time"
	"fmt"
)

const MARKER = "@@@+++@@@"

//Results is the shared portion of all results coming back from Seven5.
type CommandResult struct {
	Error          bool
	Panic          bool
	TipMsg         string
	ProcessingTime time.Duration
}

func ErrorResult() CommandResult {
	return CommandResult{Error: true}
}

// Command is sent to Seven5 first, so it knows how to parse the rest.
// Must be public for json encoding.
type Command struct {
	Name         string
	AppDirectory string
}

//Run is the equivalent of main for Seven5 when in development mode.  
//The real main uses a pill. Input should be two json strings and the output 
//the same.
func RunCommand(command string, arg string, reqJson string) (ret string) {
	var cmd Command
	var result bytes.Buffer
	var logdata bytes.Buffer
	
	
	logger := util.NewHtmlLogger(util.DEBUG, true, &logdata, false)
	
	req:=util.UnmarshalRequest(reqJson,logger)
	
	decoder := json.NewDecoder(strings.NewReader(command))
	encoder := json.NewEncoder(&result)
	decoder.Decode(&cmd)

	start := time.Now()

	defer func() {
		if rec := recover(); rec != nil {
			var r CommandResult
			r.Error = true
			r.Panic = true
			r.ProcessingTime = time.Since(start)
			var b bytes.Buffer 
			b.WriteString("Trapped a panic in command processing:\n")
			for _,i := range ([]int{4,5,6,7}) {
				file, line:=util.GetCallerAndLine(i)
				b.WriteString(fmt.Sprintf("%s:%d\n",file,line))
			}
			logger.DumpTerminal(b.String())
			encoder.Encode(&result)
			ret = result.String() + MARKER + logdata.String()
		}
	}()

	switch cmd.Name {
	case VALIDATE_PROJECT:
		var pvArgs ValidateProjectArgs
		decoder = json.NewDecoder(strings.NewReader(arg))
		decoder.Decode(&pvArgs)
		r := Seven5App.Validator.Validate(&cmd, &pvArgs, logger)
		r.CommandResult.ProcessingTime = time.Since(start)
		encoder.Encode(&r.CommandResult)
		break
	case ECHO:
		var echoArgs EchoArgs
		decoder = json.NewDecoder(strings.NewReader(arg))
		decoder.Decode(&echoArgs)
		r := Seven5App.Echo.Echo(&cmd, &echoArgs, req, logger)
		r.CommandResult.ProcessingTime = time.Since(start)
		encoder.Encode(&r)
		break
	default:
		var myRes CommandResult
		myRes.Error = true
		myRes.ProcessingTime = time.Since(start)
		logger.Error("unknown command to seven5:%s", cmd.Name)
		encoder.Encode(&myRes)
	}
	ret=result.String() + MARKER + logdata.String()
	return
}

