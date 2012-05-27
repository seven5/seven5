package plugin

import (
	"bytes"
	"encoding/json"
	"seven5/util"
	"strings"
	"time"
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
func RunCommand(command string, arg string) (ret string) {
	var cmd Command
	var result bytes.Buffer
	var logdata bytes.Buffer
	logger := util.NewHtmlLogger(util.DEBUG, true, &logdata)

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
			logger.Error("Panic was: %s", rec)
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
		encoder.Encode(&r)
	default:
		var result CommandResult
		result.Error = true
		result.ProcessingTime = time.Since(start)
		logger.Error("unknown command to seven5:%s", cmd.Name)
		encoder.Encode(&result)
	}
	return result.String() + MARKER + logdata.String()
}
