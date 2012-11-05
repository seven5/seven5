package seven5

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

/*-------------------------------------------------------------------------*/
type Ox struct {
	Id      Id
	IsLarge Boolean
}

/*-------------------------------------------------------------------------*/
var oxList []*Ox
var oxCounter = 0

func haveOx(id Id) bool {
	for _, i := range oxList {
		if i.Id == id {
			return true
		}
	}
	return false
}

func deleteOx(index int, slice []*Ox) []*Ox {
	switch index {
	case 0:
		slice = slice[1:]
	case len(slice) - 1:
		slice = slice[:len(slice)-1]
	default:
		slice = append(slice[:index], slice[index+1:]...)
	}
	return slice
}

/*-------------------------------------------------------------------------*/
type oxIndexer struct {
}

func (STATELESS *oxIndexer) Index(headers map[string]string, queryParams map[string]string) (string, *Error) {
	if len(oxList) == 0 {
		return "[]", nil
	}
	return JsonResult(&oxList, false)
}

func (STATELESS *oxIndexer) IndexDoc() *BaseDocSet {
	return &BaseDocSet{
		`FOO`,
		`Bar`,
		`Baz`,
	}
}

/*-------------------------------------------------------------------------*/
type oxPoster struct {
}

func (STATELESS *oxPoster) Post(headers map[string]string, queryParams map[string]string,
	body string) (string, *Error) {

	var BodyOx Ox

	decoder := json.NewDecoder(strings.NewReader(body))
	err := decoder.Decode(&BodyOx)
	if err != nil {
		return BadRequest(fmt.Sprintf("Could not understand json body: %s", err))
	}
	newBoy := &Ox{Id(oxCounter), BodyOx.IsLarge}
	oxCounter++
	oxList = append(oxList, newBoy)

	//we "create" a new ox and return the object
	return JsonResult(&newBoy, false)
}

func (STATELESS *oxPoster) PostDoc() *BodyDocSet {
	return &BodyDocSet{
		`Grik`,
		`Grak`,
		`Frobnitz`,
		`fleazil`,
	}
}

/*-------------------------------------------------------------------------*/
type oxPuter struct {
}

func (STATELESS *oxPuter) Put(id Id, headers map[string]string, queryParams map[string]string,
	body string) (string, *Error) {

	//check that the id in the params matches the id supplied, if it exists
	formId, ok := queryParams["Id"]
	if ok {
		parsedId, err := ParseId(formId)
		if err != "" {
			return BadRequest(err)
		}
		if id != parsedId {
			return BadRequest(fmt.Sprintf("You supplied two different ids in the url and form: %d and %d!", id, parsedId))
		}
	}
	if !haveOx(id) {
		return NotFound()
	}

	//we are using form data here
	size, ok := queryParams["IsLarge"]
	if !ok {
		return BadRequest("You can't send a PUT to the ox without a size! Nothing to change! (Maybe try GET?)")
	}
	size = strings.ToLower(size)
	if size != "false" && size != "true" {
		return BadRequest(fmt.Sprintf("Bad Boolean value for IsLarge: %s", size))
	}
	updatedOx := Ox{id, false}
	if size == "true" {
		updatedOx.IsLarge = true
	}

	return JsonResult(&updatedOx, false)

}

func (STATELESS *oxPuter) PutDoc() *BodyDocSet {
	return &BodyDocSet{
		`leia`,
		`luke`,
		`query params are used and assumed to be form data (usually from a real html form)`,

		`body isn't used by this Puter`,
	}

}

/*-------------------------------------------------------------------------*/
type oxDeleter struct {
}

func (STATELESS *oxDeleter) Delete(id Id, headers map[string]string, queryParams map[string]string) (string, *Error) {

	if !haveOx(id) {
		return NotFound()
	}
	var result *Ox
	for i, cand := range oxList {
		if cand.Id == id {
			result = cand
			oxList = deleteOx(i, oxList)
			break
		}
	}
	return JsonResult(&result, false)
}

func (STATELESS *oxDeleter) DeleteDoc() *BaseDocSet {
	return &BaseDocSet{
		`mickey`,
		`minnie`,
		`pluto`,
	}
}

/*-------------------------------------------------------------------------*/
type oxFinder struct {
}

func (STATELESS *oxFinder) Find(id Id, headers map[string]string, queryParams map[string]string) (string, *Error) {

	if !haveOx(id) {
		return NotFound()
	}
	var result *Ox
	for _, cand := range oxList {
		if cand.Id == id {
			result = cand
			break
		}
	}
	return JsonResult(&result, false)
}

func (STATELESS *oxFinder) FindDoc() *BaseDocSet {
	return &BaseDocSet{
		"How can you lose an ox?",
		"awk",
		"sed",
	}
}

/*-------------------------------------------------------------------------*/
/*                            SHARED VERIFIERS                             */
/*-------------------------------------------------------------------------*/
func verifyErrorCode(T *testing.T, err *Error, expected int, msg string) {
	if err == nil {
		T.Errorf("expected error with code %d but got no error at all (%s)", expected, msg)
		return
	}
	if err.StatusCode != expected {
		T.Errorf("expected error code %d but got %d (%s)", expected, err.StatusCode, msg)
	}
}

func verifyDispatchError(T *testing.T, h Handler, errorMap map[string]int, method string) {
	for k, v := range errorMap {
		json, err := h.Dispatch(method, k, emptyMap, emptyMap, "")
		if json != "" {
			T.Errorf("expected no json result on an error : GET %s", k)
			continue
		}
		verifyErrorCode(T, err, v, fmt.Sprintf("%s %s", method, k))
	}
}

func verifyNoError(T *testing.T, json string, err *Error, msg string) {
	if err != nil || json == "" {
		T.Fatalf("didn't expect error on %s but got %+v", msg, err)
	}
}
func verifyJsonContent(T *testing.T, json string, expected string, msg string) {
	if strings.TrimSpace(json) != strings.TrimSpace(expected) {
		T.Fatalf("expected json of '%s' but got '%s' from %s", expected, json, msg)
	}
}

/*-------------------------------------------------------------------------*/
/*                                 TEST CODE                               */
/*-------------------------------------------------------------------------*/
var emptyMap = make(map[string]string)

func TestResourceMappingForIndexerFinder(T *testing.T) {
	h := NewSimpleHandler()

	h.AddExplicitResourceMethods("ox", &Ox{}, &oxIndexer{}, &oxFinder{}, nil, nil, nil)
	h.AddExplicitResourceMethods("fart", Ox{}, nil, &oxFinder{}, nil, nil, nil)

	json, err := h.Dispatch("GET", "/ox/", emptyMap, emptyMap, "")
	verifyNoError(T, json, err, "GET /ox/")
	verifyJsonContent(T, json, "[]", "GET /ox/")

	errorMap := map[string]int{
		"/oxen/":    http.StatusNotFound,
		"/ox/fart":  http.StatusBadRequest,
		"/oxen/123": http.StatusNotFound,
		"/ox/123":   http.StatusNotFound,       //too large an id
		"/fart/":    http.StatusNotImplemented, //name is registered but not implemented (no Finder)
		"/fart/123": http.StatusNotFound,       //to large, same as /ox/123
	}

	verifyDispatchError(T, h, errorMap, "GET")

	oxList = []*Ox{
		&Ox{0, true},
	}

	json, err = h.Dispatch("GET", "/ox/0", emptyMap, emptyMap, "")
	verifyNoError(T, json, err, "GET /ox/0")
	verifyJsonContent(T, json, "{\"Id\":0,\"IsLarge\":true}", "GET /ox/0")

	oxList = nil
}

func TestResourceMappingForPosterAndPuter(T *testing.T) {
	h := NewSimpleHandler()

	h.AddExplicitResourceMethods("ox", &Ox{}, nil, nil, &oxPoster{}, &oxPuter{}, nil)
	errorMap := map[string]int{
		"/ox/123": http.StatusBadRequest, //can't post to a particular object (that should be PUT)
	}
	verifyDispatchError(T, h, errorMap, "POST")

	ct := 0
	for _, size := range []bool{false, true} {
		json, err := h.Dispatch("POST", "/ox/", emptyMap, emptyMap, fmt.Sprintf("{ \"IsLarge\":%v }", size))
		verifyNoError(T, json, err, "POST /ox/")
		verifyJsonContent(T, json, fmt.Sprintf("{\"Id\":%d,\"IsLarge\":%v}", ct, size), "POST /ox/")

		crap, err := h.Dispatch("PUT", fmt.Sprintf("/ox/%d", ct), emptyMap,
			map[string]string{"Garbage": "break json parse"}, "")
		if crap != "" {
			T.Errorf("Not expecting a return value from a Put with bad content: %s", crap)
		}
		verifyErrorCode(T, err, http.StatusBadRequest, "Bad Form Data For PUT")

		//loop to show we can make either large or small ox
		for _, i := range []string{"1st call", "2nd call"} {
			json, err := h.Dispatch("PUT", fmt.Sprintf("/ox/%d", ct), emptyMap,
				map[string]string{"Id": fmt.Sprintf("%d", ct), "IsLarge": fmt.Sprintf("%v", !size)}, "")
			verifyNoError(T, json, err, fmt.Sprintf("PUT /ox/%d (%s)", ct, i))
			verifyJsonContent(T, json, fmt.Sprintf("{\"Id\":%d,\"IsLarge\":%v}", ct, !size),
				fmt.Sprintf("PUT /ox/%d (%s)", ct, i))
		}
		ct++
	}

	oxList = nil
}

func TestResourceMappingForDeleter(T *testing.T) {
	h := NewSimpleHandler()

	h.AddExplicitResourceMethods("ox", &Ox{}, nil, nil, nil, nil, &oxDeleter{})

	oxList = []*Ox{
		&Ox{456, true},
		&Ox{459, true},
	}

	errorMap := map[string]int{
		"/ox/":      http.StatusBadRequest, //can't del to a collection object 
		"/nope/102": http.StatusNotFound,   //can't del to a non existent resource
	}
	verifyDispatchError(T, h, errorMap, "DELETE")

	json, err := h.Dispatch("DELETE", "/ox/456", emptyMap, emptyMap, "")
	verifyNoError(T, json, err, "DELETE /ox/456")
	verifyJsonContent(T, json, "{\"Id\":456,\"IsLarge\":true}", "DELETE /ox/456")

	//check semantics!
	if len(oxList) != 1 {
		T.Errorf("Attempt to delete ox 456 failed! Ox list is not size 0!")
	}
}
