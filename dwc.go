package seven5

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"seven5/auth"
	"strings"
)

const (
	homePage   = "/home.html"
	dwcDir     = "app"
	isTestMode = true
	suffix     = ".html"
)

func DartWebComponents(underlyingHandler http.Handler, truePath string, prefix string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//case 1: exactly /
		if r.URL.Path == prefix && r.URL.RawQuery=="" {
			http.Redirect(w, r, homePage, http.StatusTemporaryRedirect)
			return
		}
		//case 2: ends in "/"
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		//case 3: check for DWC
		candidate := filepath.Join(truePath, r.URL.Path)
		candidateDWC := filepath.Join(truePath, dwcDir, r.URL.Path+".dart")
		src, errSrc := os.Open(candidate)
		if errSrc != nil && !os.IsNotExist(errSrc) {
			http.Error(w, errSrc.Error(), http.StatusInternalServerError)
			return
		}
		dest, errTarget := os.Open(candidateDWC)
		if errTarget != nil && !os.IsNotExist(errTarget) {
			http.Error(w, errTarget.Error(), http.StatusInternalServerError)
			return
		}
		if isTestMode && src != nil {
			defer src.Close()
			if dest != nil || (dest == nil && strings.HasSuffix(candidate, suffix)) {
				if dest != nil {
					defer dest.Close()
				}
				//we don't want to loop, but we do want to recompile if needed
				if prefix+dwcDir == filepath.Dir(r.URL.Path) {
					//we are the dest file, not the source
					dest = src
					//XXX WORK ON WINDOWS?
					grandParent := filepath.Dir(filepath.Dir(r.URL.Path)) + filepath.Base(r.URL.Path)
					src, _ = os.Open(filepath.Join(truePath, grandParent))
					if src != nil {
						redir, err := NeedsCompile(src, dest)
						if err != nil {
							fmt.Fprintf(os.Stderr, "error trying to determine compilation status:%s\n", err)
							http.Error(w, err.Error(), http.StatusInternalServerError)
							return
						}
						if redir {
							http.Redirect(w, r, grandParent+"?"+r.URL.RawQuery, http.StatusTemporaryRedirect)
							return
						}
					}
					//if we reach here, we end up headed to case 4
				} else {
					err := CompileWebComponents(w, r, src, dest, candidate, filepath.Join(truePath, dwcDir), truePath)
					if err != nil {
						fmt.Fprintf(os.Stderr, "error in web component compilation:%s\n", err)
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
					return
				}
			}
		}
		//case 4
		if strings.HasSuffix(r.URL.Path, "_bootstrap.dart.js") {
			dart := filepath.Join(truePath, r.URL.Path[0:len(r.URL.Path)-3])
			src, _ := os.Open(dart)
			if isTestMode && src != nil {
				js := filepath.Join(truePath, r.URL.Path)
				dest, _ := os.Open(js)
				err:=CompileJS(dart,js,truePath,src,dest)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error trying to determine js compilation status:%s\n", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}
		//case 5: give up and use FS
		underlyingHandler.ServeHTTP(w, r)
	})
}

func CompileJS(dart string, js string, truePath string, src *os.File, dest *os.File) error {
	dart2js, err := NeedsCompile(src, dest)
	if err!=nil {
		return err
	}
	if !dart2js {
		fmt.Fprintf(os.Stderr, "-----------------------\n%s is up to date\n", js)
		return nil
	}
	
	cmd := exec.Command("dart2js",
		fmt.Sprintf("--package-root=%s/packages/", truePath),
		fmt.Sprintf("--out=%s", js),
		dart)
	b, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "-----------------------\n%s", string(b))
	if err != nil {
		return err
	}
	return nil
}

func NeedsCompile(src, dest *os.File) (bool, error) {
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
	srcF *os.File, destF *os.File, src string, outDir string, truePath string) error {
	
	compile, err := NeedsCompile(srcF, destF)
	if err != nil {
		return err
	}
	if compile {
		cmd := exec.Command("dart",
			fmt.Sprintf("--package-root=%s/packages/", truePath),
			fmt.Sprintf("%s/packages/web_ui/dwc.dart", truePath),
			fmt.Sprintf("--out=%s/%s", truePath, dwcDir), 
			src)
		b, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintf(os.Stderr, "------ DWC ERROR -------\n%s", string(b))
			return err
		}
		fmt.Fprintf(os.Stderr, "---------------------------\n%s", string(b))
		if err != nil {
			return err
		}
	} else {
		fmt.Fprintf(os.Stderr, "-----------   DWC   -------\n%s is up to date\n", src)
	}
	orig := r.URL.Path
	final := filepath.Dir(orig) + dwcDir + "/" + filepath.Base(orig)
	//XXX WORK ON WINDOWS?
	http.Redirect(w, r, final+"?"+r.URL.RawQuery, http.StatusTemporaryRedirect)
	return nil
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

func GenerateDartForJS(t TypeHolder, pre string, name string, pf auth.ProjectFinder) error {
	dir, err := pf.ProjectFind(filepath.Join("web", "packages"), name, auth.DART_FLAVOR)
	if err != nil {
		return err
	}
	fmt.Printf("generating dart code %s\n", dir)
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
