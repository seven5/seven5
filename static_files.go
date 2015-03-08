package seven5

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	prefix   = "/gopath"
	CONTINUE = 50
)

//StaticFilesServer is a wrapper around http.Handler that understands about
//the Gopath for debugging.  Note that most applications with complex routing
//requirements would probably be better off using SimpleComponentMatcher because
//StaticFilesServer only understands static files, not URL->file mappings.
type StaticFilesServer interface {
	http.Handler
	Gopath(w http.ResponseWriter, r *http.Request, desired string)
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

func (s *SimpleStaticFilesServer) Gopath(w http.ResponseWriter, r *http.Request, desired string) {
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

//ServeHTTP retuns a static file or a not found error. This function meets
//the requirement of net/http#Handler.
func (s *SimpleStaticFilesServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.testMode && strings.HasPrefix(r.URL.String(), prefix) {
		s.Gopath(w, r, strings.TrimPrefix(r.URL.String(), prefix))
		return
	}
	log.Printf("[STATIC CONTENT (%s)]: %v", s.staticDir, r.URL.String())
	s.fs.ServeHTTP(w, r)
}

type StaticComponent interface {
	Page(Session, []string, bool) ComponentResult
	//part of the fixed URL space, not including preceding slash
	UrlPrefix() string
}

type ComponentResult struct {
	Status           int
	ContinueAt       string //only used if Status = CONTNUE
	ContinueConsumed int    //only used if Status = CONTINUE
	Message          string //only used if non 200s
	Redir            string //only used if status is 301
	Path             string //only used if 200
}

type ComponentMatcher interface {
	http.Handler
	FormFilepath(lang, ui, path string) string
	Match(session Session, path string) ComponentResult
}

type SimpleComponentMatcher struct {
	comp     []StaticComponent
	basedir  string
	homepage ComponentResult
	cm       CookieMapper
	sm       SessionManager
	isTest   bool
}

func NewSimpleComponentMatcher(cm CookieMapper, sm SessionManager, basedir string,
	emptyURLHandler ComponentResult, isTest bool, comp ...StaticComponent) *SimpleComponentMatcher {
	return &SimpleComponentMatcher{
		comp:     comp,
		basedir:  basedir,
		homepage: emptyURLHandler,
		sm:       sm,
		cm:       cm,
		isTest:   isTest,
	}
}

func (c *SimpleComponentMatcher) AddComponents(sc ...StaticComponent) {
	c.comp = append(c.comp, sc...)
}

func (c *SimpleComponentMatcher) FormFilepath(lang, ui, path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 1 {
		//they might have fully specified the lang and ui they want, or it might
		//be the magic lang "fixed"
		if c.isKnownLang(parts[1]) {
			return filepath.Join(c.basedir, path)
		}
	}
	return filepath.Join(c.basedir, lang, ui, path)
}

func (c *SimpleComponentMatcher) isKnownLang(l string) bool {
	return l == "en" || l == "fr" || l == "zh" || l == "fixed"
}

//
// RULE: If the path is fully qualified with a lang and a chosen ui (such as
// RULE: "en" and "mobile") then it is preserved in the path values returned.
// RULE: Otherwise, we figure that out from the request and it is passed
// RULE: in at the end when FormFilepath is called. In the case where you do
// RULE: not supply a language/ui, then none is displayed to the user by
// RULE: virtue of not being present in the returned result from this call.
//
func (c *SimpleComponentMatcher) Match(session Session, path string) ComponentResult {

	parts := strings.Split(path, "/")
	slashTerminated := strings.HasSuffix(path, "/")

	/////remove any empty segments of the path

	changed := true //starting condition
outer:
	for changed {
		changed = false //test this iteration
		for i, p := range parts {
			if p == "" {
				changed = true
				parts = append(parts[:i], parts[i+1:]...)
				continue outer
			}
		}
	}

	if len(parts) == 0 || (len(parts) == 1 && parts[0] == "") {
		return c.homepage
	}

	possibleLang := parts[0]
	dispatchable := parts
	knownLang := c.isKnownLang(possibleLang)
	soFar := ""
	// IF YOU PASS A SIMPLE PATH INCLUDING LANG AND UI WE WANT TO HONOR IT
	if knownLang {
		index := ""
		if slashTerminated {
			index = "/index.html"
		}

		if len(parts) == 1 {
			//length must equal one
			return ComponentResult{
				Status: http.StatusOK,
				Path:   "/" + parts[0] + index,
			}
		}
		dispatchable = parts[2:]

		if len(parts) == 2 {
			return ComponentResult{
				Status: http.StatusOK,
				Path:   "/" + parts[0] + "/" + parts[1] + index,
			}
		}

		//more than two parts
		soFar = "/" + filepath.Join(parts[0], parts[1])
	} else {
		//not a known lang, pick a default lang and mode
		//XXX should be doing this based on browser language setting
		//XXX should be doing this based on mobile browser or not
		soFar = "/en/web"
	}

	//soFar and dispatchable have been set already, we can start
	//the processing loop
processing:
	for {

		//we have at least one more component
		for _, component := range c.comp {
			if component.UrlPrefix() == dispatchable[0] {
				r := component.Page(session, dispatchable[1:], slashTerminated)
				switch r.Status {
				case 0:
					//nothing to do want to skip this processing and just
					//ignore this part
				case CONTINUE:
					soFar = filepath.Join(soFar, r.ContinueAt)
					dispatchable = dispatchable[r.ContinueConsumed:]
					continue processing
				case http.StatusOK:
					r.Path = soFar + r.Path
					return r
				case http.StatusMovedPermanently:
					r.Redir = soFar + r.Redir
					return r
				default:
					return r
				}
			}
		}

		//nobody wants it in the dispatching position, so just try to fetch it
		fetchable := filepath.Join(soFar, filepath.Join(dispatchable...))
		if slashTerminated {
			fetchable = fetchable + "/index.html"
		}
		return ComponentResult{
			Status: http.StatusOK,
			Path:   fetchable,
		}
	}
}

func (self *SimpleComponentMatcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var session Session

	//check for a cookie?
	id, err := self.cm.Value(r)
	if err != nil {
		if err != NO_SUCH_COOKIE {
			log.Printf("[SERVE] couldn't understand cookie (%s): %v", r.URL.Path, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		//it's a no cookie, which is not a problem
	} else {
		//we had a cookie, let's try to look it up
		rtn, err := self.sm.Find(id)
		if err != nil {
			log.Printf("[SERVE] error trying to find session (%s): %v", r.URL.Path, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		//if this is nil, there is nothing that could be found (no session or it expired)
		if rtn != nil {
			//we got a return value, is it just unique id?
			if rtn.Session == nil {
				user, err := self.sm.Generate(rtn.UniqueId)
				if err != nil {
					log.Printf("[SERVE] error trying to reconstruct session (%s): %v", r.URL.Path, err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				session, err = self.sm.Assign(rtn.UniqueId, user, time.Time{})
				if err != nil {
					log.Printf("[SERVE] error trying to assign session (%s): %v", r.URL.Path, err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			} else {
				//this means that the Find() returned a session object inside rtn
				session = rtn.Session
			}
		}
	}

	//session, if there is one, is assigned here to the correct session
	result := self.Match(session, r.URL.Path)
	if result.Status != http.StatusOK {
		if result.Status == http.StatusMovedPermanently {
			http.Redirect(w, r, result.Redir, result.Status)
		} else {
			http.Error(w, result.Message, result.Status)
		}
	} else {
		finalPath := self.FormFilepath("en", "web", result.Path)
		log.Printf("[SERVE] %+v -> %v", r.URL, finalPath)
		http.ServeFile(w, r, finalPath)
		return
	}
}
