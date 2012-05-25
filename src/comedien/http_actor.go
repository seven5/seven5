package comedien

import (
	"net/http"
	"reflect"
)

//HttpActor defines a type of Actor that can consume HTTP messages. The messages
//should be of type *http.Request.
type HttpRcvActor struct {
}


// Process message takes an http.Request message and processes it.
func (self *HttpRcvActor) ProcessMessage(m Message) {
	msg := m.(*http.Request)
	logger.Printf("type of message is %s\n",msg.Method)
	logger.Printf("URI of message is %s\n",msg.RequestURI)
}

// MessageType retuns the reflective type of the objects we
// process.
func (self *HttpRcvActor) MessageType() reflect.Type {
	return reflect.TypeOf(&http.Request{})
}
