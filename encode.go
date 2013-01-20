package seven5

import (
	"encoding/json"
	"strings"
	_ "fmt"
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
		return "",err
	}
	result := string(buff)
	return strings.Trim(result, " "), nil
}

type Decoder interface {
	Decode([]byte, interface{}) (error)
}

type JsonDecoder struct {
}

//Decode is called to turn a body supplied by the client into an object of the appropriate
//wire type.  Note that the interface{} passed here _must_ be pointer.
func (self *JsonDecoder) Decode(body []byte, wireType interface{}) error {
	err:= json.Unmarshal(body,wireType)
	return err
}
