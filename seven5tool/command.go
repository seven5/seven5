package seven5tool

import (
	"log"
)

//Command is the structure used by all the seven5tool commands for dispatching calls to them and for 
//displaying help messages.

type Command struct {
	Fn func([]string, *log.Logger)
	Name string
	ShortExplanation string
	LongerHelp string
}


