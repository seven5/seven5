package seven5

import (
	"encoding/json"
	_ "fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

/*---------------------------------------------------------------------------------------*/

type someResource struct {
}

func (self *someResource) Index(p PBundle) (interface{}, error) {
	return []*someWire{&someWire{Id(1074), "index"}}, nil
}
func (self *someResource) Find(id Id, p PBundle) (interface{}, error) {
	return &someWire{id, "find"}, nil
}
func (self *someResource) Post(i interface{}, p PBundle) (interface{}, error) {
	s := i.(*someWire)
	return &someWire{Id(999), s.Foo}, nil
}
func (self *someResource) Delete(id Id, p PBundle) (interface{}, error) {
	return &someWire{Id(id), "delete!"}, nil
}
func (self *someResource) Put(id Id, i interface{}, p PBundle) (interface{}, error) {
	s := i.(*someWire)
	return &someWire{Id(id), s.Foo + "?"}, nil
}

type someWire struct {
	Id  Id
	Foo String255
}

/*---------------------------------------------------------------------------------------*/

type otherWire struct {
	Id Id
	Grik String255 
}

type otherResource struct {
}

func (self *someResource) Index(p PBundle) (interface{}, error) {
	return []*otherWire{
		&otherWire{Id: Id(0), Grik:"Grak"}, 
		&otherWire{Id: Id(42), Grik:"Frak"}, 
	}, nil
}
func (self *someResource) Find(id Id, p PBundle) (interface{}, error) {
	return &someWire{id, "get off my land!"}, nil
}

/*---------------------------------------------------------------------------------------*/

func setupMux(f RestAll) *ServeMux {
	io:=NewRawIOHook(&JsonDecoder{},&JsonEncoder{}, nil)
	raw := NewRawDispatcher(io, nil, nil, NewSimpleTypeHolder(), "/rest")

	raw.Rez(&someWire{}, f)

	mux := NewServeMux()
	//note this prefix ends up _on_ all resources
	mux.Dispatch("/rest/", raw)
	return mux
}

func TestResourceMethods(t *testing.T) {
	resource := &someResource{}
	mux := setupMux(resource)
	go func() {
		http.ListenAndServe(":8189", mux)
	}()

	client := new(http.Client)

	w := makeRequestAndCheckStatus(t, client, "GET", "http://localhost:8189/rest/somewire/", "",
		http.StatusOK, true)
	checkBody(t, w, Id(1074), "index")

	body := "{ \"Id\":-1, \"Foo\":\"grik\"}"
	w = makeRequestAndCheckStatus(t, client, "POST", "http://localhost:8189/rest/somewire", body,
		http.StatusCreated, false)
	checkBody(t, w, Id(999), "grik")

	w = makeRequestAndCheckStatus(t, client, "GET", "http://localhost:8189/rest/somewire/2989/?foo=bar", "",
		http.StatusOK, false)
	checkBody(t, w, Id(2989), "find")

	body = "{ \"Id\":214, \"Foo\":\"grak\"}"
	w = makeRequestAndCheckStatus(t, client, "PUT", "http://localhost:8189/rest/somewire/214", body,
		http.StatusOK, false)
	checkBody(t, w, Id(214), "grak?")

	w = makeRequestAndCheckStatus(t, client, "DELETE", "http://localhost:8189/rest/somewire/76199", "",
		http.StatusOK, false)
	checkBody(t, w, Id(76199), "delete!")
}

func makeRequestAndCheckStatus(t *testing.T, client *http.Client, method string, url string,
	body string, status int, isSlice bool) *someWire {
	req := makeReq(t, method, url, body)
	resp, err := client.Do(req)
	checkHttpStatus(t, resp, err, status)
	return readBody(t, resp.Body, isSlice)
}

func readBody(t *testing.T, r io.ReadCloser, isSlice bool) *someWire {
	all, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatalf("failed to read the body: %s", err)
	}
	body := string(all)
	dec := json.NewDecoder(strings.NewReader(body))
	var returned someWire
	if isSlice {
		var result []someWire
		err = dec.Decode(&result)
		if err != nil {
			t.Fatalf("failed to decode the body: %s", err)
		}
		if len(result) != 1 {
			t.Fatalf("Only expected 1 item in slice but got %d\n", len(result))
		}
		returned = result[0]
	} else {
		err = dec.Decode(&returned)
		if err != nil {
			t.Fatalf("failed to decode the body: %s", err)
		}
	}
	return &returned
}

func checkBody(t *testing.T, returned *someWire, id Id, s string) {
	if returned.Id != id {
		t.Errorf("Expected id %d but got %d\n", id, returned.Id)
	}
	if string(returned.Foo) != s {
		t.Errorf("Expected to see string '%s' but got '%s'\n", s, returned.Foo)
	}
}

type badlyWrittenResource struct {
}

func (self *badlyWrittenResource) Find(id Id, p PBundle) (interface{}, error) {
	s := "foo"
	return &s, nil
}

func TestBadResource(t *testing.T) {

	bad := &badlyWrittenResource{}

	io:=NewRawIOHook(&JsonDecoder{},&JsonEncoder{}, nil)
	raw := NewRawDispatcher(io, nil, nil, NewSimpleTypeHolder(),"/rest")
	raw.ResourceSeparate("BadCoder", &someWire{}, nil, bad, nil, nil, nil)

	mux := NewServeMux()
	//note this prefix ends up _on_ all resources
	mux.Dispatch("/rest/", raw)

	go func() {
		http.ListenAndServe(":8186", mux)
	}()
	resp, err := http.Get("http://localhost:8186/rest/BadCoder/88")
	checkHttpStatus(t, resp, err, http.StatusNotFound)

	resp, err = http.Get("http://localhost:8186/rest/badcoder/88")
	checkHttpStatus(t, resp, err, http.StatusExpectationFailed)
}

func TestBadJson(t *testing.T) {

	mux := setupMux(nil)
	go func() {
		http.ListenAndServe(":8187", mux)
	}()
	body := "{ \"Id\":1, \"Foo\":\"bar\""
	resp, err := http.Post("http://localhost:8187/rest/somewire", "text/json", strings.NewReader(body))
	checkHttpStatus(t, resp, err, http.StatusBadRequest)

	//if you try to send a really big bundle, the go level code disconnects you, so we send a bunch bunch
	//of nothing to see what happens
	x := make([]byte, 100000)
	resp, err = http.Post("http://localhost:8187/rest/somewire", "text/json", strings.NewReader(string(x)))
	checkHttpStatus(t, resp, err, http.StatusBadRequest)

	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read the body: %s", err)
	}
	body = string(all)
	if strings.Index(body, "too large") == -1 {
		t.Errorf("expected to get an error about data being too large but got %s", body)
	}
}

func TestResourceNotImplementedMethods(t *testing.T) {

	mux := setupMux(nil)
	go func() {
		http.ListenAndServe(":8188", mux)
	}()

	resp, err := http.Get("http://localhost:8188/rest/somewire")
	checkHttpStatus(t, resp, err, http.StatusNotImplemented)

	body := "{}"
	resp, err = http.Post("http://localhost:8188/rest/somewire", "text/json", strings.NewReader(body))
	checkHttpStatus(t, resp, err, http.StatusNotImplemented)

	data := url.Values(map[string][]string{"nothing": []string{"bogus"}})
	resp, err = http.PostForm("http://localhost:8188/rest/somewire", data)
	checkHttpStatus(t, resp, err, http.StatusNotImplemented)

	resp, err = http.Post("http://localhost:8188/rest/somewire/2", "text/json", strings.NewReader(body))
	checkHttpStatus(t, resp, err, http.StatusBadRequest)

	resp, err = http.Get("http://localhost:8188/rest/somewire/3")
	checkHttpStatus(t, resp, err, http.StatusNotImplemented)

	client := new(http.Client)

	req := makeReq(t, "PUT", "http://localhost:8188/rest/somewire/4", "{}")
	resp, err = client.Do(req)
	checkHttpStatus(t, resp, err, http.StatusNotImplemented)

	req = makeReq(t, "DELETE", "http://localhost:8188/rest/somewire/5", "")
	resp, err = client.Do(req)
	checkHttpStatus(t, resp, err, http.StatusNotImplemented)

}

func makeReq(t *testing.T, method string, url string, body string) *http.Request {
	var result *http.Request
	var err error
	if body != "" {
		result, err = http.NewRequest(method, url, strings.NewReader(body))
	} else {
		result, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		t.Fatalf("could not create %n request: %s", method, err)
	}
	if body != "" {
		result.Header.Add("Content-Type", "text/json")
	}
	return result
}

func TestSubResource(t *testing.T) {

	mux := setupMux(nil)
	go func() {
		http.ListenAndServe(":8139", mux)
	}()
	
	
	

	w := makeRequestAndCheckStatus(t, client, "GET", "http://localhost:8189/rest/somewire/", "",
		http.StatusOK, true)
			w := makeRequestAndCheckStatus(t, client, "GET", "http://localhost:8189/rest/somewire/", "",
				http.StatusOK, true)
			
