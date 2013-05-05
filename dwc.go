package seven5

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	homePage = "/home.html"
	dwcDir   = "out"
)

var hrefRE = regexp.MustCompile("href=\"([^\"]+)\"")

//checkNestedComponents is a recursive routine to determine if a given path contains
//referencs to other components that are newer.
func checkNestedComponents(origPath string, shortPath string, truePath string) bool {
	components := generateNestedComponents(shortPath, truePath)
	recompile:=false
	for _, comp := range components {
		compPath := comp
		if !strings.HasPrefix(comp, "/") {
			dir := filepath.Dir(shortPath)
			compPath = filepath.Join(dir,comp)
		} else {
			fmt.Fprintf(os.Stderr, "WARNING: Absolute paths in link elements may crash "+
				"dart web components compiler: %s\n", compPath)
		}
		targ := dWCTarget(compPath, truePath)
		c, err := needsCompile(filepath.Join(truePath, compPath), filepath.Join(truePath, targ))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error determining compilation needed for component source: %s\n", err.Error())
			return false
		}
		if !c {
			c, err = needsCompile(filepath.Join(truePath, shortPath), filepath.Join(truePath, targ))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error determining compilation needed for component: %s\n", err.Error())
				return false
			}
		} else {
			fmt.Fprintf(os.Stdout, "Dependent component %s out of date (compared to %s)\n", compPath, targ)
			dwc(compPath, targ, truePath)
			recompile=true;
			break
		}
		if c {
			fmt.Fprintf(os.Stdout, "----------- DWC ----------\n")
			fmt.Fprintf(os.Stdout, "Dependent component %s is newer, forcing recompilation of %s\n", targ, origPath)
			recompile=true
			break
		}
		//recurse into component
		if checkNestedComponents(origPath, compPath, truePath) {
			recompile=true
			break
		}
	}
	return recompile
}

//generateNestedComponents looks for this type of string via a regular expression
// <link rel="components" href="blah"...> and returns a slice of the href values.
func generateNestedComponents(shortPath string, truePath string) []string {
	fullPath := filepath.Join(truePath, shortPath)
	f, err := os.Open(fullPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "can't open file to look for linked components:%v\n", err)
		return []string{}
	}
	defer f.Close()
	r := bufio.NewReader(f)
	collection := []string{}
	for {
		line, err := r.ReadString(10) // 0x0A separator = newline
		if err == io.EOF {
			// do something here
			break
		} else if err != nil {
			fmt.Fprintf(os.Stderr, "error searching for linked components:%v\n", err)
			return []string{}
		}
		if strings.Index(line, "<link") != -1 {
			if strings.Index(line, "rel=\"components\"") != -1 {
				matches := hrefRE.FindStringSubmatch(line)
				if len(matches) == 0 {
					fmt.Fprintf(os.Stderr, "Unable to understand component link: %s\n", line)
					continue
				}
				collection = append(collection, matches[1])
			}
		}
	}
	return collection
}

//isDWCTargetPath returns true iff the path supplied corresponds to a dart web components generated
//dart file (derived from an html file _not_ in the app dir).  Note that this operates on a
//path only, not the filesystem, because the path may not exist because it needs to be
//generated.
func isDWCTargetPath(path string) bool {
	if !strings.HasPrefix(path, "/"+dwcDir) {
		return false
	}
	pieces := strings.Split(path, "/")
	if len(pieces) < 3 {
		return false
	}
	if !strings.HasSuffix(path, ".html") {
		return false
	}
	last := pieces[len(pieces)-1]
	if last == ".html" {
		return false
	}
	return true
}

//toDWCSource returns the html+template source for a given target (output) path. This assumes that the
//IsDWCPath function has already returned true.  This operates on a path, without touching the
//filesystem, because the path may not exist because it needs to be generated.
func toDWCSource(path string) string {
	pieces := strings.Split(path, "/")
	last := pieces[len(pieces)-1]
	start := strings.Join(pieces[2:len(pieces)-1], "/")
	if start == "" {
		return "/" + last
	}
	return "/" + start + "/" + last
}

//dWCTarget takes a path plus a piece of the filesystem and returns true if that
//path is present inside the truePath tree and if the path makes sense as a target
//file of DWC.
func dWCTarget(path string, truePath string) string {
	if !strings.HasSuffix(path, ".html") {
		return ""
	}
	pieces := strings.Split(path, "/")
	if len(pieces) > 2 {
		if pieces[1] == dwcDir {
			return ""
		}
	}
	f, err := os.Open(filepath.Join(truePath, path))
	if err != nil {
		return ""
	}
	defer f.Close()
	return "/" + dwcDir + path
}

//isJSTarget returns true if the path is one that should trigger a JS compilation
//from dart source.
func isJSTarget(path string) bool {
	return strings.HasSuffix(path, "_bootstrap.dart.js")
}

func DartWebComponents(underlyingHandler http.Handler, truePath string, prefix string, isTestMode bool) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//case 1: exactly /
		if r.URL.Path == prefix && r.URL.RawQuery == "" {
			http.Redirect(w, r, homePage, http.StatusTemporaryRedirect)
			return
		}
		//case 2: ends in "/"
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		w.Header().Add("Cache-Control", "no-cache, must-revalidate") //HTTP 1.1
		w.Header().Add("Pragma", "no-cache")                         //HTTP 1.0

		//fmt.Printf("---'%v'---%v\n", r.URL.Path, IsDWCTargetPath(r.URL.Path))
		//case 3: could be a true source file that we need to convert to a DWT target
		
		t := dWCTarget(r.URL.Path, truePath)
		if t != "" {
			//we have a target, need to do a redir to force compilation
			suffix := r.URL.Query().Encode()
			if suffix != "" {
				suffix = "?" + suffix
			}
			http.Redirect(w, r, t+suffix, http.StatusFound)
			return
		}
		//case 4: check for DWC TARGET being passed to us.... note that this may not
		//exist at this point
		if isDWCTargetPath(r.URL.Path) {
			sourceCode := toDWCSource(r.URL.Path)
			compileWebComponents(w, r, sourceCode, r.URL.Path, truePath, isTestMode)
			//we ran the compiler, now can let this go to completion
		}
		//case 5
		if isJSTarget(r.URL.Path) {
			compileJS(w, r, r.URL.Path[0:len(r.URL.Path)-3], r.URL.Path, truePath, isTestMode)
		}
		
		//fmt.Printf("---> %v\n", r.URL)
		//give up and use FS
		underlyingHandler.ServeHTTP(w, r)
	})
}

//compileJS runs the dart2js compiler.
func compileJS(w http.ResponseWriter, r *http.Request,
	dartSource string, jsTarget string, truePath string, isTestMode bool) {
	dart := filepath.Join(truePath, dartSource)
	js := filepath.Join(truePath, jsTarget)

	need, err := needsCompile(dart, js)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//fmt.Printf("isTestMode? %v %v, '%s'->'%s'\n",isTestMode, need, dart,js)
	if !isTestMode && need {
		fmt.Fprintf(os.Stderr, "Out of date generated javascript file %s!\n", jsTarget)
		return
	}
	if !need {
		fmt.Fprintf(os.Stdout, "----------- DART2JS ----------\n")
		fmt.Fprintf(os.Stdout, "%s is up to date\n", jsTarget)
		return
	}

	t1 := time.Now()
	cmd := exec.Command("dart2js",
		"--minify",
		fmt.Sprintf("--package-root=%s/packages/", truePath),
		fmt.Sprintf("--out=%s", js),
		dart)
	b, err := cmd.CombinedOutput()
	t2 := time.Now()
	if err != nil {
		fmt.Fprintf(os.Stdout, "-------- DART2JS ERROR -------\n%s", string(b))
		fmt.Fprintf(os.Stdout, "Dart source was %s\n", dartSource)
		return
	}
	fmt.Fprintf(os.Stdout, "----------- DART2JS ----------\n%s", string(b))
	fmt.Fprintf(os.Stdout, "Compilation time: %2.2f seconds on %s\n", t2.Sub(t1).Seconds(), dartSource)
	return
}

func needsCompile(src, dest string) (bool, error) {
	d, errDest := os.Open(dest)
	if errDest != nil && !os.IsNotExist(errDest) {
		return false, errDest
	}
	if d != nil {
		defer d.Close()
	}
	s, errSrc := os.Open(src)
	if errSrc != nil {
		if !os.IsNotExist(errSrc) {
			return false, nil
		}
		return false, errSrc
	}
	defer s.Close()

	return needsCompileFiles(s, d)
}

func needsCompileFiles(src, dest *os.File) (bool, error) {
	if dest == nil {
		return true, nil
	}
	s, err := src.Stat()
	if err != nil {
		return false, err
	}
	d, err := dest.Stat()
	if err != nil {
		return false, err
	}
	return d.ModTime().Before(s.ModTime()), nil
}

func compileWebComponents(w http.ResponseWriter, r *http.Request,
	src string, dest string, truePath string, isTestMode bool) {

	fullSource := filepath.Join(truePath, src)
	fullDest := filepath.Join(truePath, dest)
	needCompile, err := needsCompile(fullSource, fullDest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !isTestMode && needCompile {
		fmt.Fprintf(os.Stderr, "Out of date generated dart file %s! %s,%s\n", dest, fullSource, fullDest)
		return
	}

	if !needCompile {
		needCompile = checkNestedComponents(src, src, truePath)
	}

	if !needCompile {
		fmt.Fprintf(os.Stderr, "-----------  DWC  ----------\n")
		fmt.Fprintf(os.Stdout, "%s is up to date\n", dest)
		return
	}
	dwc(src,dest,truePath)
}

//dwc actually runs the compiler and prints output to the terminal
func dwc(src string, dest string, truePath string) {
	fullSource := filepath.Join(truePath, src)
	
	packageParent := filepath.Dir(truePath)
	
	cmd := exec.Command("dart",
		fmt.Sprintf("--package-root=%s/packages/", packageParent),
		fmt.Sprintf("%s/packages/web_ui/dwc.dart", packageParent),
		fmt.Sprintf("--out=%s/%s", truePath, dwcDir),
		fullSource)
	b, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "-------- DWC ERROR -------\n%s", string(b))
		fmt.Fprintf(os.Stderr, "COMMAND LINE WAS: dart %s %s %s %s\n",
			fmt.Sprintf("--package-root=%s/packages/", packageParent),
			fmt.Sprintf("%s/packages/web_ui/dwc.dart", packageParent),
			fmt.Sprintf("--out=%s/%s", truePath, dwcDir),
			fullSource)
		return
	}
	fmt.Fprintf(os.Stderr, "----------- DWC ----------\n%s", string(b))
	return
}

func createPath(parent string, d string, f string) (*os.File, error) {
	dir := filepath.Join(parent, d)
	x, err := os.Open(dir)
	if x == nil {
		err := os.Mkdir(dir, os.ModeDir|0x1ED)
		if err != nil {
			return nil, err
		}
	} else {
		fi, err := x.Stat()
		if err != nil {
			return nil, err
		}
		if !fi.IsDir() {
			return nil, errors.New(fmt.Sprintf("%s exists and is not a directory!", dir))
		}
	}
	file := filepath.Join(dir, f)
	cr, err := os.Create(file)
	if err != nil {
		return nil, err
	}
	return cr, nil
}

//GenerateDartForWireTypes emits dart source code that allows the client side dart code
//to manipulate the defined types (in TypeHolder argument) conveniently.  It uses
//the projectfinder supplied to know where to place the resulting file.
func GenerateDartForWireTypes(t TypeHolder, pre string, name string, pf ProjectFinder) error {
	dir, err := pf.ProjectFind("lib", name, DART_FLAVOR)
	if err != nil {
		return err
	}
	_, err = os.Open(dir)
	if os.IsNotExist(err) {
		panic(fmt.Sprintf("No package directory for dart packages (%s): did you forget to run pub install?", dir))
	}
	fmt.Printf("seven5: generating dart code to %s\n", dir)
	buffer := wrappedCodeGen(t, pre)
	c, err := createPath(dir, "generated", fmt.Sprintf("%s.dart",name))
	if err != nil {
		return err
	}
	_, err = c.Write(buffer.Bytes())
	if err != nil {
		return err
	}
	c, err = createPath(dir, "seven5", "support.dart")
	if err != nil {
		return err
	}
	_, err = c.Write([]byte(seven5_dart))
	if err != nil {
		return err
	}

	return nil
}
