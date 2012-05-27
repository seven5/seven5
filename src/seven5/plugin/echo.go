package plugin

import (
	"seven5/util"
	"net/http"
	"bytes"
	"fmt"
)



//EchoArgs is passed to the Echo plugin to do its
//work.Must be public for json encoding.
type EchoArgs struct {
	util.BrowserRequest
}

//EchoResult is the result type of a call on the Echo plugin.
//Must be public for json encoding.
type EchoResult struct {
	CommandResult
	Response http.Response
	Body string
}

// ProjectValidator checks to see if the layout of the project is
// acceptable for future phases.
type Echo interface {
	Echo(cmd *Command, args *EchoArgs, request *http.Request, log util.SimpleLogger) *EchoResult
}

// Default echo plugin just prints unformatted version of what you sent
type DefaultEcho struct {
}

func (self *DefaultEcho) Echo(cmd *Command, args *EchoArgs, 
	request *http.Request, log util.SimpleLogger) *EchoResult {
	
	log.Info("this is a log message from the echo groupie");
	
	result:= EchoResult{}
	var body bytes.Buffer
	
	body.WriteString("<H1>Echo To You</H1>")
	for i,j := range(request.Header){
		for _,k:=range(j) {
			body.WriteString(fmt.Sprintf("<span>%s:%s</span><br/>",i,k))
		}
	}
	result.Body = body.String()
	result.CommandResult.Error = false
	return &result
}


