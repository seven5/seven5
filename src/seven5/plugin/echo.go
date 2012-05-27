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
	body.WriteString("<H3>Headers</H3>")
	for i,j := range(request.Header){
		for _,k:=range(j) {
			body.WriteString(fmt.Sprintf("<span>%s:%s</span><br/>",i,k))
		}
	}
	body.WriteString("<H3>Cookies</H3>")
	for _,cookie := range(request.Cookies()) {
		c:=fmt.Sprintf("<span>%s,%s,%s,%s</span><br/>",cookie.Name,
			cookie.Expires.String(),cookie.Domain,cookie.Path)
		body.WriteString(c)
	}
	body.WriteString("<h3>Big Stuff</h3>")
	body.WriteString(fmt.Sprintf("<span>%s %s</span><br/>",
		request.Method,request.URL))
	values:=request.URL.Query()
	if len(values)>0 {
		body.WriteString("<h4>Query Params</h4>")
		for k,l:=range values {
			for _,v:=range l {
				body.WriteString(fmt.Sprintf("<span>%s:%s</span><br/>",k,v))
			}
		}
	}

	result.Body = body.String()
	result.CommandResult.Error = false
	return &result
}


