package seven5

import (
	"fmt"
	"net/http"
	"testing"
	"io/ioutil"
	"strings"
)

var called = 0x00
var allow = false
var newRezId = 1200010
type allowResource struct {
	//resource	
}

func (self *allowResource) Index(pb PBundle) (interface{}, error) {
	called |= 0x01
	return nil, nil
}
func (self *allowResource) Find(id Id, pb PBundle) (interface{}, error) {
	called |= 0x02
	return nil, nil
}
func (self *allowResource) Delete(id Id, pb PBundle) (interface{}, error) {
	called |= 0x04
	return nil, nil
}

func (self *allowResource) Put(id Id, i interface{}, pb PBundle) (interface{}, error) {
	called |= 0x08
	return nil, nil
}

func (self *allowResource) Post(i interface{}, pb PBundle) (interface{}, error) {
	called |= 0x10
	return &someWire{Id(newRezId),""}, nil
}

func (self *allowResource) AllowRead(pb PBundle) bool {
	return allow
}
func (self *allowResource) AllowWrite(pb PBundle) bool {
	return allow
}
func (self *allowResource) Allow(id Id, method string, pb PBundle) bool {
	return allow
}

func TestAllow(t *testing.T) {
	base := NewBaseDispatcher("myappname",nil)

	serveMux := NewServeMux()
	serveMux.Dispatch("/rest/", base)
	
	res := &allowResource{}
	base.Rez(&someWire{}, res)

	go func() {
		http.ListenAndServe(":8191", serveMux)
	}()

	client := new(http.Client)

	for _, b := range []bool{false, true} {
		allow = b
		status := http.StatusUnauthorized
		if b {
			status = http.StatusOK
		}
		makeRequestCheckStatusNullBody(t, client, "GET", "http://localhost:8191/rest/somewire", "", status)
		makeRequestCheckStatusNullBody(t, client, "GET", "http://localhost:8191/rest/somewire/12", "", status)
		makeRequestCheckStatusNullBody(t, client, "POST", "http://localhost:8191/rest/somewire", "{}", status)
		makeRequestCheckStatusNullBody(t, client, "PUT", "http://localhost:8191/rest/somewire/345", "{}", status)
		makeRequestCheckStatusNullBody(t, client, "DELETE", "http://localhost:8191/rest/somewire/678", "", status)
	}
}

func makeRequestCheckStatusNullBody(t *testing.T, client *http.Client, method string, url string, body string,
	status int) {

	req := makeReq(t, method, url, body)
	resp, err := client.Do(req)
	if method=="POST" && status==http.StatusOK{
		status=http.StatusCreated
	}
	checkHttpStatus(t, resp, err, status)
	if method=="POST" && status==http.StatusOK {
		loc:=resp.Header.Get("Location")
		if loc!=fmt.Sprintf("/rest/frobnitz/%d",newRezId) {
			t.Fatalf("Didn't find the new resource we were expecting on POST (create): %s\n",loc)
		}
	}
	all, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read the body: %s", err)
	}
	read := string(all)
	if (status==http.StatusUnauthorized) {
		if !strings.HasPrefix(read,"Not authorized") {
			t.Errorf("expected not authorized message but got '%s'", read)
		}
	} else {
		if method=="POST" {
			if strings.Index(read, fmt.Sprintf("%d",newRezId))==-1 {
				t.Errorf("Probably a bad body found on POST (%d) '%s'", len(read), read)
			}
		} else {
			//nil is "null" in JSON
			if read != "null" {
				t.Errorf("expected null body but got (%d) '%s'", len(read), read)
			}
		}
	}
}
