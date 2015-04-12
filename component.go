package seven5

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

//SimpleIdComponent is designed to allow urls like /foo/1 to work.  IdComponent serves
//the static file /foo/view.html for the requested urls of the form
// /foo/123,/foo/123/, and /foo/123/view.html.  If the user has a current
//session and is an admin urls like /foo/123/edit will resolve to the file
// /foo/edit.html.  Given a url of the form /foo/ it will try to serve
// /foo/index.html, but this (or any of the resulting files) can result in a
// 404 when trying to serve the static content.
type SimpleIdComponent struct {
	singular string
	exists   ExistsCheck
	viewedit ViewEditCheck
	newcheck NewCheck
}

//NewCheck is a function that is called by SimpleComponentMatcher if it is
//acceptable to allow the current user to view the page /singular/new or
// /singluar/new.html. Like ViewEditCheck, this function should return  errors
//in the second parameter and true in the first parameter only if it is
//ok to continue processing.
type NewCheck func(pb PBundle) (bool, error)

//ExistsCheck is a function that is called by SimpleComponentMatcher to test
//if a given id is valid. It should return false, error in case of an error
//false,nil to indicate the id was not found and true, nil to indicate that
//processing can continue.
type ExistsCheck func(pb PBundle, id int64) (bool, error)

//ViewEditCheck is a function called by SimpleComponentMatcher to determine if
//a given user (represented by the PBundle from their client) can even see the
//static file for the given id.  Note that this prevents the entire page from
//loading, independent from what the rest resources would do in the face of
//such a request.  This should return false, err for an error, false, nil
//to refuse access, and true, nil to allow access.  isView is set to true
//if the url is of the form /foo/123/view or /foo/123/view.html, otherwise
//isView is false and the url is of the form /foo/123/edit or /foo/123/edit.html.
type ViewEditCheck func(pb PBundle, id int64, isView bool) (bool, error)

//NewSimpleIdComponent creates a simple id (int64 type) based implementation of
//StaticComponent.  The caller may choose to provide either or both of the
//check functions to do access control during URL evaluation.  The url's processed
//are of the form /singular/123.  If any of the check functions are not supplied
//it is assumed that access is ok (and it can still be disallowed in the rest
//apis later).
func NewSimpleIdComponent(singular string, exists ExistsCheck, newcheck NewCheck, viewedit ViewEditCheck) *SimpleIdComponent {
	return &SimpleIdComponent{
		singular,
		exists,
		viewedit,
		newcheck,
	}
}

//IndexOnlyComponent returns the page /plural/index.html for the requested
//url /plural, /plural/, or /plural/index.html.  Note that this can still 404
//if the actual file is not present.
type IndexOnlyComponent struct {
	plural    string
	indexPath string
}

//NewIndexOnlyComponent returns a StaticComponent that understands exacly three
//urls.  The urls understood is /plural, /plural/, and /plural/index.html and
//all of these return the file provided in path.  For convenience, it is often
//useful to map /plural/index.html to /singular/index.html so that all the files
//for a given type are in the same directory, rather than having a single file
//in the plural directory.
func NewIndexOnlyComponent(plural string, path string) *IndexOnlyComponent {
	return &IndexOnlyComponent{plural, path}
}

//Page does the work of turning a path plus a PBundle into a ComponentResult. That
//ComponentResult might have Status==CONTINUE.
func (self *SimpleIdComponent) Page(pb PBundle, path []string, trailingSlash bool) ComponentResult {

	//allow singular to be used as synonym for plural
	if len(path) == 0 {
		return ComponentResult{
			Path:   "/" + filepath.Join(self.UrlPrefix(), "index.html"),
			Status: http.StatusOK,
		}
	}

	if (len(path) == 1 && path[0] == "new") || (len(path) == 1 && path[0] == "new.html") {
		if self.newcheck != nil {
			ok, err := self.newcheck(pb)
			if err != nil {
				return ComponentResult{
					Status:  http.StatusInternalServerError,
					Message: err.Error(),
				}
			}
			if !ok {
				return ComponentResult{
					Message: "you lose",
					Status:  http.StatusUnauthorized,
				}
			}
		}
		return ComponentResult{
			Path:   "/" + filepath.Join(self.UrlPrefix(), "new.html"),
			Status: http.StatusOK,
		}
	}

	if (len(path) == 1 && path[0] == "index") || (len(path) == 1 && path[0] == "index.html") {
		return ComponentResult{
			Path:   "/" + filepath.Join(self.UrlPrefix(), "index.html"),
			Status: http.StatusOK,
		}
	}

	// parse the id in the URL
	rawId := path[0]
	id, err := strconv.ParseInt(rawId, 10, 64)
	if err != nil {
		//handle the case of /post/new.js
		return ComponentResult{
			Status:           CONTINUE,
			ContinueAt:       "post",
			ContinueConsumed: 1,
		}
	}
	if self.exists != nil {
		ok, err := self.exists(pb, id)
		if err != nil {
			return ComponentResult{
				Status:  http.StatusInternalServerError,
				Message: err.Error(),
			}
		}
		if !ok {
			return ComponentResult{
				Status:  http.StatusNotFound,
				Message: fmt.Sprintf("could not find %d", id),
			}
		}
	}

	// show view.html for /foo/1 or /foo/1/
	if len(path) == 1 {
		if self.viewedit != nil {
			ok, err := self.viewedit(pb, id, true)
			if err != nil {
				return ComponentResult{
					Status:  http.StatusInternalServerError,
					Message: err.Error(),
				}
			}
			if !ok {
				return ComponentResult{
					Message: "you lose",
					Status:  http.StatusUnauthorized,
				}
			}
		}
		return ComponentResult{
			Path:   "/" + filepath.Join(self.UrlPrefix(), "view.html"),
			Status: http.StatusOK,
		}
	}

	// see if it's a verb we understand
	if len(path) == 2 {
		verb := path[1]
		switch verb {
		case "edit", "edit.html", "view", "view.html":
			if !strings.HasSuffix(verb, ".html") {
				verb = verb + ".html"
			}
			if self.viewedit != nil {
				ok, err := self.viewedit(pb, id, verb == "view.html")
				if err != nil {
					return ComponentResult{
						Status:  http.StatusInternalServerError,
						Message: err.Error(),
					}
				}
				if !ok {
					return ComponentResult{
						Message: "you lose",
						Status:  http.StatusUnauthorized,
					}
				}
				return ComponentResult{
					Path:   "/" + filepath.Join(self.UrlPrefix(), verb),
					Status: http.StatusOK,
				}

			}
		default:
			return ComponentResult{
				Status:           CONTINUE,
				ContinueAt:       self.UrlPrefix(),
				ContinueConsumed: 2,
			}
		}
	}

	//we don't accept any more parts to the URL
	return ComponentResult{
		Status:  http.StatusBadRequest,
		Message: fmt.Sprintf("Unacceptable url for %s", self.singular),
	}
}

func (self *SimpleIdComponent) UrlPrefix() string {
	return self.singular
}

//Page does the work of turning a path plus a PBundle into a ComponentResult. That
//ComponentResult might have Status==CONTINUE for a url like /plural/index.css.
func (self *IndexOnlyComponent) Page(pb PBundle, path []string, trailingSlash bool) ComponentResult {
	//allow singular to be used as synonym for plural
	if len(path) == 0 {
		return ComponentResult{
			Path:   "/" + self.indexPath,
			Status: http.StatusOK,
		}
	}

	if len(path) == 1 && path[0] == "index.html" {
		return ComponentResult{
			Path:   "/" + self.indexPath,
			Status: http.StatusOK,
		}
	}

	//just try to serve up the content
	return ComponentResult{
		Status:           CONTINUE,
		ContinueAt:       "post", //SINGULAR
		ContinueConsumed: 1,
	}

}

func (self *IndexOnlyComponent) UrlPrefix() string {
	return self.plural
}

type StaticComponent interface {
	Page(PBundle, []string, bool) ComponentResult
	//part of the fixed URL space, not including preceding slash
	UrlPrefix() string
}

//ComponentResult is produced when we try to match a URL provided by the client
//to a fixed file.  If the Status is 200, then server should *try* to serve
//the file at Path. Note that this file may not exists and the final result
//would thus be a 404 back to the client.  If the status is 301 (MovedPermanently)
//the url (path) given by Redir is sent to client.  If the Status is CONTINUE
//the processing of the URL can continue.  This continuation allows for part
//of the URL to be consumed by one component while allowing other components
//to still receive the later portions.  If the Status is anything else, it
//is sent to the client, along with the Message.
type ComponentResult struct {
	Status           int
	ContinueAt       string //only used if Status = CONTNUE
	ContinueConsumed int    //only used if Status = CONTINUE
	Message          string //only used if non 200s
	Redir            string //only used if status is 301
	Path             string //only used if 200
}

//ComponentMatcher takes a requested URL and converts it to a static filename.
//Because the semantics of the transformation may involve things other than
//URL path (such as current session), Match is called with the fully constructed
//PBundle.  Note that this processing (in Match) usually stateless in that
//the _exact_ URL is typically not checked; a request for /foo/123 might result
//in the static file /foo/view.html even though no foo exists with id 123.
type ComponentMatcher interface {
	http.Handler
	FormFilepath(lang, ui, path string) string
	Match(pb PBundle, path string) ComponentResult
}

//SimpleComponentMatcher is an implementation of a ComponentMatcher that serves
//us static files from a specific directory.
type SimpleComponentMatcher struct {
	comp     []StaticComponent
	basedir  string
	homepage ComponentResult
	cm       CookieMapper
	sm       SessionManager
	isTest   bool
}

//NewSimpleComponentMatcher takes any number of StaticComponent objects and
//uses these to parse the URLs provided.  When a URL is successfully parsed
//it returns a static file from basedir (or a 404 if its not there).  The
//CookieMapper and SessionManager are needed because some StaticComponents
//need to have the full PBundle to do their work--and thus we must be able to
//decode the current session from cookies sent by the browser.  If isTest is
//true we will also attempt to serve content from the GOPATH to allow more
//convenient debugging in a browser.  The emptyURLHandler is used as the result
//of the user requesting the url "/".
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

//AddComponent adds any number of StaticComponents to this matcher.
func (c *SimpleComponentMatcher) AddComponents(sc ...StaticComponent) {
	c.comp = append(c.comp, sc...)
}

//FormFilepath is used to convert a url like "/foo/123/view.html" into
//"/en/web/foo/123/view.html" for processing in the filesystem.  If the lang
//and ui are already present in the path, such as /fr/mobile/foo/123/view.html,
//they are honored and the url is not changed. Otherwise we join the lang
//and ui onto the front of the path before servicing the request for a file.
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

//XXX fix me
func (c *SimpleComponentMatcher) isKnownLang(l string) bool {
	return l == "en" || l == "fr" || l == "zh" || l == "fixed"
}

//
// Match takes in a path and a PBundle, derived from a client request,
// and returns a ComponentResult which will be either a redirect, a path
// to a file (that may or may not exist), or some other error.  Note that
// this return value never has Status==CONTINUE because that is only used
// during the internal processing inside this function.
//
func (c *SimpleComponentMatcher) Match(pb PBundle, path string) ComponentResult {

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
	}

	//soFar and dispatchable have been set already, we can start
	//the processing loop
processing:
	for {

		//we have at least one more component
		for _, component := range c.comp {
			if component.UrlPrefix() == dispatchable[0] {
				r := component.Page(pb, dispatchable[1:], slashTerminated)
				switch r.Status {
				case 0:
					//nothing to do want to skip this processing and just
					//ignore this part
				case CONTINUE:
					soFar = "/" + filepath.Join(soFar, r.ContinueAt)
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
		if soFar == "" {
			soFar = "/"
		}
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

//ServeHTTP makes SimpleComponentMatcher meet the interface http.Handler.
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
				sd, err := self.sm.Generate(rtn.UniqueId)
				if err != nil {
					log.Printf("[SERVE] error trying to reconstruct session (%s): %v", r.URL.Path, err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				session, err = self.sm.Assign(rtn.UniqueId, sd, time.Time{})
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
	pbundle, err := NewSimplePBundle(r, session, self.sm)
	if err != nil {
		log.Printf("[SERVE] error trying to create parameter bundle (%s): %v", r.URL.Path, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result := self.Match(pbundle, r.URL.Path)
	if result.Status != http.StatusOK {
		if result.Status == http.StatusMovedPermanently {
			log.Printf("[REDIR] %+v -> %v", r.URL, result.Redir)
			http.Redirect(w, r, result.Redir, result.Status)
		} else {
			log.Printf("[ERROR] %+v -> %d %v", r.URL, result.Status, result.Message)
			http.Error(w, result.Message, result.Status)
		}
	} else {
		if self.isTest && strings.HasPrefix(r.URL.String(), GOPATH_PREFIX) {
			GopathLookup(w, r, strings.TrimPrefix(r.URL.String(), GOPATH_PREFIX))
			return
		}
		finalPath := self.FormFilepath("en", "web", result.Path)
		if self.isTest {
			path := GopathSearch(result.Path)
			if path != "" {
				log.Printf("[GOPATH] %v -> %v", r.URL, path)
				http.ServeFile(w, r, path)
				return
			}
		}
		log.Printf("[SERVE] %+v -> %v", r.URL, finalPath)
		http.ServeFile(w, r, finalPath)
		return
	}
}
