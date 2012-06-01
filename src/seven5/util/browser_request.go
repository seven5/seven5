package util

import (
	"bytes"
	"encoding/json"
	"net/url"
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
func UnmarshalRequest(reqJson string, logger SimpleLogger) (*http.Request, error) {
	var browserReq BrowserRequest
	decoder := json.NewDecoder(strings.NewReader(reqJson))
	decoder.Decode(&browserReq)
	result, err := http.NewRequest(browserReq.Method, browserReq.Url,
		strings.NewReader(browserReq.Body))
	if err != nil {
		return nil,err
	}
	for k, l := range browserReq.Header {
		for _, v := range l {
			result.Header.Add(k, v)
		}
	}
	for _, i := range browserReq.Cookie {
		result.AddCookie(&i)
	}
	result.URL, err = url.Parse(browserReq.Url)
	if err!=nil {
		return nil, err
	}
	return result, nil
}

//MarshalRequest is the routine that converts a "Real" http requset into
//something more suitable for our use over the json connection.
func MarshalRequest(request *http.Request, logger SimpleLogger) (*BrowserRequest, error) {
	var result BrowserRequest
	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(request.Body); err != nil {
		logger.Error("could not read contents of HTTP requset body:%s",
			err.Error())
		return nil, err
	}
	result.Header = make(map[string][]string)
	result.Cookie = []http.Cookie{}

	result.Header = request.Header
	result.Body = buffer.String()
	result.Url = request.URL.String()
	result.Method = request.Method
	result.UserAgent = request.UserAgent()
	
	clist := NewBetterList()
	for _, c := range(request.Cookies()) {
		clist.PushBack(c)
	}
	count:=0
	for e := clist.Front(); e != nil; e = e.Next() {
		count++
	}
	result.Cookie = make([]http.Cookie,count,count)
	count=0
	for e := clist.Front(); e != nil; e = e.Next() {
		var tmp *http.Cookie
		tmp= e.Value.(*http.Cookie)
		result.Cookie[count]=*tmp
		count++
	}
	return &result, nil
}
