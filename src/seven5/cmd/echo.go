package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	"seven5/util"
)

//
// Echo is used to display back to the browser what the command received.  This
// can be useful for either debugging or for understanding how to implement a command
// that has browser output.
//
var Echo = &CommandDecl{
	Arg: []*CommandArgPair{
		ClientSideRequest, //need the request so we can do output
	},
	Ret: BuiltinReturnWithBody,
	Impl: defaultEcho,
}

func defaultEcho(log util.SimpleLogger, v...interface{}) interface{} {
	log.Info("this is a log message from the echo command")

	request:=v[0].(*http.Request)
	
	var output bytes.Buffer
	result:=&BodyReturn{}
	
	output.WriteString("<H1>Echo To You</H1>")
	output.WriteString("<H3>Headers</H3>")
	for i, j := range request.Header {
		for _, k := range j {
			output.WriteString(fmt.Sprintf("<span>%s:%s</span><br/>", i, k))
		}
	}
	output.WriteString("<H3>Cookies</H3>")
	for _, cookie := range request.Cookies() {
		c := fmt.Sprintf("<span>%s,%s,%s,%s</span><br/>", cookie.Name,
			cookie.Expires.String(), cookie.Domain, cookie.Path)
		output.WriteString(c)
	}
	output.WriteString("<h3>Big Stuff</h3>")
	output.WriteString(fmt.Sprintf("<span>%s %s</span><br/>",
		request.Method, request.URL))
	values := request.URL.Query()
	if len(values) > 0 {
		output.WriteString("<h4>Query Params</h4>")
		for k, l := range values {
			for _, v := range l {
				output.WriteString(fmt.Sprintf("<span>%s:%s</span><br/>", k, v))
			}
		}
	}

	result.Body = output.String()
	result.Error = false
	return &result
}
