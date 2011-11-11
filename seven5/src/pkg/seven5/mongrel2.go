package seven5

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"github.com/alecthomas/gozmq"
	"io"
	"strconv"
	"strings"
)

type Mongrel2 struct {
	Context   gozmq.Context
	InSocket,OutSocket  gozmq.Socket	
	PullSpec, PubSpec, Identity string
}

type Request struct {
	RawRequest []byte
	Body       []byte
	ServerId   string
	ClientId   int
	BodySize   int
	Path       string
	Header     map[string]string
}

type Response struct {
	UUID       string
	Client     []int
	Body       string
	StatusCode int
	StatusMsg  string
	Header    map[string]string
}

const HTTPFORMAT = "HTTP/1.1 %(code)s %(status)s\r\n%(headers)s\r\n\r\n%(body)s"

func (self *Mongrel2) init(pullSpec string, pubSpec string, id string) error {

	context, err := gozmq.NewContext()
	if err != nil {
		return err
	}

	s, err := context.NewSocket(gozmq.PULL)
	if err != nil {
		return err
	}
	self.InSocket = s

	err = self.InSocket.Connect(pullSpec)
	if err != nil {
		return err
	}

	s, err = context.NewSocket(gozmq.PUB)
	if err != nil {
		return err
	}
	self.OutSocket = s
	
	err = self.OutSocket.SetSockOptString(gozmq.IDENTITY,id)
	if err!=nil {
		return err
	}
	

	//err=self.OutSocket.Bind(pubSpec)
	
	err = self.OutSocket.Connect(pubSpec)
	if err != nil {
		return err
	}


	return nil
}

func NewMongrel2(pullSpec string, pubSpec string, id string) *Mongrel2{

	result:=new(Mongrel2)
	result.PullSpec=pullSpec
	result.PubSpec=pubSpec
	result.Identity=id
	return result
}

func (self *Mongrel2) ReadMessage() (*Request, error) {
	if self.Context == nil {
		if err := self.init(self.PullSpec, self.PubSpec, self.Identity); err != nil {
			return nil, err
		}
	}
	fmt.Printf("about to try recv...\n")
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

func (self *Mongrel2) WriteMessage(response *Response) error {

	if self.Context == nil {
		if err := self.init(self.PullSpec, self.PubSpec, self.Identity); err != nil {
			return err
		}
	}
	fmt.Printf("about to try write...\n")


	c := make([]string, len(response.Client), len(response.Client))
	for i, x := range response.Client {
		c[i] = strconv.Itoa(x)
	}
	clientList := strings.Join(c, " ")

	//create the properly mangled body in HTTP format
	buffer := new(bytes.Buffer)
	if response.StatusMsg == "" {
		buffer.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n",200,"OK"))
	} else {
		buffer.WriteString(fmt.Sprintf("HTTP/1.1 %d %s\r\n",response.StatusCode,response.StatusMsg))
	}

	buffer.WriteString(fmt.Sprintf("Content-Length: %d\r\n",len(response.Body)))

	for k,v:=range response.Header {
		buffer.WriteString(fmt.Sprintf("%s: %s\r\n",k,v))
	}
	
	//critical, separating extra newline
	buffer.WriteString("\r\n")
	//then the body
	buffer.WriteString(response.Body)
	
	//now we have the true size the body and can put it all together
	msg := fmt.Sprintf("%s %d:%s, %s", response.UUID, len(clientList), clientList, buffer.String())

	buffer = new(bytes.Buffer)
	buffer.WriteString(msg)

	fmt.Printf("message:\n%s\nwhich is %d bytes\n", msg, len(buffer.Bytes()))
	
	err := self.OutSocket.Send(buffer.Bytes(),0)
	return err
}

/*
 * Posted to go-nuts by Russ Cox
 * http://groups.google.com/group/golang-nuts/msg/5ebbdd72e2d40c09
 */
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
