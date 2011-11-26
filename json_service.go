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
}

func (self *JsonRunnerDefault) RunJson(config *ProjectConfig, target Jsonified) {
	in := make(chan *mongrel2.JsonRequest)
	out := make(chan *mongrel2.JsonResponse)

	go self.ReadLoop(in)
	go self.WriteLoop(out)

	for {
		req:=<-in
		for _, resp:=range target.ProcessJson(req) {
			if resp!=nil {
				out <- resp
			}
		}
	}
}
