package seven5

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	prefix = "gopath"
)

//SimpleStaticFileServer is a simple implementation of a file server
//for static files that is sufficient for most simple applications.  It
//defaults to serving "/" from "static" (child of the current directory)
//but this can be changed with the environment variable STATIC_DIR.
type SimpleStaticFilesServer struct {
	testMode  bool
	mountedAt string
	staticDir string
	fs        http.Handler
}

//NewSimpleStaticFilesServer returns a new file server and if isTestMode
//is true and the environment variable GOPATH is set, it will
//also serve up go source files from the GOPATH.  It expects that the
//prefix "/gopath" will be used for gopath requests.  You should supply
//the place this has been "mounted" in the URL space (usually "/"") in
//the first parameter.
func NewSimpleStaticFilesServer(mountedAt string, isTestMode bool) *SimpleStaticFilesServer {
	staticDir := "static"
	env := os.Getenv("STATIC_DIR")
	if env != "" {
		log.Printf("STATIC_DIR is set, using  %s for static files", env)
		staticDir = env
	}
	return &SimpleStaticFilesServer{
		testMode:  isTestMode,
		mountedAt: mountedAt,
		staticDir: staticDir,
		fs:        http.StripPrefix(mountedAt, http.FileServer(http.Dir(staticDir))),
	}
}

func (s *SimpleStaticFilesServer) gopath(w http.ResponseWriter, r *http.Request, desired string) {
	gopathSegments := strings.Split(os.Getenv("GOPATH"), ":")
	for _, gopath := range gopathSegments {
		_, err := os.Stat(filepath.Join(gopath, desired))
		if err != nil {
			continue
		}
		log.Printf("serving gopath content  %s: /%v", gopath, desired)
		http.ServeFile(w, r, filepath.Join(gopath, desired))
		return
	}
	http.NotFound(w, r)
}

//ServeHTTP retuns a static file or a not found error. This function meets
//the requirement of net/http#Handler.
func (s *SimpleStaticFilesServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.testMode && strings.HasPrefix(r.URL.String(), prefix) {
		s.gopath(w, r, r.URL.String()[len(prefix):])
		return
	}
	log.Printf("[STATIC CONTENT (%s)]: %v", s.staticDir, r.URL.String())
	s.fs.ServeHTTP(w, r)
}
