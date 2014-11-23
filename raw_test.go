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

type someSubResource struct {
}

var accepted = false
var outOfBandLocation = "http://foo.bar/baz"

func (self *someResource) Index(p PBundle) (interface{}, error) {
	if accepted {
		return nil, HTTPError(http.StatusAccepted, outOfBandLocation)
	}
	return []*someWire{&someWire{1074, "index"}}, nil
}
func (self *someResource) Find(id int64, p PBundle) (interface{}, error) {
	return &someWire{id, "find"}, nil
}
func (self *someResource) Post(i interface{}, p PBundle) (interface{}, error) {
	s := i.(*someWire)
	return &someWire{999, s.Foo}, nil
}
func (self *someResource) Delete(id int64, p PBundle) (interface{}, error) {
	return &someWire{id, "delete!"}, nil
}
func (self *someResource) Put(id int64, i interface{}, p PBundle) (interface{}, error) {
	s := i.(*someWire)
	return &someWire{id, s.Foo + "?"}, nil
}

func (self *someSubResource) Put(id int64, i interface{}, p PBundle) (interface{}, error) {
	panic("NYI")
}
func (self *someSubResource) Find(id int64, p PBundle) (interface{}, error) {
	panic("NYI")
}
func (self *someSubResource) Delete(id int64, p PBundle) (interface{}, error) {
	panic("NYI")
}
func (self *someSubResource) Index(p PBundle) (interface{}, error) {
	panic("NYI")
}

func (self *someSubResource) Post(i interface{}, p PBundle) (interface{}, error) {
	w := p.ParentValue(&someWire{}).(*someWire)
	sub := i.(*someSubWire)
	return &someSubWire{
		Id:       668, //the neighbor of the beast
		MyParent: w.Id,
		Bar:      sub.Bar,
	}, nil
}

type someSubWire struct {
	Id       int64
	MyParent int64
	Bar      string
}
type someWire struct {
	Id  int64
	Foo string
}

/*---------------------------------------------------------------------------------------*/

func setupMux(f RestAll, s RestAll) *ServeMux {
	io := NewRawIOHook(&JsonDecoder{}, &JsonEncoder{}, nil)
	raw := NewRawDispatcher(io, nil, nil, "/rest")

	raw.Rez(&someWire{}, f)

	if s != nil {
		raw.SubResourceSeparate(&someWire{}, &someSubWire{}, s, s, s, s, s)
	}

	mux := NewServeMux()
	//note this prefix ends up _on_ all resources
	mux.Dispatch("/rest/", raw)
	return mux
}

func TestResourceMethods(t *testing.T) {
	resource := &someResource{}
	mux := setupMux(resource, nil)
	go func() {
		http.ListenAndServe(":8189", mux)
	}()

	client := new(http.Client)

	w := makeRequestAndCheckStatus(t, client, "GET", "http://localhost:8189/rest/somewire/", "",
		http.StatusOK, true)
	checkBody(t, w, 1074, "index")

	body := "{ \"Id\":-1, \"Foo\":\"grik\"}"
	w = makeRequestAndCheckStatus(t, client, "POST", "http://localhost:8189/rest/somewire", body,
		http.StatusCreated, false)
	checkBody(t, w, 999, "grik")

	w = makeRequestAndCheckStatus(t, client, "GET", "http://localhost:8189/rest/somewire/2989/?foo=bar", "",
		http.StatusOK, false)
	checkBody(t, w, 2989, "find")

	body = "{ \"Id\":214, \"Foo\":\"grak\"}"
	w = makeRequestAndCheckStatus(t, client, "PUT", "http://localhost:8189/rest/somewire/214", body,
		http.StatusOK, false)
	checkBody(t, w, 214, "grak?")

	w = makeRequestAndCheckStatus(t, client, "DELETE", "http://localhost:8189/rest/somewire/76199", "",
		http.StatusOK, false)
	checkBody(t, w, 76199, "delete!")
}

func TestSubResourceMethods(t *testing.T) {
	resource := &someResource{}
	subresource := &someSubResource{}
	mux := setupMux(resource, subresource)
	go func() {
		http.ListenAndServe(":8192", mux)
	}()
	client := new(http.Client)
	//CREATE THE PARENT
	body := "{ \"Id\":-1, \"Foo\":\"grik\"}"
	w := makeRequestAndCheckStatus(t, client, "POST",
		"http://localhost:8192/rest/somewire", body,
		http.StatusCreated, false)
	checkBody(t, w, 999, "grik")

	//CREATE THE CHILD
	body = "{ \"Id\":-1, \"Bar\":\"grak\"}"
	req := makeReq(t, "POST", "http://localhost:8192/rest/somewire/999/somesubwire/", body)
	resp, err := client.Do(req)
	checkHttpStatus(t, resp, err, http.StatusCreated)

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

func checkBody(t *testing.T, returned *someWire, id int64, s string) {
	if returned.Id != id {
		t.Errorf("Expected id %d but got %d\n", id, returned.Id)
	}
	if string(returned.Foo) != s {
		t.Errorf("Expected to see string '%s' but got '%s'\n", s, returned.Foo)
	}
}

type badlyWrittenResource struct {
}

func (self *badlyWrittenResource) Find(id int64, p PBundle) (interface{}, error) {
	s := "foo"
	return &s, nil
}

func TestBadResource(t *testing.T) {

	bad := &badlyWrittenResource{}

	io := NewRawIOHook(&JsonDecoder{}, &JsonEncoder{}, nil)
	raw := NewRawDispatcher(io, nil, nil, "/rest")
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

	mux := setupMux(nil, nil)
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

	mux := setupMux(nil, nil)
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
	//t.Logf("make REQ: %s, %s, ---- %s", method, url, body)
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

func TestResponseCodes(t *testing.T) {
	resource := &someResource{}
	mux := setupMux(resource, nil)
	go func() {
		http.ListenAndServe(":8190", mux)
	}()

	accepted = true
	resp, err := http.Get("http://localhost:8190/rest/somewire")
	checkHttpStatus(t, resp, err, http.StatusAccepted)

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("could not ready body of response: %s", err)
	}
	if strings.TrimSpace(string(b)) != outOfBandLocation {
		t.Fatalf("didn't find expected location in body: %s", string(b))
	}
}
