package mongrel2

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alecthomas/gozmq"
	"io"
	"strconv"
	"strings"
)

//Handler is a low-level abstraction for connecting, via 0MQ, to a mongrel2 
//server.  This Handler type maps directly to the mongrel2 notion from the
//mongrel2 documentation.  The class uses one go routine to read messages from
//a mongrel2 server (or a pack of servers) and send them through a channel for
//higher level processing.  Similarly, it has a goroutine that reads from a
//channel in go and then queues messages for transmission to mongrel.  Although
//callers of NewHandler() may supply buffered channels, this is not advised.
//The Out channel is asynchronous with respect to mongrel2, so buffering on this
//channel should not be necessary.  The caller should create enough goroutines
//to read in the In channel to insure that this Handler will not block.  The
//purpose of this type is to turn the mongrel2 handler protocol into go 
//abstractions, not to implement semantics.
//
//Because this abstraction uses 0MQ and Mongrel2's (small) protocol on top of it,
//it is possible to do N:M communication by creating several of these structs with
//different HandlerAddresses.  
//See http://mongrel2.org/static/mongrel2-manual.html#x1-680005.1.7
type Handler struct {
	Context                     gozmq.Context
	InSocket, OutSocket         gozmq.Socket
	PullSpec, PubSpec, Identity string
	In                          chan *Request
	Out                         chan *Response
}

//Request structs are the "raw" information sent to the handler by the Mongrel2 server.
//The primary fields of the mongrel2 protocol are broken out in this struct and the
//headers (supplied by the client, passed through by Mongrel2) are included as a map.
//The RawRequest slice is the byte slice that holds all the data.  The Body byte slice
//points to the same underlying storage.  The other fields, for convenienced have been
//parsed and _copied_ out of the RawRequest byte slice.
type Request struct {
	RawRequest []byte
	Body       []byte
	ServerId   string
	ClientId   int
	BodySize   int
	Path       string
	Header     map[string]string
}

//Response structss are sent back to Mongrel2 servers. The Mongrel2 server you wish
//to target should be specified with the UUID and the client of that server you wish
//to target should be in the ClientId field.  Note that this is a slice since you
//can target up to 128 clients with a single Response struct.  The other fields are
//passed through at the HTTP level to the client or clients.  There is no need to 
//set the Content-Length header as this is added automatically.  The easiest way
//to correctly target a Response is by looking at the values supplied in a Request
//struct.
type Response struct {
	ServerId   string
	ClientId   []int
	Body       string
	StatusCode int
	StatusMsg  string
	Header     map[string]string
}

//initZMQ creates the necessary ZMQ machinery and sets the fields of the
//Mongrel2 struct.
func (self *Handler) initZMQ() error {

	c, err := gozmq.NewContext()
	if err != nil {
		return err
	}
	self.Context = c

	s, err := self.Context.NewSocket(gozmq.PULL)
	if err != nil {
		return err
	}
	self.InSocket = s

	err = self.InSocket.Connect(self.PullSpec)
	if err != nil {
		return err
	}

	s, err = self.Context.NewSocket(gozmq.PUB)
	if err != nil {
		return err
	}
	self.OutSocket = s

	err = self.OutSocket.SetSockOptString(gozmq.IDENTITY, self.Identity)
	if err != nil {
		return err
	}

	//not sure why this generates an error... seems legit
/*	var opt int64
	opt=0
	err = self.OutSocket.SetSockOptInt64(gozmq.LINGER, &opt)
	if err != nil {
		return err
	}*/

	err = self.OutSocket.Connect(self.PubSpec)
	if err != nil {
		return err
	}

	return nil
}

//NewHandler creates a Handler struct that can handle requests from a Mongrel2
//server and returns a pointer to the initialized connection.  The address
//parameter is typically created by a call to the function GetHandlerAddress().
//Clients must supply the two channels used to communicate with the raw level
//of the mongrel2 protocol.
func NewHandler(address *HandlerAddr, in chan *Request, out chan *Response) (*Handler, error) {

	result := new(Handler)
	result.PullSpec = address.PullSpec
	result.PubSpec = address.PubSpec
	result.Identity = address.UUID
	result.In = in
	result.Out = out
	err := result.initZMQ()
	if err != nil {
		return nil, errors.New("0mq init:" + err.Error())
	}
	//read loop
	go result.readloop()
	//write loop
	go result.writeloop()
	return result, nil
}

// readloop is a loop that reads mongrel two message until it gets an error.
func (self *Handler) readloop() {
	for {
		r, err := self.ReadMessage()
		if err != nil {
			e := err.(gozmq.ZmqErrno)
			if (e==gozmq.ETERM) {
				//fmt.Printf("read loop ignoring ETERM...\n")
				return
			}
			panic(err)
		}
		self.In <- r
	}
}
// writeloop is a loop that sends mongrel two message until it gets an error
// or a message to close.
func (self *Handler) writeloop() {
	for {
		m := <-self.Out
		if m == nil {
			return //end of goroutine b/c of shutdown
		}

		err := self.WriteMessage(m)
		if err != nil {
			e := err.(gozmq.ZmqErrno)
			if (e==gozmq.ETERM) {
				//fmt.Printf("write loop ignoring ETERM...\n")
				return
			}
			panic(err)
		}
	}

}

//ReadMessage creates a new Request struct based on the values sent from a Mongrel2
//instance. This call blocks until it receives a Request.  Note that you can have
//several different Handler structs all waiting on messages from the same
//server and they will be 
//delivered in a round-robin fashion.  This call tries to be efficient and look
//at each byte only when necessary.  The body of the request is not examined by
//this method.
func (self *Handler) ReadMessage() (*Request, error) {
	req, err := self.InSocket.Recv(0)
	if err != nil {
		return nil, err
	}

	endOfServerId := readSome(' ', req, 0)
	serverId := string(req[0:endOfServerId])

	endOfClientId := readSome(' ', req, endOfServerId+1)
	clientId, err := strconv.Atoi(string(req[endOfServerId+1 : endOfClientId]))
	if err != nil {
		return nil, err
	}

	endOfPath := readSome(' ', req, endOfClientId+1)
	path := string(req[endOfClientId+1 : endOfPath])

	endOfJsonSize := readSome(':', req, endOfPath+1)
	jsonSize, err := strconv.Atoi(string(req[endOfPath+1 : endOfJsonSize]))
	if err != nil {
		return nil, err
	}

	jsonMap := make(map[string]string)
	jsonStart := endOfJsonSize + 1

	if jsonSize > 0 {
		err = json.Unmarshal(req[jsonStart:jsonStart+jsonSize], &jsonMap)
		if err != nil {
			return nil, err
		}
	}

	bodySizeStart := (jsonSize + 1) + jsonStart
	bodySizeEnd := readSome(':', req, bodySizeStart)
	bodySize, err := strconv.Atoi(string(req[bodySizeStart:bodySizeEnd]))

	if err != nil {
		return nil, err
	}

	result := new(Request)
	result.RawRequest = req
	result.Body = req[bodySizeStart:bodySizeEnd]
	result.Path = path
	result.BodySize = bodySize
	result.ServerId = serverId
	result.ClientId = clientId
	result.Header = jsonMap

	return result, nil
}

func readSome(terminationChar byte, req []byte, start int) int {
	result := start
	for {
		if req[result] == terminationChar {
			break
		}
		result++
	}
	return result
}

//WriteMessage takes a Response structs and enques it for transmission.  This call 
//does _not_ block.  The Response struct must be targeted for a specific server
//(ServerId) and one or more clients (ClientID).  The Response struct may be received
//by many Mongrel2 server instances, but only the server addressed in the Request
//will transmit process the response --sending the result on to the client or clients.
func (self *Handler) WriteMessage(response *Response) error {
	c := make([]string, len(response.ClientId), len(response.ClientId))
	for i, x := range response.ClientId {
		c[i] = strconv.Itoa(x)
	}
	clientList := strings.Join(c, " ")

	//create the properly mangled body in HTTP format
	buffer := new(bytes.Buffer)
	if response.StatusMsg == "" {
		buffer.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", 200, "OK"))
	} else {
		buffer.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n", response.StatusCode, response.StatusMsg))
	}

	buffer.WriteString(fmt.Sprintf("Content-Length: %d\r\n", len(response.Body)))

	for k, v := range response.Header {
		buffer.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}

	//critical, separating extra newline
	buffer.WriteString("\r\n")
	//then the body
	buffer.WriteString(response.Body)

	//now we have the true size the body and can put it all together
	msg := fmt.Sprintf("%s %d:%s, %s", response.ServerId, len(clientList), clientList, buffer.String())

	buffer = new(bytes.Buffer)
	buffer.WriteString(msg)

	err := self.OutSocket.Send(buffer.Bytes(), 0)
	return err
}
//Type4UUID generates a RFC 4122 compliant UUID.  This code was originally posted
//by Ross Cox to the go-nuts mailing list.
//http://groups.google.com/group/golang-nuts/msg/5ebbdd72e2d40c09
func Type4UUID() (string, error) {
	b := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, b)
	if err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0F) | 0x40
	b[8] = (b[8] &^ 0x40) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}

//Shutdown cleans up the resources associated with this mongrel2 connection.
//Normally this function should be part of a defer call that is immediately after
//allocating the resources, like this:
//	mongrel:=NewHandler(...)
//  defer mongrel.Shutdown()
func (self *Handler) Shutdown() error {

	//tell writeloop to exit
	close(self.Out)

	//tell everybody listening on the input channel to exit
	close(self.In)

	//dump the ZMQ level sockets
	if err := self.InSocket.Close(); err != nil {
		return err
	}
	if err := self.OutSocket.Close(); err != nil {
		return err
	}

	self.Context.Close()
	return nil
}
