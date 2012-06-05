package seven5

import (
	"bytes"
	"encoding/json"
	"fmt"
	"seven5/util"
	"strings"
	"seven5/cmd"
)

//all commands
const (
	VALIDATEPROJECT      = "ValidateProject"
	ECHO                 = "Echo"
	PROCESSCONTROLLER    = "ProcessController"
	PROCESSVOCAB         = "ProcessVocab"
	BUILDUSERLIB         = "BuildUserLib"
	EXPLODETYPE          = "ExplodeType"
	DESTROYGENERATEDFILE = "DestroyGeneratedFile"
)

//seven5app this is the "application" that is seven5.
var Seven5app = make(map[string]*cmd.CommandDecl)

//RunCommand is the equivalent of main for Seven5 when in development mode.  
//The real main uses a pill. The command name is the name the other side of
//the wire supplied as our name. 
func RunCommand(logLevel string, commandName string, param ... string) (ret string) {
	var resultBuffer bytes.Buffer
	var logdataBuffer bytes.Buffer

	log := util.NewHtmlLogger(util.LogLevelStringToLevel(logLevel), &logdataBuffer, false)
	command:=Seven5app[commandName]

	//
	// DECODE PARAMS
	//
	toPass:=[]interface{}{}
	count:=0
	for _,argDefn:=range command.Arg {
		raw := param[count]
		count++
		if argDefn.Unmarshalled == nil {
			toPass=append(toPass,raw/*the implementations should expect a string*/)
			continue
		}
		enc:=json.NewDecoder(strings.NewReader(raw))
		arg:=argDefn.Unmarshalled();
		if err:=enc.Decode(arg); err!=nil {
			log.DumpJson(util.ERROR, "Raw parameter with problem",raw);
			log.Error("Unable to decode into %+v inside seven5",arg)
			ret = createResultString(nil, logdataBuffer)
			return 
		}
		toPass = append(toPass,arg)
	}
	//we have now encoded all the args into toPass

	//prep encoder
	encoder := json.NewEncoder(&resultBuffer)

	//in case there is a panic, we want to have some hope of seeing what
	//happened in this address space
	defer func() {
		if rec := recover(); rec != nil {
			var b bytes.Buffer
			for _, i := range []int{4, 5, 6, 7} {
				file, line := util.GetCallerAndLine(i)
				b.WriteString(fmt.Sprintf("%s:%d\n", file, line))
			}
			log.DumpTerminal(util.ERROR, "Trapped 'Panic' processing command",
				b.String())
			ret = createResultString(nil, logdataBuffer)
		}
	}()

	//invoke command and run it
	rawResult := command.Impl(log,toPass...)
	if err:=encoder.Encode(rawResult); err!=nil {
		log.Error("Unable to encode %+v inside seven5",rawResult)
		ret = createResultString(nil, logdataBuffer)
		return 
	}
	//everything encoded ok
	ret = createResultString(&resultBuffer, logdataBuffer)
	return
}
