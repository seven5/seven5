package seven5

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//FileFlavor is used to "find" parts of your application at run-time. Use of these constants
//means that if the default project layout changes you only have to move your files around,
//not change your code.
type FileFlavor int

const (
	GO_SOURCE_FLAVOR = iota
	DART_FLAVOR
	ASSET_FLAVOR
	TOP_LEVEL_FLAVOR
)

//ProjectObjectFromGOPATH computes a directory inside a seven5 project that has the default
//layout.  You need to supply a project name and the directory you are looking for.  If
//you set
//
//For project foo, the default layout is:
// foo/
//    Procfile
//    .godir
//    dart/
//					foo/
//					      web/
//					      			assets/
//                pubspec.yaml
//                pubspec.lock
//                packages/
//                ...
//    db/
//    go/
//         bin/
//         pkg/
//         src/
//               foo/
//                     runfoo/
//                     				main.go

func ProjectObjectFromGOPATH(target string, projectName string, flavor FileFlavor) (string, error) {
	env := os.Getenv("GOPATH")
	if env == "" {
		return "", BAD_GOPATH
	}
	pieces := strings.Split(env, ":")
	if len(pieces) > 1 {
		env = pieces[0]
	}
	switch flavor {
	case GO_SOURCE_FLAVOR:
		return filepath.Join(env, "src", target), nil
	case DART_FLAVOR:
		return filepath.Join(filepath.Dir(env), "dart", projectName, target), nil
	case ASSET_FLAVOR:
		return filepath.Join(filepath.Dir(env), "dart", projectName, "assets", target), nil
	case TOP_LEVEL_FLAVOR:
		return filepath.Join(filepath.Dir(env), target), nil
	}
	panic("unknown type of object searched for in the project!")
}

//StaticDartContent adds an http handler for the content subdir of a dart app.  The project
//name is the subdir of 'dart' so the content is dart/projectName/web.  The prefix
//can be used if you don't want the static content mounted at '/' (the default).  IF
//you supply a prefix, it should end with /.
func StaticDartContent(h Handler, projectName string, prefix string) {
	//setup static content
	truePath, err := ProjectObjectFromGOPATH("web", projectName, DART_FLAVOR)
	if err != nil {
		log.Fatalf("can't understand GOPATH or not using default project layout: %s", err)
	}
	if prefix == "" || prefix == "/" {
		h.ServeMux().Handle("/", http.FileServer(http.Dir(truePath)))
	} else {
		//strip the path from requests so that /prefix/fart = dart/projectName/web/fart
		h.ServeMux().Handle(prefix, http.StripPrefix(prefix, http.FileServer(http.Dir(truePath))))
	}
}

//GeneratedContent adds an http handler for static content in a subdirectory.  It computes
//the generated code at the point of call, so it should be called after resources are already
//bound.
func GeneratedContent(h Handler, urlPath string) {
	desc := h.APIs()
	h.ServeMux().HandleFunc(fmt.Sprintf("%sdart", urlPath), generateDartFunc(desc))
	h.ServeMux().HandleFunc(fmt.Sprintf("%sapi/", urlPath), generateAPIDocPrinter(h))
}

//Seven5Content maps internal handlers to be inside the urlPath provided.
func Seven5Content(h Handler, urlPath string, projectName string) {
	if projectName != "" {
		h.ServeMux().HandleFunc(fmt.Sprintf("%spublicsetting/%s/", urlPath, projectName), publicSettingHandler)
	}
}

//SetIcon creates a go handler in h that will return an icon to be displayed in response to /favicon.ico.
//The binaryIcon should be an array of bytes (usually created via 'seven5tool embedfile')
func SetIcon(h Handler, binaryIcon []byte) {
	h.ServeMux().HandleFunc("/favicon.ico", generateBinPrinter(binaryIcon, "image/x-icon"))
}

func generateStringPrinter(content string, contentType string) func(http.ResponseWriter, *http.Request) {
	return generateBinPrinter([]byte(content), contentType)
}

func generateBinPrinter(content []byte, contentType string) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Add("Content-type", contentType)
		_, err := writer.Write(content)
		if err != nil {
			fmt.Printf("error writing constant binary string: %s\n", err)
		}
	}
}

func generateAPIDocPrinter(h Handler) func(writer http.ResponseWriter, req *http.Request) {
	rez := h.APIs()

	return func(writer http.ResponseWriter, req *http.Request) {
		var origin string
		if origin = req.Header.Get("Origin"); origin != "" {
			writer.Header().Add("Access-Control-Allow-Origin", "*")
			writer.Header().Add("Access-Control-Request-Method", "GET")
		}

		found := 0
		var buffer bytes.Buffer

		//look for 'api'
		pieces := strings.Split(req.URL.String(), "/")
		for i, s := range pieces {
			if s == "api" {
				found = i
				break
			}
		}
		//found it?
		if found == 0 {
			http.NotFound(writer, req)
			return
		}
		//in the right place in the sequence of pieces?  if last, make sure it is /api/
		if found == len(pieces)-1 && !strings.HasSuffix(req.URL.String(), "/") {
			http.Error(writer, "Seven5 only considers named collections with trailing slash", http.StatusNotFound)
			return
		}
		//api in bogus place in sequence
		if found != len(pieces)-1 && found != len(pieces)-2 {
			http.NotFound(writer, req)
			return
		}

		var candidate string
		result := []*APIDoc{}

		//candidate is either "" for all APIs or the 
		if found == len(pieces)-2 {
			candidate = pieces[len(pieces)-1]
		}

		//count the number
		for _, r := range rez {
			if candidate == "" || candidate == r.ResourceName {
				result = append(result, r)
			}
		}
		if len(result) == 0 {
			http.NotFound(writer, req)
			return
		}
		//encode our result
		enc := json.NewEncoder(&buffer)
		var err error
		if candidate != "" {
			err = enc.Encode(result[0])
		} else {
			err = enc.Encode(result)
		}
		//check for error
		if err != nil {
			http.Error(writer, fmt.Sprintf("Error computing json encoding: %s", err), http.StatusInternalServerError)
			return
		}

		writer.Header().Add("Content-type", "application/javascript")

		//send result
		_, err = writer.Write(buffer.Bytes())
		if err != nil {
			fmt.Printf("Whoa! Couldn't write result to other side: %s\n", err)
		}
	}
}

//generateDartFunc returns a function that outputs text string for all the dart code
//in the system.
func generateDartFunc(desc []*APIDoc) func(http.ResponseWriter, *http.Request) {

	var text bytes.Buffer
	resourceStructs := []*FieldDescription{}
	supportStructs := []*FieldDescription{}

	//generate code for known resources
	for _, d := range desc {
		text.WriteString(generateDartForResource(d))
		resourceStructs = append(resourceStructs, d.Field)
	}
	//collect up supporting structs
	for _, d := range desc {
		candidates := collectStructs(d.Field)
		for _, s := range candidates {
			if !containsType(resourceStructs, s) && !containsType(supportStructs, s) {
				supportStructs = append(supportStructs, s)
			}
		}
	}
	for _, i := range supportStructs {
		text.WriteString(generateDartForSupportStruct(i))
	}
	return generateBinPrinter(text.Bytes(), "text/plain")
}

//DefaultProjects adds the resources that we expect to be present for a typical
//seven5 project to the handler provided and returns it as as http.Handler so it can be
//use "in the normal way" with http.ServeHttp. This adds generated code to the URL
//mount point /generated.  If you call DefaultProjectBindings you don't need to worry
//about calling StaticDefaultContent or GeneratedContent.  Note that this should be called
//_after_ all resources are added to the handler as there is caching of the generated
//code.  If the project name is "" then the publicsetting.json file will not be read and
//no settings will be visible from the client side (nothing at /seven5/publicsetting)
func DefaultProjectBindings(h Handler, projectName string) http.Handler {
	StaticDartContent(h, projectName, "/")
	GeneratedContent(h, "/generated/")
	Seven5Content(h, "/seven5/", projectName)
	SetIcon(h, gopher_ico)
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
	indent := 0
	var result bytes.Buffer

	for i := 0; i < len(raw); i++ {
		c := raw[i]
		switch state {
		case WAITING_ON_NON_WS:
			if c == '\t' || c == ' ' || c == '\n' {
				continue
			}
			switch c {
			case '{':
				indent += 2
			case '}':
				indent -= 2
			}
			for j := 0; j < indent; j++ {
				result.WriteString(" ")
			}
			result.Write([]byte{c})
			state = WAITING_ON_EOL
			continue
		case WAITING_ON_EOL:
			if c == '\n' {
				result.WriteString("\n")
				state = WAITING_ON_NON_WS
				continue
			}
			switch c {
			case '{':
				indent += 2
			case '}':
				indent -= 2
			}
			result.Write([]byte{c})
			continue
		}
	}
	return result.Bytes()
}
