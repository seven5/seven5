package seven5

import (
	"fmt"
	"net/http"
	"testing"
)

/***********************************************************************************************/
type countDispatch struct {
	count int
}

func (self *countDispatch) Dispatch(s *ServeMux, w http.ResponseWriter, r *http.Request) *ServeMux {
	self.count++
	return nil
}

func TestDispatchBasic(t *testing.T) {
	d := &countDispatch{}
	serveMux := NewServeMux()
	serveMux.Dispatch("/countme", d)

	go func() {
		http.ListenAndServe(":8088", serveMux)
	}()

	resp, err := http.Get("http://localhost:8088/countme")
	checkHttpStatus(t, resp, err, 200)
	if d.count != 1 {
		t.Errorf("didn't invoke count dispatcher")
	}
	resp, err = http.Get("http://localhost:8088/countme2")
	checkHttpStatus(t, resp, err, 404)
	resp, err = http.Get("http://localhost:8088/")
	checkHttpStatus(t, resp, err, 404)
}

/***********************************************************************************************/
type errDispatch struct {
	count int
	msg   string
}

const BOGUS_ERROR = 427

func (self *errDispatch) ErrorDispatch(status int, w http.ResponseWriter, r *http.Request) {
	self.count++
	w.WriteHeader(BOGUS_ERROR)
}

func (self *errDispatch) PanicDispatch(i interface{}, w http.ResponseWriter, r *http.Request) {
	self.count++
	self.msg = i.(string)

	//we have to be careful because there is a wrapper that will trap error status codes back
	//into ErrorDispatch above and we don't want that
	wrapper := w.(*ErrWrapper)
	//use the "wrapped" response writer which is the real one
	http.Error(wrapper.ResponseWriter, self.msg, 500)
}

func TestDispatchError(t *testing.T) {
	e := &errDispatch{}
	serveMux := NewServeMux()
	serveMux.SetErrorDispatcher(e)
	go func() {
		http.ListenAndServe(":8089", serveMux)
	}()
	resp, err := http.Get("http://localhost:8089/bar")
	checkHttpStatus(t, resp, err, BOGUS_ERROR)
	if e.count != 1 {
		t.Fatalf("failed to invoke error dispatcher")
	}
}

/***********************************************************************************************/
type continueDispatch struct {
	count   int
	wrapped *ServeMux
}

func (self *continueDispatch) Dispatch(s *ServeMux, w http.ResponseWriter, r *http.Request) *ServeMux {
	self.count++
	//rewrite URL path, use with care
	r.URL.Path = "/inner"
	return self.wrapped
}

func TestDispatchContinue(t *testing.T) {
	count := &countDispatch{}

	nestedMux := NewServeMux()
	nestedMux.Dispatch("/inner", count)

	c := &continueDispatch{0, nestedMux}
	serveMux := NewServeMux()
	serveMux.Dispatch("/outer", c)

	go func() {
		http.ListenAndServe(":8090", serveMux)
	}()
	resp, err := http.Get("http://localhost:8090/outer")
	checkHttpStatus(t, resp, err, 200)

	if count.count != 1 {
		t.Fatalf("didn't get hit on inner dispatcher: %d", count.count)
	}
	if c.count != 1 {
		t.Fatalf("didn't get hit on outer dispatcher: %d", c.count)
	}
}

/***********************************************************************************************/
type panicDispatch struct {
}

const FLEAZIL = "fleazil"

func (self *panicDispatch) Dispatch(s *ServeMux, w http.ResponseWriter, r *http.Request) *ServeMux {
	panic(FLEAZIL)
}

func TestDispatchPanic(t *testing.T) {
	p := &panicDispatch{}
	e := &errDispatch{}

	serveMux := NewServeMux()
	serveMux.SetErrorDispatcher(e)
	serveMux.Dispatch("/die", p)

	go func() {
		http.ListenAndServe(":8091", serveMux)
	}()

	resp, err := http.Get("http://localhost:8091/die")
	checkHttpStatus(t, resp, err, 500)

	if e.msg != FLEAZIL {
		t.Errorf("Bad panic message! expected %s but got %s", FLEAZIL, e.msg)
	}
}

/***********************************************************************************************/

func checkHttpStatus(t *testing.T, resp *http.Response, err error, expected int) {
	if err != nil {
		t.Fatalf("couldn't do HTTP request: %s", err)
	}
	if resp.StatusCode != expected {
		t.Fatalf("unexpected http status! expected %d but got %d! (%s)", expected, resp.StatusCode, resp.Status)
	}
}

/***********************************************************************************************/

var portCount = 9000

func BenchmarkDispatching(b *testing.B) {
	b.StopTimer()
	port := portCount
	portCount++
	c := &countDispatch{}
	e := &errDispatch{}
	serveMux := NewServeMux()
	serveMux.SetErrorDispatcher(e)
	serveMux.Dispatch("/count", c)

	go func() {
		http.ListenAndServe(fmt.Sprintf(":%d", port), serveMux)
	}()

	target := fmt.Sprintf("http://localhost:%d/count")
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, _ = http.Get(target)
	}

}

var handleCount = 0

func BenchmarkHandle(b *testing.B) {
	b.StopTimer()
	port := portCount
	portCount++
	h := func(w http.ResponseWriter, r *http.Request) {
		handleCount++
	}
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/count", h)

	go func() {
		http.ListenAndServe(fmt.Sprintf(":%d", port), serveMux)
	}()

	target := fmt.Sprintf("http://localhost:%d/count")

	b.StartTimer()

	for i := 0; i < b.N; i++ {
		_, _ = http.Get(target)

	}

}
