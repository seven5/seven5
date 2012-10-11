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
	"bytes"
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
	h.ServeMux().HandleFunc(fmt.Sprintf("%sdart",urlPath), generateDartFunc(desc))
}

//Seven5Content maps /seven5/ to be the place where static content contained _inside_
//the seven5 can viewed such as /seven5/seven5.dart for the seven5 support library.
func Seven5Content(h Handler, urlPath string) {
	h.ServeMux().HandleFunc("/seven5/seven5.dart", 
		func (writer http.ResponseWriter, request *http.Request) {
			_, err:=writer.Write([]byte(seven5_dart))
			if err!=nil {
				fmt.Printf("error writing constant code (seven5_dart): %s\n",err)
			}
		});
}

//generateDartFunc returns a function that outputs text string for all the dart code
//in the system.
func generateDartFunc(desc []*ResourceDescription) func (http.ResponseWriter, *http.Request) {
	
	var text bytes.Buffer
	resourceStructs:= []*FieldDescription{}
	supportStructs:=[]*FieldDescription{}
	
	//generate code for known resources
	for _,d := range desc {
		text.WriteString(generateDartForResource(d))
		resourceStructs = append(resourceStructs, d.Field)
	}
	//collect up supporting structs
	for _,d := range desc {	
		candidates:=collectStructs(d.Field)
		for _, s:= range candidates {
			if !containsType(resourceStructs,s) && !containsType(supportStructs,s) {
				supportStructs = append(supportStructs, s)
			}
		}
	}
	for _,i:=range supportStructs {
		text.WriteString(generateDartForSupportStruct(i))
	}
	return func (writer http.ResponseWriter, req *http.Request) {
		_, err:=writer.Write(dartPrettyPrint(text.String()))
		if err!=nil {
			fmt.Printf("error writing generated code: %s\n",err)
		}
	}
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
  Seven5Content(h, "/seven5/")
	return h
}

const (
	WAITING_ON_NON_WS = iota
	WAITING_ON_EOL 
)
//dartPrettyPrint is a very naive dart formatter. It doesn't understand much of the lexical
//structure of dart but it's enough for our generated code (which doesn't do things like embed
//{ inside a string and does has too many, not too few, line breaks)
func dartPrettyPrint(raw string) []byte {
	
	state := WAITING_ON_NON_WS
	indent :=0
	var result bytes.Buffer
	
	for i:=0; i<len(raw); i++ {
		c := raw[i]
		switch state {
		case WAITING_ON_NON_WS:
			if c=='\t' || c==' ' || c=='\n'  {
				continue
			}
			switch c {
			case '{':
				indent+=2
			case '}':
				indent-=2
			}
			for j:=0; j<indent; j++ {
				result.WriteString(" ")
			}
			result.Write([]byte{c})
			state=WAITING_ON_EOL
			continue
		case WAITING_ON_EOL:
			if c=='\n' {
				result.WriteString("\n")
				state=WAITING_ON_NON_WS
				continue
			}
			switch c {
			case '{':
				indent+=2
			case '}':
				indent-=2
			}
			result.Write([]byte{c})
			continue
		}
	}
	return result.Bytes()
}



