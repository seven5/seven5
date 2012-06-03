package seven5

import (
	"bytes"
	"encoding/json"
	"fmt"
	"seven5/util"
	"strings"
)

//all commands
const (
	VALIDATEPROJECT   = "ValidateProject"
	ECHO              = "Echo"
	PROCESSCONTROLLER = "ProcessController"
	PROCESSVOCAB      = "ProcessVocab"
	BUILDUSERLIB      = "BuildUserLib"
	EXPLODETYPE       = "ExplodeType"
)

//seven5app this is the "application" that is seven5.
var Seven5app = make(map[string]Command)

//RunCommand is the equivalent of main for Seven5 when in development mode.  
//The real main uses a pill. The command name is the name the other side of
//the wire supplied as our name.  The dir is the application's root dir.
//The logLevel is the current roadie log level as a string.  The configJson
//is a blob that represents a *seven5.ApplicationConfig.  The reqJson is
//represents a *util.BrowserRequest which is later converted a real 
//*http.Request.  The last argument are specific arguments for this command,
//and may be empty ("").
func RunCommand(commandName string, dir string, logLevel string, configJson string,
	reqJson string, argJson string) (ret string) {
	var config ApplicationConfig
	var resultBuffer bytes.Buffer
	var logdataBuffer bytes.Buffer
	
	logger := util.NewHtmlLogger(util.LogLevelStringToLevel(logLevel), &logdataBuffer, false)

	//requests have to be treated specilaly, not using the "normal" path
	//of json decoding
	req, err := util.UnmarshalRequest(reqJson, logger)
	if err != nil {
		logger.Error("Error decoding request structure inside seven5!")
		ret = createResultString(nil, logdataBuffer)
		return
	}

	//decode the app config
	decoder := json.NewDecoder(strings.NewReader(configJson))
	err = decoder.Decode(&config)
	if err != nil {
		logger.Error("Error decoding application config inside seven5!")
		ret = createResultString(nil, logdataBuffer)
		return
	}
	
	//arg?
	arg:=Seven5app[commandName].GetArg()
	if argJson=="" {
		if arg!=nil {
			logger.Error("Stubs and implementation out of sync for command '%s': "+
				"no arg supplied, but one expected!",commandName)
			ret = createResultString(nil, logdataBuffer)
			return
		}
	} else {
		argDec := json.NewDecoder(strings.NewReader(argJson))
		if arg==nil {
			logger.Error("Stubs and implementation out of sync for command '%s': "+
				"arg supplied but it was not expected!",commandName)
			ret = createResultString(nil, logdataBuffer)
			return
		}
		err = argDec.Decode(arg)
		if err!=nil {
			logger.Error("Error decoding the argument for command '%s': %s",
				commandName, err)
			ret = createResultString(nil, logdataBuffer)
			return
		}
	}

	//prep encoder
	encoder := json.NewEncoder(&resultBuffer)

	defer func() {
		if rec := recover(); rec != nil {
			var b bytes.Buffer
			for _, i := range []int{4, 5, 6, 7} {
				file, line := util.GetCallerAndLine(i)
				b.WriteString(fmt.Sprintf("%s:%d\n", file, line))
			}
			logger.DumpTerminal(util.ERROR, "Trapped 'Panic' processing command",
				b.String())
			ret = createResultString(nil, logdataBuffer)
		}
	}()

	cmd, ok := Seven5app[commandName]
	if ok {
		resultStruct := cmd.Exec(commandName, dir, &config, req, arg, logger)
		err := encoder.Encode(resultStruct)
		if err != nil {
			logger.Error("Error encoding result: %s", err)
			ret = createResultString(nil, logdataBuffer)
			return
		}
	} else {
		logger.Error("unknown command to seven5:'%s'", commandName)
		ret = createResultString(nil, logdataBuffer)
		return
	}
	logger.Debug("command '%s' ran to completion, size of marshalled result %d bytes",
		commandName, resultBuffer.Len())
	ret = createResultString(&resultBuffer, logdataBuffer)
	return
}
