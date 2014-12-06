package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
)

type FailureFunc func(int, string)
type SuccessNewFunc func(int64)
type SuccessPutFunc func(js.Object)

func getFieldName(f reflect.StructField) string {
	name := f.Tag.Get("json")
	jsonPreferred := ""
	if name != "" {
		parts := strings.Split(name, ",")
		for _, part := range parts {
			if part == "-" {
				return "-"
			}
			if part == "omitempty" {
				continue
			}
			jsonPreferred = part
		}
	}
	if jsonPreferred != "" {
		return jsonPreferred
	}
	return f.Name
}
func UnpackJson(ptrToStruct interface{}, jsonBlob js.Object) error {
	t := reflect.TypeOf(ptrToStruct)
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer to struct, but got %v", t.Kind())
	}
	elem := t.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("expected pointer to struct, but got pointer to  %v", elem.Kind())

	}
	v := reflect.ValueOf(ptrToStruct)
	v = v.Elem()
	for i := 0; i < elem.NumField(); i++ {
		f := v.Field(i)
		fn := getFieldName(elem.Field(i))
		if fn == "-" || jsonBlob.Get(fn).IsUndefined() || jsonBlob.Get(fn).IsNull() {
			continue
		}
		switch f.Type().Kind() {
		case reflect.Int64:
			f.SetInt(jsonBlob.Get(fn).Int64())
		case reflect.String:
			f.SetString(jsonBlob.Get(fn).Str())
		case reflect.Float64:
			f.SetFloat(jsonBlob.Get(fn).Float())
		case reflect.Bool:
			f.SetBool(jsonBlob.Get(fn).Bool())
		default:
			print("warning: %s", fn, " has a type other than int64, string, float64 or bool")
		}
	}
	return nil
}

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
