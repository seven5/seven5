package seven5

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
)

//Dispatcher is the low-level interface to requests and responses.  Most user level code should not
//need to implement Dispatchers but rather should choose one or more dispatchers to "install"
//in their application. A dispatcher may return an instance of ServeMux (possibly the same
//one passed to it as the first parameter) if it wants the processing of the request to continue
//after it executes; this allows Dispatchers that wrap the entire url space.
type Dispatcher interface {
	Dispatch(*ServeMux, http.ResponseWriter, *http.Request) *ServeMux
}

//ServeMux is drop in replacement for http.ServeMux that implements the Seven5 dispatch protocol
//on top of the existing "handler" abstraction in the net/http package.
type ServeMux struct {
	*http.ServeMux
	err ErrorDispatcher
}

//ErrorDispatcher is a special case of dispatcher that is only invoked when other Dispatchers return
//some type of error condition.  An "error" is defined as an http response code greater than 300 or a panic.
//The original response writer and request are passed to the ErrorDispatch() method to allow
//the error dispatcher to take any action desired.  An error dispatcher that does nothing will
//implicitly allow whatever calls the other dispatcher placed on the response writer to proceed.
//ErrorDispatcher is also called when no dispatcher is found (404).
type ErrorDispatcher interface {
	ErrorDispatch(int, http.ResponseWriter, *http.Request)
	PanicDispatch(interface{}, http.ResponseWriter, *http.Request)
}

//NewServeMux creates a new server mux (compatible with http.ServeMux) with the specified error
//handler, which may be nil.
func NewServeMux() *ServeMux {
	return &ServeMux{
		http.NewServeMux(), nil,
	}
}

func (self *ServeMux) ErrorDispatcher() ErrorDispatcher {
	return self.err
}

func (self *ServeMux) SetErrorDispatcher(e ErrorDispatcher) {
	self.err = e
}

//ServeHTTP is a simple wrapper around the http.ServeMux method of the same name that incorporates
//an error wrapper to allow it to implement the ErrorDispatcher protocol.
func (self *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if self.err != nil {
		w = &ErrWrapper{w, r, self.err}
	}
	self.ServeMux.ServeHTTP(w, r)
}

//Dispatch has the same function as "HandleFunc" on an http.ServeMux with the exception that
//we require the Dispatcher interface rather than a "HandleFunc" function.
func (self *ServeMux) Dispatch(pattern string, dispatcher Dispatcher) {
	h := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Fprintf(os.Stderr, "++++++++++++ PANIC +++++++++++++++++++\n")
				fmt.Fprintf(os.Stderr, "++++++++++++ ORIGINAL ERROR: %v ++++++++++++n", err)
				buf := make([]byte, 16384)
				l := runtime.Stack(buf, false)
				fmt.Fprintf(os.Stderr, "%s\n", string(buf[:l]))
				fmt.Fprintf(os.Stderr, "++++++++++++++++++++++++++++++++++++++\n")

				if self.err != nil {
					self.err.PanicDispatch(err, w, r)
				} else {
					fmt.Fprintf(os.Stderr, "++++++++++++ FORCING ANOTHER PANIC: %v ++++++++++++\n", err)
					panic(err)
				}
			}
		}()
		w.Header().Add("Cache-Control", "no-cache, must-revalidate") //HTTP 1.1
		w.Header().Add("Pragma", "no-cache")                         //HTTP 1.0
		b := dispatcher.Dispatch(self, w, r)
		if b != nil {
			b.ServeHTTP(w, r)
		}
	}
	self.ServeMux.HandleFunc(pattern, h)
}

//ErrWrapper is a wrapper around http.ResponseWriter that holds enough state to call its held
//ErrorDispatcher in case of an error status code being written.
type ErrWrapper struct {
	http.ResponseWriter
	req *http.Request
	err ErrorDispatcher
}

//WriteHeader is a wrapper around the http.ResponseWriter method of the same name.  It simply
//traps status code writes of 300 or greater and calls the error dispatcher to handle it.
func (self *ErrWrapper) WriteHeader(status int) {
	if (status / 100) > 2 {
		self.err.ErrorDispatch(status, self.ResponseWriter, self.req)
	}
}
