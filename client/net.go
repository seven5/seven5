package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
)

// AjaxError is returned on the error channel after a call to an Ajax method.
type AjaxError struct {
	StatusCode int
	Message    string
}

//AjaxPut behaves indentically to AjaxPost other than using the method PUT.
func AjaxPut(ptrToStruct interface{}, path string) (chan interface{}, chan AjaxError) {
	return putPostDel(ptrToStruct, path, "PUT", true)
}

//AjaxDelete behaves indentically to AjaxPost other than using the method DELETE
//and not sending the object to be deleted's contents, just its id. First parameter
//here is just for the type.
func AjaxDelete(ptrToStruct interface{}, path string) (chan interface{}, chan AjaxError) {
	return putPostDel(ptrToStruct, path, "DELETE", false)
}

//AjaxPost sends an instance of a wire type to the server.  The first argument
//should be a wire type and must be a pointer to a struct or this function
//will panic. The value sent to the server is supplied in the first argument.
//The two returned values are a content channel and an error channel.  If the
//call succeeds, the content channel will be sent a different instance of the
//same type as the first argument. If the result from the server cannot be understood
//as the type of the first argument, the special error code 418 will be sent
//on the error channel.  If we fail to encode the object to be sent, the error
//code 420 will be sent on the error channel and no call to the server is made.
func AjaxPost(ptrToStruct interface{}, path string) (chan interface{}, chan AjaxError) {
	return putPostDel(ptrToStruct, path, "POST", true)
}

//AjaxIndex retreives a collection of wire types from the server.
//If the first argument is not a pointer to a slice of pointer to struct,
//it will panic.  The first element should be a slice of wire types.
//The returned values are a content channel and an error channel.  The content
//channel will receive the same type as your first argument if anything.  The error
//channel is used for non-200 http responses and the special error code 418
//is used to indicate that the received json from the server could not be successfully
//parsed as the type of the first argument.
func AjaxIndex(ptrToSliceOfPtrToStruct interface{}, path string) (chan interface{}, chan AjaxError) {
	isPointerToSliceOfPointerToStructOrPanic(ptrToSliceOfPtrToStruct)
	contentCh := make(chan interface{})
	errCh := make(chan AjaxError)
	ajaxRawChannels(ptrToSliceOfPtrToStruct, "", contentCh, errCh, "GET", path)
	return contentCh, errCh
}

//AjaxGet retreives an instance of a wire type from the server and decodes the result as
//Json. If the first argument is not a pointer to a struct, it will panic.
//The first argument should be a wire type that you expect to receive in the success
//case.  The returned values are a content channel and an error channel.  The content
//channel will receive the same type as your first argument if anything.  The error
//channel is used for non-200 http responses and the special error code 418
//is used to indicate that the received json from the server could not be successfully
//parsed as the type of the first argument.  The error code 0 indicates that the
//server could not be contacted at all.
func AjaxGet(ptrToStruct interface{}, path string) (chan interface{}, chan AjaxError) {
	isPointerToStructOrPanic(ptrToStruct)
	contentCh := make(chan interface{})
	errCh := make(chan AjaxError)
	ajaxRawChannels(ptrToStruct, "", contentCh, errCh, "GET", path)
	return contentCh, errCh
}

func ajaxRawChannels(output interface{}, body string, contentChan chan interface{}, errChan chan AjaxError,
	method string, path string) error {

	m := map[string]interface{}{
		"contentType": "application/json",
		"dataType":    "text",
		"type":        method,
		"url":         path,
		"cache":       false,
	}
	if body != "" {
		m["data"] = body
	}

	jquery.Ajax(m).
		Then(func(valueCreated *js.Object) {
		rd := strings.NewReader(valueCreated.String())
		dec := json.NewDecoder(rd)
		if err := dec.Decode(output); err != nil {
			go func() {
				errChan <- AjaxError{418, err.Error()}
			}()
			return
		}
		go func() {
			contentChan <- output
		}()
	}).
		Fail(func(p1 *js.Object) {
		go func() {
			ajaxerr := AjaxError{p1.Get("status").Int(), p1.Get("responseText").String()}
			if p1.Get("status").Int() == 0 {
				ajaxerr.StatusCode = 0
				ajaxerr.Message = "Server not reachable"
			}
			errChan <- ajaxerr
		}()
	})

	return nil
}

//
// HELPERS
//

func typeToUrlName(i interface{}) string {
	name, ok := i.(string)
	if !ok {
		name = fmt.Sprintf("%T", i)
	}
	pair := strings.Split(name, ".")
	if len(pair) != 2 {
		panic(fmt.Sprintf("unable to understand type name: %s", name))
	}
	return strings.ToLower(pair[1])
}

func encodeBody(i interface{}) (string, error) {
	//encode body
	var w bytes.Buffer
	enc := json.NewEncoder(&w)
	err := enc.Encode(i)
	if err != nil {
		return "", fmt.Errorf("error encoding body: %v ", err)
	}
	return w.String(), nil
}

func isPointerToStructOrPanic(i interface{}) reflect.Type {
	t := reflect.TypeOf(i)
	if t.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("expected ptr to struct but got %T", i))
	}
	if t.Elem().Kind() != reflect.Struct {
		panic(fmt.Sprintf("expected ptr to struct but got ptr to %v", t.Elem().Kind()))
	}
	return t
}

func isPointerToSliceOfPointerToStructOrPanic(i interface{}) reflect.Type {
	t := reflect.TypeOf(i)
	if t.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("expected ptr to slice of ptr to struct but got %T", i))
	}
	if t.Elem().Kind() != reflect.Slice {
		panic(fmt.Sprintf("expected ptr to SLICE of ptr to struct but got ptr to %v", t.Elem().Kind()))
	}
	if t.Elem().Elem().Kind() != reflect.Ptr {
		panic(fmt.Sprintf("expected ptr to slice of PTR to struct but got ptr to slice of %v", t.Elem().Elem().Kind()))
	}
	if t.Elem().Elem().Elem().Kind() != reflect.Struct {
		panic(fmt.Sprintf("expected ptr to slice of ptr to STRUCT but got ptr to slice of ptr to %v", t.Elem().Elem().Elem().Kind()))
	}
	return t
}

func putPostDel(ptrToStruct interface{}, path string, method string, sendBody bool) (chan interface{}, chan AjaxError) {
	t := isPointerToStructOrPanic(ptrToStruct)
	output := reflect.New(t.Elem())
	var body string

	contentCh := make(chan interface{})
	errCh := make(chan AjaxError)

	if sendBody {
		var err error
		body, err = encodeBody(ptrToStruct)
		if err != nil {
			go func() {
				errCh <- AjaxError{420, err.Error()}
			}()
			return contentCh, errCh
		}
	}
	ajaxRawChannels(output.Interface(), body, contentCh, errCh, method, path)
	return contentCh, errCh
}

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

//
// DEPRECATED
//

//UnpackJson has been deprecated in favor of the Ajax methods.  This method
//is a naive json unpacker that uses reflection on the go struct to convert
//javascript values.  It cannot handle arbitrary types of fields, cannot handle
//nested structures, nor can it handle the UnmarshalJson interface.
func UnpackJson(ptrToStruct interface{}, jsonBlob *js.Object) error {
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
		if fn == "-" || jsonBlob.Get(fn) == js.Undefined || jsonBlob.Get(fn) == nil {
			continue
		}

		//
		// Time is really useful
		//
		if f.Type().Name() == "Time" && f.Type().PkgPath() == "time" {
			str := jsonBlob.Get(fn).String()
			//2015-01-17T17:48:30.346218Z
			//2006-01-02T15:04:05.999999999Z
			t, err := time.Parse(time.RFC3339Nano, str)
			if err != nil {
				print("warning: could not convert string", str, ":", err)
			} else {
				f.Set(reflect.ValueOf(t))
			}
			continue
		}
		switch f.Type().Kind() {
		case reflect.Int64:
			f.SetInt(jsonBlob.Get(fn).Int64())
		case reflect.Int:
			f.SetInt(int64(jsonBlob.Get(fn).Int()))
		case reflect.String:
			f.SetString(jsonBlob.Get(fn).String())
		case reflect.Float64:
			f.SetFloat(jsonBlob.Get(fn).Float())
		case reflect.Bool:
			f.SetBool(jsonBlob.Get(fn).Bool())
		default:
			//print("warning: %s", fn, " has a type other than int64, string, float64 or bool")
		}
	}
	return nil
}
