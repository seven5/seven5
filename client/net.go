package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
)

type FailureFunc func(int, string)
type SuccessFunc func(int64)

//PostNew sends an instance of a wire type to the server.  It returns an error
//only if the Post could not be sent.  The success or failure are indications are
//communicated through the callback functions.
func PostNew(i interface{}, root string, success SuccessFunc, failure FailureFunc) error {
	var w bytes.Buffer
	enc := json.NewEncoder(&w)
	err := enc.Encode(i)
	if err != nil {
		return fmt.Errorf("error encoding post msg: %v ", err)
	}
	name := fmt.Sprintf("%T", i)
	pair := strings.Split(name, ".")
	if len(pair) != 2 {
		panic(fmt.Sprintf("unable to understand type name: %s", name))
	}
	urlname := strings.ToLower(pair[1])
	jquery.Ajax(
		map[string]interface{}{
			"contentType": "application/json",
			"dataType":    "json",
			"type":        "POST",
			"url":         fmt.Sprintf("%s/%s", root, urlname),
			"data":        w.String(),
		}).
		Then(func(valueCreated js.Object) {
		id := valueCreated.Get("Id").Int64()
		if success != nil {
			success(id)
		}
	}).
		Fail(func(p1 js.Object) {
		if failure != nil {
			failure(p1.Get("status").Int(), p1.Get("responseText").Str())
		}
	})

	return nil
}
