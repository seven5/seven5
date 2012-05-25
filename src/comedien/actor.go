package comedien

import (
	"log"
	"reflect"
	"os"
)

// Actor is the basis of the very simple Actor system (ala Akka or Glam)
// that we are using to implement asynchronous programming techniques.
type Actor interface {
	//ProcessMessage is called to process a single message of the 
	//appropriate type for this Actor.  The code that implements this
	//message should do a type assertion that the correct type has been
	//passed to them.  
	ProcessMessage(Message)
	//MessageType returns the reflected type of the type of Message 
	//(which must implement Message) that it can consume.
	MessageType() reflect.Type
}

// Messages are the bundles of data received by Actors.
type Message interface {
}

var (
	logger = log.New(os.Stdout, "comedien", log.LstdFlags)
)
