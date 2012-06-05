package seven5

import (
	"bytes"
)

const MARKER = "@@@+++@@@"

//PanicResult is returned only when there has been a trapped panic.  This should
//be rarely or never.
type PanicResult struct {
	Error bool
	Panic bool
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
