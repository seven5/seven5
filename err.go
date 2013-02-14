package seven5

import (
	"fmt"
)

//Error is a type that can be used by a resource that wants to send a particular
//HTTP response back to the client.  If a resource returns any error _other_ than
//this one, it is considered an internal server error.  This should not be used
//to return 200 "OK" results, use nil instead.
type Error struct {
	StatusCode int 
	Msg string
}

func (self *Error) Error() string {
	return fmt.Sprintf("HTTP Level Error (%d): %s",self.StatusCode, self.Msg)
}

func HTTPError(code int, msg string) *Error {
	return &Error{code,msg}
}