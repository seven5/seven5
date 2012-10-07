package seven5

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//defines a new error type
var BAD_GOPATH = errors.New("GOPATH is not defined or is empty")

//dumpOutput send the output to the calling client over the network.  Not used in tests,
//only when running against real network.
func DumpOutput(response http.ResponseWriter, json string, err *Error) {
	if err != nil && json != "" {
		log.Printf("ignoring json reponse (%d bytes) because also have error func %+v", len(json), err)
	}
	if err != nil {
		if err.Location != "" {
			response.Header().Add("Location", err.Location)
		}
		http.Error(response, err.Message, err.StatusCode)
		return
	}
	if _, err := response.Write([]byte(json)); err != nil {
		log.Printf("error writing json response %s", err)
	}
	return
}

//BadRequest returns an error struct representing a 402, Bad Request HTTP response. This should be returned
//when the parameters passed the by client don't make sense.
func BadRequest(msg string) (string, *Error) {
	return "", &Error{http.StatusBadRequest, "", fmt.Sprintf("BadRequest - %s ", msg)}
}

//NoContent Returns an error struct representing a "succcess" in the sense of the protocol but 
//a semantic error of "empty".
func NoContent() (string, *Error) {
	return "", &Error{http.StatusNoContent, "", "No content"}
}

//NotFound Returns a 'these are not the droids you're looking for... move along.'
func NotFound() (string, *Error) {
	return "", &Error{http.StatusNotFound, "", "Not found"}
}

//NotImplemented returns an http 501.  This happens if we find a resource at the URL _but_ the 
//implementing struct doesn't have the correct type, for example /foobars/ is a known mapping
//but the struct does not implement Indexer.
func NotImplemented() (string, *Error) {
	return "", &Error{http.StatusNotImplemented, "", "Not implemented"}
}

//InternalErr is a convenience for returning a 501 when an error has been found at the go level.
func InternalErr(err error) (string, *Error) {
	return "", &Error{http.StatusInternalServerError, "", err.Error()}
}

//ToSimpleMap converts an http level map with multiple strings as value to single string value.
//There are a number of places in HTTP (such as headers and query parameters) where this is
//possible and legal according to the spec, but still silly so we just use single valued
//values.
func ToSimpleMap(m map[string][]string) map[string]string {
	result := make(map[string]string)
	for k, v := range m {
		result[k] = strings.TrimSpace(v[0])
	}
	return result
}

//JsonResult returns a json string from the supplied value or return an error (caused by the encoder)
//via the InternalErr function.  This is the normal path for functions that return Json values.
//pretty can be set to true for pretty-printed json.
func JsonResult(v interface{}, pretty bool) (string, *Error) {
	var buff []byte
	var err error

	if pretty {
		buff, err = json.MarshalIndent(v, "", " ")
	} else {
		buff, err = json.Marshal(v)
	}
	if err != nil {
		return InternalErr(err)
	}
	result := string(buff)
	return strings.Trim(result, " "), nil
}

//ProjDirectoryFromGOPATH computes a directory inside the project level of a seven5 project
//that has the default layout.  For a project foo
// foo/
//    dart/
//    db/
//    go/
//         bin/
//         pkg/
//         src/
//               foo/
//    web/
// 
func ProjDirectoryFromGOPATH(rootDir string) (string, error) {
	env := os.Getenv("GOPATH")
	if env == "" {
		return "", BAD_GOPATH
	}
	pieces := strings.Split(env, ":")
	if len(pieces) > 1 {
		env = pieces[0]
	}
	return filepath.Join(filepath.Dir(env), rootDir), nil
}

//StaticContent adds an http handler for static content in a subdirectory
func StaticContent(h Handler, urlPath string, subdir string) {
	//setup static content
	truePath, err := ProjDirectoryFromGOPATH(subdir)
	if err != nil {
		log.Fatalf("can't understand GOPATH or not using default project layout: %s", err)
	}
	//strip the path from requests so that /urlPath/fart = modena/subdir/fart
	h.ServeMux().Handle(urlPath, http.StripPrefix(urlPath, http.FileServer(http.Dir(truePath))))
}

//GeneratedContent adds an http handler for static content in a subdirectory.  It computes
//the generated code at the point of call, so it should be called after resources are already
//bound.
func GeneratedContent(h Handler, urlPath string) {
	desc := h.Resources()
	h.ServeMux().HandleFunc(fmt.Sprintf("%sdart"), generateDartFunc(desc))
}

//generateDartFunc returns a function that outputs text string for all the dart code
//in the system.
func generateDartFunc(desc []*ResourceDescription) func (http.ResponseWriter, *http.Request) {
	
	for _,_ = range desc {
	}
	return nil
}

//DefaultProjects adds the resources that we expect to be present for a typical
//seven5 project to the handler provided and returns it as as http.Handler so it can be
//use "in the normal way" with http.ServeHttp. This adds generated code to the URL
//mount point /generated.  If you call DefaultProjectBindings you don't need to worry
//about calling StaticContent or GeneratedContent.  Note that this should be called
//_after_ all resources are added to the handler as there is caching of the generated
//code.
func DefaultProjectBindings(h Handler) http.Handler {
	StaticContent(h, "/static/", "static")
  StaticContent(h, "/dart/", "dart")
  GeneratedContent(h, "/generated/")
	return h
}
