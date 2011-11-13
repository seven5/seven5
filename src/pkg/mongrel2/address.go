package mongrel2

import (
	"fmt"
)

// HandlerAddr is returned in response a request for the location (in 
// 0mq terms) of a particular named handler.  It contains the mongrel2
// necessary specifications of the Pull and Pub sockets, plus the unique
// id of the handler.  The Pull socket is assigned the lower of the two
// port numbers.
type HandlerAddr struct {
	Name     string
	PubSpec  string
	PullSpec string
	UUID     string
}

var (
	//handler is the private mapping that keps the binding between names and
	//the handler addresses.
	handler = make(map[string]*HandlerAddr)

	//currentPort is the next port number to be assigned by the GetAssignment
	//function.  It never decreases.
	currentPort = 10070
)

//GetAssignment is used to find a mapping for a handler of a given name.  If
//name has been previously assigned a HandlerAddr the previously allocated
//address is returned, otherwise a new HandlerAddr is created and returned.
func GetHandlerAddress(name string) (*HandlerAddr, error) {
	a := handler[name]
	if a != nil {
		return a, nil
	}
	result := new(HandlerAddr)
	result.Name= name
	result.PullSpec = fmt.Sprintf("tcp://127.0.0.1:%d", currentPort)
	currentPort++
	result.PubSpec = fmt.Sprintf("tcp://127.0.0.1:%d", currentPort)
	currentPort++
	u, err := Type4UUID()
	if err != nil {
		return nil, err
	}
	result.UUID = u
	handler[name] = result
	return result, nil
}
