package seven5

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	GOPATH_PREFIX = "/gopath"
	CONTINUE      = 50
)

//StaticFilesServer is a wrapper around http.Handler that understands about
//the Gopath for debugging.  Note that most applications with complex routing
//requirements would probably be better off using SimpleComponentMatcher because
//StaticFilesServer only understands static files, not URL->file mappings.
type StaticFilesServer interface {
	http.Handler
}

//SimpleStaticFileServer is a simple implementation of a file server
//for static files that is sufficient for most applications.  It
//defaults to serving "/" from "static" (child of the current directory)
//but this can be changed with the environment variable STATIC_DIR.
type SimpleStaticFilesServer struct {
	testMode  bool
	mountedAt string
	staticDir string
	fs        http.Handler
}

//NewStaticFilesServer returns a new file server and if isTestMode
//is true and the environment variable GOPATH is set, it will
//also serve up go source files from the GOPATH.  It expects that the
//prefix "/gopath" will be used for gopath requests.  You should supply
//the place this has been "mounted" in the URL space (usually "/"") in
//the first parameter.
func NewStaticFilesServer(mountedAt string, isTestMode bool) *SimpleStaticFilesServer {
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

//Utility function for taking a requested URL and finding it in the Gopath
//and returning its contents.  It will return 404 if the file cannot be found
//and expects that the desired parameter already has had the gopath prefix
//stripped of it, e.g. /foo.go not /gopath/foo.go.
func GopathLookup(w http.ResponseWriter, r *http.Request, desired string) {
	gopathSegments := strings.Split(os.Getenv("GOPATH"), ":")
	for _, gopath := range gopathSegments {
		_, err := os.Stat(filepath.Join(gopath, desired))
		if err != nil {
			continue
		}
		log.Printf("[GOPATH CONTENT] %s", filepath.Join(gopath, desired))
		http.ServeFile(w, r, filepath.Join(gopath, desired))
		return
	}
	http.NotFound(w, r)
}

//Return the path to the file that has the desired content, or return "" if
//the file cannot be found.
func GopathSearch(desired string) string {
	gopathSegments := strings.Split(os.Getenv("GOPATH"), ":")
	for _, gopath := range gopathSegments {
		_, err := os.Stat(filepath.Join(gopath, "src", desired))
		if err != nil {
			continue
		}
		return filepath.Join(gopath, "src", desired)
	}
	return ""
}

//ServeHTTP retuns a static file or a not found error. This function meets
//the requirement of net/http#Handler.
func (s *SimpleStaticFilesServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.testMode && strings.HasPrefix(r.URL.String(), GOPATH_PREFIX) {
		GopathLookup(w, r, strings.TrimPrefix(r.URL.String(), GOPATH_PREFIX))
		return
	}
	log.Printf("[STATIC CONTENT (%s)]: %v", s.staticDir, r.URL.String())
	s.fs.ServeHTTP(w, r)
}
