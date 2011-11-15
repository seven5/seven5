package seven5

import "mongrel2"

type Handler interface {
	Run(config *ProjectConfig, in chan *mongrel2.Request, out chan *mongrel2.Response)
}