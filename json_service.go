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

type Jsonified interface {
	Named
	JsonRunner() JsonRunner
	ProcessJson(map[string]interface{}) map[string]interface{}
}

