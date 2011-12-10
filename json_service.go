package seven5

import (
	"mongrel2"
)

type JsonRunner interface {
	mongrel2.JsonHandler
	RunJson(config *ProjectConfig, target Jsonified)
}


type Jsonified interface {
	Named
	ProcessJson(*mongrel2.JsonRequest) []*mongrel2.JsonResponse
}


type JsonRunnerDefault struct {
	*mongrel2.JsonHandlerDefault
	In  chan *mongrel2.JsonRequest
	Out chan *mongrel2.JsonResponse
	
}

func (self *JsonRunnerDefault) RunJson(config *ProjectConfig, target Jsonified) {
	i := make(chan *mongrel2.JsonRequest)
	o := make(chan *mongrel2.JsonResponse)

	self.In=i
	self.Out=o

	go self.ReadLoop(self.In)
	go self.WriteLoop(self.Out)

	for {
		req:=<-self.In
		for _, resp:=range target.ProcessJson(req) {
			if resp!=nil {
				self.Out <- resp
			}
		}
	}
}

//Shutdown here is a bit trickier than it might look.  This sends a close message through
//the channels so the receiving goroutine can close down cleanly.  This is important because
//from a different thread (or goroutine) 0MQ will barf if you try to close the sockets 
//directly.
func (self *JsonRunnerDefault) Shutdown() {
	//this check is needed because if you call shutdown before things get rolling, you'll
	//try to close a nil channel
	if self.In != nil {
		close(self.In)
	}
	if self.Out != nil {
		close(self.Out)
	}
}

