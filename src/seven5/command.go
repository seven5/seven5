package seven5

import (
	"bytes"
	"net/http"
	"seven5/util"
)

const MARKER = "@@@+++@@@"

//PanicResult is returned only when there has been a trapped panic.  This should
//be rarely or never.
type PanicResult struct {
	Error bool
	Panic bool
}

//Command represents the way to talk to command inside seven5.  This interface
//is used after json unmarshalling is completed in RunCommand().
type Command interface {
	//Exec is called to run the command inside seven5.  All the
	//parameters will have be unmarshallled and the return value
	//will be marshalled for transmission back to roadie.
	Exec(name string, cwd string,
		config *ApplicationConfig, request *http.Request, arg interface{},
		log util.SimpleLogger) interface{}
	//Return nil if you don't need any extra arg love
	GetArg() interface{}
}

//createResultString is used for putting together the STDOUT for a comand.
//it constructs this from the result (which is a string of json) a marker and
//the log buffer (which is shoved into the browser by the roadie without
//mods)
func createResultString(resultBuffer *bytes.Buffer, logDataBuffer bytes.Buffer) string {
	logData := logDataBuffer.String()
	result := ""
	if resultBuffer != nil {
		result = resultBuffer.String()
	}
	return result + MARKER + logData

}
