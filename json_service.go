package seven5

import (
	"mongrel2"
)

type JsonRunner interface {
	mongrel2.M2JsonHandler
	RunJson(config *ProjectConfig, target Jsonified)
}

type JsonRunnerDefault struct {
	*mongrel2.M2JsonHandlerDefault
}

func (self *JsonRunnerDefault) RunJson(config *ProjectConfig, target Jsonified) {
	
}

type Jsonified interface {
	Named
	ProcessJson(map[string]interface{}) map[string]interface{}
}

