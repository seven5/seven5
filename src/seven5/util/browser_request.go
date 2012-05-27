package util

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
)

//BrowserRequest is a more "json flattening" format for a request (compared
//to http.Request).  This is the intermediary between the roadie process
//and the seven5 app.
type BrowserRequest struct {
	Header    map[string][]string
	Cookie    []http.Cookie
	Url       string
	Body      string
	Method    string
	UserAgent string
}

//UnmarshalRequest converts a blob of json into an http.Request via
//our BrowserRequet intermediate type
func UnmarshalRequest(reqJson string, logger SimpleLogger) *http.Request {
	var browserReq BrowserRequest
	decoder := json.NewDecoder(strings.NewReader(reqJson))
	decoder.Decode(&browserReq)
	result, err := http.NewRequest(browserReq.Method, browserReq.Url,
		strings.NewReader(browserReq.Body))
	if err != nil {
		logger.Panic("Can't create request!")
	}
	for k, l := range browserReq.Header {
		for _, v := range l {
			result.Header.Add(k, v)
		}
	}

	return result
}

//MarshalRequest is the routine that converts a "Real" http requset into
//something more suitable for our use over the json connection.
func MarshalRequest(request *http.Request, logger SimpleLogger) BrowserRequest {
	var result BrowserRequest
	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(request.Body); err != nil {
		logger.Panic("could not read contents of HTTP requset body:%s",
			err.Error())
	}
	result.Header = make(map[string][]string)
	result.Cookie = []http.Cookie{}

	result.Header = request.Header
	result.Body = buffer.String()
	result.Url = request.URL.String()
	result.Method = request.Method
	result.UserAgent = request.UserAgent()
	//for _, c := range request.Cookies() {
	//	result.Cookie[c.Name] = c.Unparsed
	//}
	return result
}
