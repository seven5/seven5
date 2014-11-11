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
type SuccessNewFunc func(int64)
type SuccessPutFunc func(js.Object)

//PutExisting sends a new version of the object given by id to the server.
//It returns an error only if the put could not be started, otherwise success
//or failure are communicated throught the callback functions.  The root should
//be the root of the rest heirarchy, probably "/rest".  The id is not examined
//by this routine, it can be the string representation of an integer or a UDID.
func PutExisting(i interface{}, root string, id string,
	success SuccessPutFunc, failure FailureFunc) error {

	var w bytes.Buffer
	enc := json.NewEncoder(&w)
	err := enc.Encode(i)
	if err != nil {
		return fmt.Errorf("error encoding put msg: %v ", err)
	}

	urlname := typeToUrlName(i)
	jquery.Ajax(
		map[string]interface{}{
			"contentType": "application/json",
			"dataType":    "json",
			"type":        "PUT",
			"url":         fmt.Sprintf("%s/%s/%s", root, urlname, id),
			"data":        w.String(),
			"cache":       false,
		}).
		Then(func(v js.Object) {
		success(v)
	}).
		Fail(func(p1 js.Object) {
		if failure != nil {
			failure(p1.Get("status").Int(), p1.Get("responseText").Str())
		}
	})

	return nil
}
func typeToUrlName(i interface{}) string {
	name := fmt.Sprintf("%T", i)
	pair := strings.Split(name, ".")
	if len(pair) != 2 {
		panic(fmt.Sprintf("unable to understand type name: %s", name))
	}
	return strings.ToLower(pair[1])
}

//PostNew sends an instance of a wire type to the server.  It returns an error
//only if the Post could not be sent.  The success or failure are indications are
//communicated through the callback functions.
func PostNew(i interface{}, root string, success SuccessNewFunc, failure FailureFunc) error {
	var w bytes.Buffer
	enc := json.NewEncoder(&w)
	err := enc.Encode(i)
	if err != nil {
		return fmt.Errorf("error encoding post msg: %v ", err)
	}
	urlname := typeToUrlName(i)
	jquery.Ajax(
		map[string]interface{}{
			"contentType": "application/json",
			"dataType":    "json",
			"type":        "POST",
			"url":         fmt.Sprintf("%s/%s", root, urlname),
			"data":        w.String(),
			"cache":       false,
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
