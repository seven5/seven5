package seven5

import (
	"fmt"
	"net/http"
)

//Error is a type that can be used by a resource that wants to send a particular
//HTTP response back to the client.  If a resource returns any error _other_ than
//this one, it is considered an internal server error.  This should not be used
//to return 200 "OK" results, use nil instead.
type Error struct {
	StatusCode int
	Msg        string
}

//error() makes this an implementation of the type error
func (self *Error) Error() string {
	return fmt.Sprintf("HTTP Level Error (%d): %s", self.StatusCode, self.Msg)
}

func HTTPError(code int, msg string) *Error {
	return &Error{code, msg}
}

//WriteError returns an error to the client side.  If the err is of type
//Error, we decode the fields to produce the correct message and response
//code.  Otherwise, we return the string of the error content plus the code
//http.StatusInternalServerError.
func WriteError(w http.ResponseWriter, err error) {
	ourError, ok := err.(*Error)
	if ok {
		http.Error(w, ourError.Msg, ourError.StatusCode)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
