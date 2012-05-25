package comedien

import (
	"reflect"
)

type Supervisor struct {
	channel     chan Message
	messageType reflect.Type
	actor       Actor
}

//Supervise constructs a supervisor for the provided actor.  
func Supervise(actor Actor) *Supervisor {
	return &Supervisor{messageType: actor.MessageType(), actor: actor}
}

// Run starts the process of reading messages from the channel and
// dispatching them to the process message routine. It returns true
// if the decision to exit run was made by the object itself. It returns
// false if the channel was closed.  It will panic if passed a nil value.
func (self *Supervisor) Run() bool {

	for true {
		m, ok := <-self.channel
		if !ok {
			return false
		}
		if m == nil {
			panic("Cannot process a nil value!")
		}
		if !self.ProcessOne(m) {
			panic("Default actor was expecting " + self.messageType.String() +
				"but got " + reflect.TypeOf(m).String() + " message type!")
		}
	}

	return true //can't happen
}

// ProcessOne handles a single message and returns true if the type of message
// was correct for this Actor.  If it was not, it returns false and the caller
// should probably do something drastic, like panic().
func (self *Supervisor) ProcessOne(m Message) bool {
	candidate := reflect.TypeOf(m)
	if candidate != self.messageType {
		return false
	}
	self.actor.ProcessMessage(m)
	return true
}

