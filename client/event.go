package client

import (
	"fmt"
	"github.com/gopherjs/jquery"
	//"honnef.co/go/js/console"
)

//EventFunc is something that can handle an event.  It should not
//return anything, it should use the event object to stop default
//event handling and similar.
type EventFunc func(jquery.JQuery, jquery.Event)

type EventName int

const (
	//keys
	BLUR = iota
	CHANGE
	CLICK
	DBLCLICK
	FOCUS
	FOCUSIN
	FOCUSOUT
	HOVER
	KEYDOWN
	KEYPRESS
	KEYUP
)

func (self EventName) String() string {
	switch self {
	case BLUR:
		return jquery.BLUR
	case CHANGE:
		return jquery.CHANGE
	case CLICK:
		return jquery.CLICK
	case DBLCLICK:
		return jquery.DBLCLICK
	case FOCUS:
		return jquery.FOCUS
	case FOCUSIN:
		return jquery.FOCUSIN
	case FOCUSOUT:
		return jquery.FOCUSOUT
	case HOVER:
		return jquery.HOVER
	case KEYDOWN:
		return jquery.KEYDOWN
	case KEYPRESS:
		return jquery.KEYPRESS
	case KEYUP:
		return jquery.KEYUP
	}
	panic(fmt.Sprintf("unknown event name %v", self))
}

type eventHandler struct {
	name EventName
	j    jquery.JQuery
	fn   EventFunc
}

func (self *eventHandler) register() {
	self.j.On(self.name.String(), self.handle)
}

func (self *eventHandler) handle(event jquery.Event) {
	self.fn(self.j, event)
	DrainEagerQueue()
}
