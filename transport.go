package seven5

import (
	"github.com/seven5/gozmq"
)
//Transport, for now, just holds the ability to shutdown the transport in a controlled way
//for test code.  This eventually should allow connection to different http connectivity
//types.
type Transport interface {
	Shutdown() error
}

//NewTransport gets you a transport.  For now, this is always ZeroMQ because that is the way
//you talk to Mongrel2.
func NewTransport(c gozmq.Context) Transport {
	return &zmqTransport{c}
}

//zmqTransport is a just a tiny wrapper for the ZeroMQ context object.
type zmqTransport struct {
	Ctx gozmq.Context
}

//Shutdown is used to close everything down when the test code wants to shutdown the web
//processing.
func (self *zmqTransport) Shutdown() error {
	self.Ctx.Close()
	return nil
}
