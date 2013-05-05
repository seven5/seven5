package seven5

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDWCDir(t *testing.T) {
	
	x := []string{"/foo", "/foo.dart", "foo.html", "foo/app", "app", "/out/foo", "/out/dart",
		"/out/foo.dart", "/foo.html.dart", "/out/.html.dart"}
	for _, path := range x {
		if isDWCTargetPath(path) {
			t.Errorf("%s should not be considered a DWC path!", path)
		}
	}
	x = []string{"/out/foo.html", "/out/bar/foo.html", "/out/app/foo.html"}
	for _, path := range x {
		if !isDWCTargetPath(path) {
			t.Errorf("%s should be considered a DWC path!", path)
		}
	}
}

func TestToDWCSource(t *testing.T) {
	
	x := []string{"/out/foo.html", "/out/bar/foo.html", "/out/app/app/app.html"}
	y := []string{"/foo.html", "/bar/foo.html", "/app/app/app.html"}
	for i, path := range x {
		if toDWCSource(path) != y[i] {
			t.Errorf("%s should have source path %s!", path, y[i])
		}
	}

}

func TestDWCTargetPath(t *testing.T) {
	
	truePath := filepath.Join(os.TempDir(), fmt.Sprintf("t%d", os.Getpid()))
	if err := os.Mkdir(truePath, os.ModeDir|0x1ED); err != nil {
		t.Fatalf("couldn't create temp dir: %s", err)
	}
	defer os.Remove(truePath)
	
	x := []string{"/foo", "/foo.dart", "foo/app", "app", "/out/foo", "/out/dart",
		"/out/foo.dart", "/foo.html.dart", "/out/.html.dart", "/out/foo.html"}
	for _, path := range x {
		if dWCTarget(path, truePath) != "" {
			t.Errorf("%s should not have target path!", path)
		}
	}

	x = []string{"/foo.html", "/foo/bar/fleazil.html", "/out.html"}
	for _, path := range x {
		if dWCTarget(path, truePath) != "" {
			t.Errorf("%s should not have target path because it does not exist!", path)
		}
		//create the file now so we can test again
		full:=filepath.Join(truePath, path)
		dir:=filepath.Dir(full)
		err:=os.MkdirAll(dir, os.ModeDir|0x1ED)
		if err!=nil {
			t.Fatal("could not create parent directory %s of test file: %v", dir, err)
		}
		f, err:=os.Create(full)
		if err!=nil {
			t.Fatalf("couldn't create test file %s:%v",full,err)
		}
		defer f.Close()
	}
	
	y := []string{"/out/foo.html", "/out/foo/bar/fleazil.html", "/out/out.html"}
	for i, path := range x {
		if dWCTarget(path, truePath) != y[i] {
			t.Errorf("%s should have target path %d %s but got '%s'!", path, i, y[i], dWCTarget(path,truePath))
		}
	}
}
