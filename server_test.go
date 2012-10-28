package seven5

import (
	"testing"
	"fmt"
	"net/http"
	"encoding/json"
	"strings"
	"bytes"
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
type ExamplePoster_correct struct {
}

func (STATELESS *ExamplePoster_correct) Post(headers map[string]string, queryParams map[string]string,
	body string) (string,*Error)  {
	
	var BodyOx Ox
	
	decoder:=json.NewDecoder(strings.NewReader(body))
	err:=decoder.Decode(&BodyOx)
	if err!=nil {
		return BadRequest(fmt.Sprintf("Could not understand json body: %s", err))
	}
	//we "create" a new ox and return the value of IsLarge sent in the body
	return JsonResult(&Ox{456,BodyOx.IsLarge},false);
}

func (STATELESS *ExamplePoster_correct) PostDoc() []string{
	return []string{"Grik","grak","frik", "fleazil"}
}
/*-------------------------------------------------------------------------*/
type ExamplePuter_correct struct {
}

func (STATELESS *ExamplePuter_correct) Put(id Id, headers map[string]string, queryParams map[string]string,
	body string) (string,*Error)  {

	//check that the id in the params matches the id supplied, if it exists
	formId, ok := queryParams["Id"]
	if ok {
		parsedId, err:=ParseId(formId)
		if err!="" {
			return BadRequest(err)
		}
		if id!=parsedId {
			return BadRequest(fmt.Sprintf("You supplied two different ids in the url and form: %d and %d!", id, parsedId))
		}
	}
	//return the "new" values... we just echo them for this test
	size, ok:=queryParams["IsLarge"]
	if !ok {
		return BadRequest("You can't send a PUT to the Ox without a size! Nothing to change! (Maybe try GET?)")
	}
	size=strings.ToLower(size)
	if size!="false" && size!="true" {
		return BadRequest(fmt.Sprintf("Bad Boolean value for IsLarge: %s", size))
	}
	sent:=Ox{id, false}
	if size=="true" {
		sent.IsLarge = true
	}
	
	//prepare response
	var buffer bytes.Buffer
	enc:=json.NewEncoder(&buffer)
	if err:=enc.Encode(&sent); err!=nil {
		return InternalErr(err)
	}
	
	return buffer.String(),nil
}

func (STATELESS *ExamplePuter_correct) PutDoc() []string{
	return []string{"leia","jawa","query params are used and assumed to be a form", "body isn't used by this Put"}
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

func verifyDispatchError(T *testing.T, h Handler, errorMap map[string]int, method string) {
	for k,v := range errorMap {
		json, err := h.Dispatch( method,k, emptyMap, emptyMap, "")
		if json!="" {
			T.Errorf("expected no json result on an error : GET %s",k)
			continue
		}
		verifyErrorCode(T,err,v, fmt.Sprintf("%s %s",method, k))
	}
}

func verifyNoError(T *testing.T, json string, err *Error, msg string){
	if err!=nil || json=="" {
		T.Fatalf("didn't expect error on %s but got %+v",msg, err);
	}
}
func verifyJsonContent(T *testing.T, json string, expected string, msg string){
	if strings.TrimSpace(json)!=strings.TrimSpace(expected) {
		T.Fatalf("expected json of '%s' but got '%s' from %s",expected, json,msg);
	}
}

/*-------------------------------------------------------------------------*/
/*                                 TEST CODE                               */
/*-------------------------------------------------------------------------*/
var emptyMap = make(map[string]string)


func TestResourceMappingForIndexerFinder(T *testing.T) {
	h := NewSimpleHandler()

	h.AddExplicitResourceMethods("ox",&Ox{}, &ExampleIndexer_correct{},&ExampleFinder_correct{},nil,nil)
	h.AddExplicitResourceMethods("fart", Ox{}, nil, &ExampleFinder_correct{},nil, nil)

	json, err := h.Dispatch("GET","/ox/", emptyMap, emptyMap, "")
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

	verifyDispatchError(T, h, errorMap, "GET")

	json, err = h.Dispatch("GET","/ox/0", emptyMap, emptyMap, "")
	verifyNoError(T,json,err, "GET /ox/0")
	verifyJsonContent(T,json,"{\"Id\":0,\"IsLarge\":true}", "GET /ox/0")	
}


func TestResourceMappingForPosterAndPuter(T *testing.T) {
	h := NewSimpleHandler()

	h.AddExplicitResourceMethods("ox",&Ox{}, nil, nil, &ExamplePoster_correct{}, &ExamplePuter_correct{})
	errorMap :=map[string]int{
		"/ox/123": http.StatusBadRequest,  //can't post to a particular object (that should be PUT)
	}
	verifyDispatchError(T, h, errorMap, "POST")

	for _,size:=range []bool{false, true} {
		json, err := h.Dispatch("POST","/ox/", emptyMap, emptyMap, fmt.Sprintf("{ \"IsLarge\":%v }",size))
		verifyNoError(T,json,err, "POST /ox/")
		verifyJsonContent(T,json, fmt.Sprintf("{\"Id\":456,\"IsLarge\":%v}", size), "POST /ox/")	

		crap, err := h.Dispatch("PUT","/ox/456", emptyMap, map[string]string{"Garbage":"break json parse"}, "")
		if crap!="" {
			T.Errorf("Not expecting a return value from a Put with bad content: %s", crap)
		}
		verifyErrorCode(T, err, http.StatusBadRequest, "Bad Form Data For PUT")

		//loop to show we can make valid
		for _, i := range []string{"1st call","2nd call"} {
			json, err := h.Dispatch("PUT","/ox/456", emptyMap, map[string]string{"Id":"456","IsLarge":fmt.Sprintf("%v",!size)}, "")
			verifyNoError(T,json,err, fmt.Sprintf("PUT /ox/456 (%s)",i))
			verifyJsonContent(T,json, fmt.Sprintf("{\"Id\":456,\"IsLarge\":%v}", !size), fmt.Sprintf("PUT /ox/456 (%s)",i))	
		}
	}
}

