package seven5

import (
	"bytes"
	"encoding/json"
	"fmt"
	_ "fmt"
	"log"
	"net/http"
	"strings"
)

type Encoder interface {
	Encode(wireType interface{}, prettyPrint bool) (string, error)
}

type JsonEncoder struct {
}

func (self *JsonEncoder) Encode(wireType interface{}, prettyPrint bool) (string, error) {
	var buff []byte
	var err error
	if prettyPrint {
		buff, err = json.MarshalIndent(wireType, "", " ")
	} else {
		buff, err = json.Marshal(wireType)
	}
	if err != nil {
		return "", err
	}
	result := string(buff)
	return strings.Trim(result, " "), nil
}

type Decoder interface {
	Decode([]byte, interface{}) error
}

type JsonDecoder struct {
}

//Decode is called to turn a body supplied by the client into an object of the appropriate
//wire type.  Note that the interface{} passed here _must_ be pointer.
func (self *JsonDecoder) Decode(body []byte, wireType interface{}) error {
	err := json.Unmarshal(body, wireType)
	return err
}

//Utility routine to send a json blob to the client side.  Encoding errors
//are logged to the terminal and the client will recv a 500 error.
func SendJson(w http.ResponseWriter, i interface{}) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(i); err != nil {
		http.Error(w, fmt.Sprintf("unable to encode: %v", err), http.StatusInternalServerError)
		log.Printf("[SENDJSON] unable to encode json output in SendUserDetails: %v", err)
		return err
	}
	count := 0
	for count < buf.Len() {
		w, err := w.Write(buf.Bytes()[count:])
		if err != nil {
			log.Printf("[SENDJSON] failed to write: %v", err)
			return err
		}
		count += w
	}
	return nil
}
