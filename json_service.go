package seven5

import (
	"mongrel2"
)

type JsonRunner interface {
	mongrel2.M2JsonHandler
	RunJson(config *ProjectConfig, target Jsonified)
}


type Jsonified interface {
	Named
	ProcessJson(*mongrel2.M2JsonRequest) *mongrel2.M2JsonResponse
}


type JsonRunnerDefault struct {
	*mongrel2.M2JsonHandlerDefault
}

func (self *JsonRunnerDefault) RunJson(config *ProjectConfig, target Jsonified) {
	in := make(chan *mongrel2.M2JsonRequest)
	out := make(chan *mongrel2.M2JsonResponse)

	go self.ReadLoop(in)
	go self.WriteLoop(out)

	for {
		req:=<-in
		resp:=target.ProcessJson(req)
		if resp!=nil {
			out <- resp
		}
	}
}
