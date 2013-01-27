package seven5

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDWCDir(t *testing.T) {

	x := []string{"/foo", "/foo.dart", "foo.html", "foo/app", "app", "/app/foo", "/app/dart",
		"/app/foo.dart", "/foo.html.dart", "/app/.html.dart"}
	for _, path := range x {
		if IsDWCTargetPath(path) {
			t.Errorf("%s should not be considered a DWC path!", path)
		}
	}
	x = []string{"/app/foo.html", "/app/bar/foo.html", "/app/app/foo.html"}
	for _, path := range x {
		if !IsDWCTargetPath(path) {
			t.Errorf("%s should be considered a DWC path!", path)
		}
	}
}

func TestToDWCSource(t *testing.T) {
	x := []string{"/app/foo.html", "/app/bar/foo.html", "/app/app/app/app.html"}
	y := []string{"/foo.html", "/bar/foo.html", "/app/app/app.html"}
	for i, path := range x {
		if ToDWCSource(path) != y[i] {
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
	
	x := []string{"/foo", "/foo.dart", "foo/app", "app", "/app/foo", "/app/dart",
		"/app/foo.dart", "/foo.html.dart", "/app/.html.dart", "/app/foo.html"}
	for _, path := range x {
		if DWCTarget(path, truePath) != "" {
			t.Errorf("%s should not have target path!", path)
		}
	}

	x = []string{"/foo.html", "/foo/bar/fleazil.html", "/app.html"}
	for _, path := range x {
		if DWCTarget(path, truePath) != "" {
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
	
	y := []string{"/app/foo.html", "/app/foo/bar/fleazil.html", "/app/app.html"}
	for i, path := range x {
		if DWCTarget(path, truePath) != y[i] {
			t.Errorf("%s should have target path %d %s but got '%s'!", path, i, y[i], DWCTarget(path,truePath))
		}
	}
}
