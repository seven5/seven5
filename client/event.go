package client

import (
	"fmt"
	"github.com/gopherjs/jquery"
	//"honnef.co/go/js/console"
)

//EventFunc is something that can handle an event.  It should not
//return anything, it should use the event object to stop default
//event handling and similar.
type EventFunc func(jquery.Event)

type EventName int

const (
	//keys
	BLUR EventName = iota
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
	INPUT_EVENT
	MOUSE_ENTER
	MOUSE_LEAVE
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
	case INPUT_EVENT:
		return "input"
	case MOUSE_ENTER:
		return jquery.MOUSEENTER
	case MOUSE_LEAVE:
		return jquery.MOUSELEAVE
	}
	panic(fmt.Sprintf("unknown event name %v", self))
}

type eventHandler struct {
	name EventName
	t    NarrowDom
	fn   EventFunc
}

func (self eventHandler) register() {
	self.t.On(self.name, self.handle)
}

func (self eventHandler) handle(event jquery.Event) {
	self.fn(event)
	DrainEagerQueue()
}
