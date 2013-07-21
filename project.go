package seven5

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

const (
	WAITING_ON_NON_WS = iota
	WAITING_ON_EOL
)

//generateStringPrinter creates a function suitable for use with a ServeMux's handle func.  
func generateStringPrinter(content string, contentType string) func(http.ResponseWriter, *http.Request) {
	return generateBinPrinter([]byte(content), contentType)
}

//generateBinPrinter creates a function suitable for use with a ServeMux's handle func. It writes out the
//content type as a sequence bytes.
func generateBinPrinter(content []byte, contentType string) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, req *http.Request) {
		writer.Header().Add("Content-type", contentType)
		_, err := writer.Write(content)
		if err != nil {
			fmt.Printf("error writing constant binary string: %s\n", err)
		}
	}
}

//DefaultProjects adds the resources that we expect to be present for a typical
//seven5 project.  The ProjectFinder is used to find things inside the project, notably the
//static web content. Content added here is all fixed by the build of seven5 and the 
//underlying filesystem.
func DefaultProjectBindings(projectName string, pf ProjectFinder, dep DeploymentEnvironment) *ServeMux {
	mux := NewServeMux()
	WebContent(mux, projectName, "/", pf, dep.IsTest())
	SetIcon(mux, gopher_ico)
	return mux
}

//SetIcon creates a go handler in h that will return an icon to be displayed in response to /favicon.ico.
//The binaryIcon should be an array of bytes (usually created via 'seven5tool embedfile')
func SetIcon(mux *ServeMux, binaryIcon []byte) {
	mux.HandleFunc("/favicon.ico", generateBinPrinter(binaryIcon, "image/x-icon"))
}

//FileContent cause the generation of files that need refreshing on each restart.  Typically, this
//is code derived from the go structs that the application exposes. 
func FileContent(projectName string, pf ProjectFinder, holder TypeHolder, restPrefix string) {
	sourceDir := filepath.Join("web", "packages", projectName, "src")
	p, err := pf.ProjectFind(sourceDir, projectName, DART_FLAVOR)
	if err != nil {
		panic(fmt.Sprintf("can't find the place to put generated dart source! expected %s but got an error: %s",
			sourceDir, err))
	}
	outputPath := filepath.Join(p, fmt.Sprintf("%s.dart", projectName))
	file, err := os.Create(outputPath)
	if err != nil {
		panic(fmt.Sprintf("can't open dart source output path %s, got an error: %s", outputPath, err))
	}
	code := wrappedCodeGen(holder, restPrefix, projectName)
	file.Write(code.Bytes())
	file.Close()
}

//WebContent adds an http handler for the 'web' subdir of a dart app.  The project
//name is the subdir of 'dart' so the content is dart/projectName/web.  The prefix
//can be used if you don't want the static content mounted at '/' (the default if you pass ""
//as the prefix).  If you supply a prefix, it should end with /.
func WebContent(mux *ServeMux, projectName string, prefix string, pf ProjectFinder, isTestMode bool) {
	truePath, err := pf.ProjectFind("web", projectName, DART_FLAVOR)
	if err != nil {
		panic(fmt.Sprintf("can't understand GOPATH or not using default project layout: %s", err))
	}
	_, err = os.Open(truePath)
	if err != nil {
		panic(fmt.Sprintf("unable to open file resources at %s\n\tderived from your GOPATH\n", truePath))
	}
		mux.Handle(prefix, http.StripPrefix(prefix, http.FileServer(http.Dir(truePath))))
}
