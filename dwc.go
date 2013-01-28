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
	homePage   = "/home.html"
	dwcDir     = "app"
)

var hrefRE = regexp.MustCompile("href=\"([^\"]+)\"")

func checkNestedComponents(origPath string, shortPath string, truePath string) bool {
	components := generateNestedComponents(shortPath, truePath)
	for _, comp := range components {
		compPath := comp
		if !strings.HasPrefix(comp, "/") {
			dir := filepath.Dir(shortPath)
			compPath = dir + comp
		} else {
			fmt.Fprintf(os.Stderr, "WARNING: Absolute paths in link elements may crash "+
				"dart web components compiler: %s\n", compPath)
		}
		c, err := NeedsCompile(filepath.Join(truePath, shortPath), filepath.Join(truePath, compPath))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error determining compilation needed for component: %s\n", err.Error())
			return false
		}
		if c {
			fmt.Printf("Dependent component %s is newer, forcing recompilation of %s\n", compPath, origPath)
			return true
		}
		return checkNestedComponents(origPath, compPath, truePath)
	}
	return false
}

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

//IsDWCTargetPath returns true iff the path supplied corresponds to a dart web components generated
//dart file (derived from an html file _not_ in the app dir).  Note that this operates on a
//path only, not the filesystem, because the path may not exist because it needs to be
//generated.
func IsDWCTargetPath(path string) bool {
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

//ToDWCSource returns the html+template source for a given target (output) path. This assumes that the
//IsDWCPath function has already returned true.  This operates on a path, without touching the
//filesystem, because the path may not exist because it needs to be generated.
func ToDWCSource(path string) string {
	pieces := strings.Split(path, "/")
	last := pieces[len(pieces)-1]
	start := strings.Join(pieces[2:len(pieces)-1], "/")
	if start == "" {
		return "/" + last
	}
	return "/" + start + "/" + last
}

func DWCTarget(path string, truePath string) string {
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

func IsJSTarget(path string) bool {
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
		w.Header().Add("Cache-Control","no-cache, must-revalidate"); //HTTP 1.1
		w.Header().Add("Pragma","no-cache"); //HTTP 1.0
		
		//fmt.Printf("---'%v'---%v\n", r.URL.Path, IsDWCTargetPath(r.URL.Path))
		//case 3: could be a true source file that we need to convert to a DWT target
		t := DWCTarget(r.URL.Path, truePath)
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
		if IsDWCTargetPath(r.URL.Path) {
			sourceCode := ToDWCSource(r.URL.Path)
			CompileWebComponents(w, r, sourceCode, r.URL.Path, truePath, isTestMode)
			//we ran the compiler, now can let this go to completion
		}
		//case 5
		if IsJSTarget(r.URL.Path) {
			CompileJS(w, r, r.URL.Path[0:len(r.URL.Path)-3], r.URL.Path, truePath, isTestMode)
		}
		//give up and use FS
		underlyingHandler.ServeHTTP(w, r)
	})
}

func CompileJS(w http.ResponseWriter, r *http.Request,
	dartSource string, jsTarget string, truePath string, isTestMode bool) {
	dart := filepath.Join(truePath, dartSource)
	js := filepath.Join(truePath, jsTarget)

	need, err := NeedsCompile(dart, js)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !isTestMode && need {
		fmt.Fprintf(os.Stderr, "Out of date generated javascript file %s!\n", jsTarget)
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
		return
	}
	fmt.Fprintf(os.Stdout, "----------- DART2JS ----------\n%s", string(b))
	fmt.Fprintf(os.Stdout, "Compilation time: %2.2f seconds\n", t2.Sub(t1).Seconds())
	return
}

func NeedsCompile(src, dest string) (bool, error) {
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

	return NeedsCompileFiles(s, d)
}

func NeedsCompileFiles(src, dest *os.File) (bool, error) {
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

func CompileWebComponents(w http.ResponseWriter, r *http.Request,
	src string, dest string, truePath string, isTestMode bool) {

	fullSource := filepath.Join(truePath, src)
	fullDest := filepath.Join(truePath, dest)
	needCompile, err := NeedsCompile(fullSource, fullDest)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if !isTestMode && needCompile {
		fmt.Fprintf(os.Stderr, "Out of date generated dart file %s!\n", dest)
		return
	}

	if !needCompile {
		needCompile=checkNestedComponents(src, src, truePath)
	}

	if !needCompile {
		fmt.Fprintf(os.Stderr, "---------------------------\n")
		fmt.Fprintf(os.Stdout, "%s is up to date\n", dest)
		return
	}
	cmd := exec.Command("dart",
		fmt.Sprintf("--package-root=%s/packages/", truePath),
		fmt.Sprintf("%s/packages/web_ui/dwc.dart", truePath),
		fmt.Sprintf("--out=%s/%s", truePath, dwcDir),
		fullSource)
	b, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "-------- DWC ERROR -------\n%s\n", string(b))
		fmt.Fprintf(os.Stderr, "COMMAND LINE WAS: dart %s %s %s %s\n",
			fmt.Sprintf("--package-root=%s/packages/", truePath),
			fmt.Sprintf("%s/packages/web_ui/dwc.dart", truePath),
			fmt.Sprintf("--out=%s/%s", truePath, dwcDir),
			fullSource)
		return
	}
	fmt.Fprintf(os.Stderr, "----------- DWC ----------\n%s\n", string(b))
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
	dir, err := pf.ProjectFind(filepath.Join("web", "packages"), name, DART_FLAVOR)
	if err != nil {
		return err
	}
	_, err=os.Open(dir)
	if os.IsNotExist(err) {
		panic(fmt.Sprintf("No package directory for dart packages (%s): did you forget to run pub install?", dir))
	}
	buffer := wrappedCodeGen(t, pre)
	c, err := createPath(dir, "generated", "dart")
	if err != nil {
		return err
	}
	_, err = c.Write(buffer.Bytes())
	if err != nil {
		return err
	}
	c, err = createPath(dir, "seven5", "support")
	if err != nil {
		return err
	}
	_, err = c.Write([]byte(seven5_dart))
	if err != nil {
		return err
	}

	return nil
}
