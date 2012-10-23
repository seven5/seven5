package seven5

import (
	"testing"
	"fmt"
	"net/http"
)

/*-------------------------------------------------------------------------*/
type ExampleIndexer_correct struct {
}

func (STATELESS *ExampleIndexer_correct) Index(headers map[string]string,queryParams map[string]string) (string,*Error)  {
		return "[]",nil
}
func (STATELESS *ExampleIndexer_correct) IndexDoc() []string{
	return []string{"FOO","bar","Baz"}
}

/*-------------------------------------------------------------------------*/
type Ox struct {
	Id Id
	IsLarge Boolean
}

type ExampleFinder_correct struct {
}

func (STATELESS *ExampleFinder_correct)	Find(id Id, headers map[string]string, 
	queryParams map[string]string) (string,*Error) {
		switch (int64(id)) {
		case 0,1:
			return JsonResult(&Ox{id,true},false);
		}	
		return NotFound();
}
func (STATELESS *ExampleFinder_correct)	FindDoc() []string {
	return []string{"How can you lose an ox?","fleazil","frack for love"}
}
/*-------------------------------------------------------------------------*/
/*                            SHARED VERIFIERS                             */
/*-------------------------------------------------------------------------*/
func verifyErrorCode(T *testing.T, err *Error, expected int, msg string) {
	if err==nil {
		T.Errorf("expected error with code %d but got no error at all (%s)", expected, msg)
		return
	}
	if err.StatusCode != expected {
		T.Errorf("expected error code %d but got %d (%s)", expected, err.StatusCode, msg)
	}
}

func verifyDispatchError(T *testing.T, h Handler, errorMap map[string]int) {
	for k,v := range errorMap {
		json, err := h.Dispatch("GET",k, emptyMap, emptyMap)
		if json!="" {
			T.Errorf("expected no json result on an error : GET %s",k)
			continue
		}
		verifyErrorCode(T,err,v, fmt.Sprintf("GET %s",k))
	}
}

func verifyNoError(T *testing.T, json string, err *Error, msg string){
	if err!=nil || json=="" {
		T.Fatalf("didn't expect error on %s but got %+v",msg, err);
	}
}
func verifyJsonContent(T *testing.T, json string, expected string, msg string){
	if json!=expected {
		T.Fatalf("expected json of '%s' but got '%s' from %s",expected, json,msg);
	}
}

/*-------------------------------------------------------------------------*/
/*                                 TEST CODE                               */
/*-------------------------------------------------------------------------*/
var emptyMap = make(map[string]string)


func TestResourceMappingForIndexerFinder(T *testing.T) {
	h := NewSimpleHandler()

	h.AddFindAndIndex("ox",&ExampleFinder_correct{},&ExampleIndexer_correct{},Ox{})
	h.AddFindAndIndex("fart",&ExampleFinder_correct{},nil,Ox{})

	json, err := h.Dispatch("GET","/ox/", emptyMap, emptyMap)
	verifyNoError(T,json,err,"GET /ox/")
	verifyJsonContent(T,json,"[]", "GET /ox/")	


	errorMap :=map[string]int{
		"/oxen/": http.StatusNotFound,
		"/ox/fart": http.StatusBadRequest,
		"/oxen/123": http.StatusNotFound,
		"/ox/123": http.StatusNotFound, //too large an id
		"/fart/": http.StatusNotImplemented, //name is registered but not implemented (no Finder)
		"/fart/123": http.StatusNotFound,  //to large, same as /ox/123
	}

	verifyDispatchError(T, h, errorMap)

	json, err = h.Dispatch("GET","/ox/0", emptyMap, emptyMap)
	verifyNoError(T,json,err, "GET /ox/0")
	verifyJsonContent(T,json,"{\"Id\":0,\"IsLarge\":true}", "GET /ox/0")	
}

