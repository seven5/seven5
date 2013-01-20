package seven5

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"seven5/auth"
)

//StaticDartContent adds an http handler for the content subdir of a dart app.  The project
//name is the subdir of 'dart' so the content is dart/projectName/web.  The prefix
//can be used if you don't want the static content mounted at '/' (the default).  If
//you supply a prefix, it should end with /.
func StaticDartContent(h Handler, projectName string, prefix string, pf auth.ProjectFinder) {
	truePath, err := pf.ProjectFind("web", projectName, auth.DART_FLAVOR)
	if err != nil {
		log.Fatalf("can't understand GOPATH or not using default project layout: %s", err)
	}
	_, err = os.Open(truePath)
	if err != nil {
		panic(fmt.Sprintf("unable to open file resources at %s\n\tderived from your GOPATH\n", truePath))
	}
	if prefix == "" || prefix == "/" {
		h.ServeMux().Handle("/", http.FileServer(http.Dir(truePath)))
	} else {
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
func Seven5Content(h Handler, urlPath string) {
	h.ServeMux().HandleFunc(fmt.Sprintf("%ssupport", urlPath),
		generateStringPrinter(seven5_dart, "text/plain"))
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

		pieces := strings.Split(req.URL.String(), "/")
		for i, s := range pieces {
			if s == "api" {
				found = i
				break
			}
		}

		if found == 0 {
			http.NotFound(writer, req)
			return
		}

		if found == len(pieces)-1 && !strings.HasSuffix(req.URL.String(), "/") {
			http.Error(writer, "Seven5 only considers named collections with trailing slash", http.StatusNotFound)
			return
		}

		if found != len(pieces)-1 && found != len(pieces)-2 {
			http.NotFound(writer, req)
			return
		}

		var candidate string
		result := []*APIDoc{}

		if found == len(pieces)-2 {
			candidate = pieces[len(pieces)-1]
		}

		for _, r := range rez {
			if candidate == "" || candidate == r.ResourceName {
				result = append(result, r)
			}
		}
		if len(result) == 0 {
			http.NotFound(writer, req)
			return
		}

		enc := json.NewEncoder(&buffer)
		var err error
		if candidate != "" {
			err = enc.Encode(result[0])
		} else {
			err = enc.Encode(result)
		}

		if err != nil {
			http.Error(writer, fmt.Sprintf("Error computing json encoding: %s", err), http.StatusInternalServerError)
			return
		}

		writer.Header().Add("Content-type", "application/javascript")

		_, err = writer.Write(buffer.Bytes())
		if err != nil {
			fmt.Printf("Whoa! Couldn't write result to other side: %s\n", err)
		}
	}
}

const LIBRARY_INFO = `
library generated;
import '/seven5/support';
import 'dart:json';
`

//generateDartFunc returns a function that outputs text string for all the dart code
//in the system.
func generateDartFunc(desc []*APIDoc) func(http.ResponseWriter, *http.Request) {
	var text bytes.Buffer
	resourceStructs := []*FieldDescription{}
	supportStructs := []*FieldDescription{}
	text.WriteString(LIBRARY_INFO)
	for _, d := range desc {
		text.WriteString(generateDartForResource(d))
		resourceStructs = append(resourceStructs, d.Field)
	}
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
//code.  
func DefaultProjectBindings(h Handler, projectName string, pf auth.ProjectFinder) http.Handler {
	StaticDartContent(h, projectName, "/", pf)
	GeneratedContent(h, "/generated/")
	Seven5Content(h, "/seven5/")
	SetIcon(h, gopher_ico)
	return h
}

const (
	WAITING_ON_NON_WS	= iota
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
