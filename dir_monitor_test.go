package seven5

import (
	"bufio"
	"os"
	"path/filepath"
	"testing"
	"time" //for time.Sleep(nanosecs)
)

type TestDirectoryListener struct {
	AddedFiles   []os.FileInfo
	RemovedFiles []os.FileInfo
	ChangedFiles []os.FileInfo
}

func (listener *TestDirectoryListener) Clear() {
	listener.AddedFiles = []os.FileInfo{}
	listener.RemovedFiles = []os.FileInfo{}
	listener.ChangedFiles = []os.FileInfo{}
}
func (listener *TestDirectoryListener) FileChanged(fileInfo os.FileInfo) {
	listener.ChangedFiles = append(listener.ChangedFiles, fileInfo)
}
func (listener *TestDirectoryListener) FileRemoved(fileInfo os.FileInfo) {
	listener.RemovedFiles = append(listener.RemovedFiles, fileInfo)
}
func (listener *TestDirectoryListener) FileAdded(fileInfo os.FileInfo) {
	listener.AddedFiles = append(listener.AddedFiles, fileInfo)
}

func TestInit(t *testing.T) {

	const dirName string = "./temp-test-dir"
	monitor, err := NewDirectoryMonitor(dirName, ".foo")
	if monitor != nil {
		t.Error("Should not return a monitor for a non-existant directory")
	}
	if err == nil {
		t.Error("Should return an error for a non-existant directory")
	}

	os.Mkdir(dirName, 16877)
	defer os.RemoveAll(dirName)
	monitor, err = NewDirectoryMonitor(dirName, ".foo") //implies Poll() to set things up
	if monitor == nil {
		t.Error("Should return a monitor for a directory")
	}
	if err != nil {
		t.Error("Should not return an error for a directory")
	}

	var testListener *TestDirectoryListener = new(TestDirectoryListener)
	monitor.Listen(testListener)

	changed, err := monitor.Poll() //we haven't done anything yet
	if err != nil {
		t.Error("Should not return an error")
	}
	if changed {
		t.Error("Nothing should have changed", testListener.ChangedFiles[0].Name)
	}
	if !changed {
		checkListener(testListener,0,0,0,t)
	}

	tempFile, err := os.Create(filepath.Join(dirName, "temp-test-file.foo"))
	//don't need this, because done again later: defer os.Remove(tempFile.Name())
	tempFileInfo, err := tempFile.Stat()
	wr := bufio.NewWriter(tempFile)
	wr.WriteString("I like traffic lights.\n")
	wr.Flush()
	tempFile.Sync()
	changed, err = monitor.Poll()
	if err != nil {
		t.Error("Should not return an error")
	}
	if !changed {
		t.Error("Should have changed")
	}
	if changed {
		checkListener(testListener,1,0,0,t)
	}
	if testListener.AddedFiles[0].Name != tempFileInfo.Name {
		t.Error("Should have added the temp file")
	}
	testListener.Clear()

	time.Sleep(1000000000) /*1secs*/
	tempFile, err = os.Create(filepath.Join(dirName, "temp-test-file.foo"))
	defer os.Remove(tempFile.Name())
	wr = bufio.NewWriter(tempFile)
	wr.WriteString("But not when they are red.\n")
	wr.Flush()
	tempFile.Sync()
	//fmt.Fprintf(os.Stderr, "about to hit poll after we slept 1 secs and re-wrote the file\n")
	changed, err = monitor.Poll()
	if err != nil {
		t.Error("Should not return an error")
	}
	if !changed {
		t.Error("Should have changed")
	}
	if changed {
		checkListener(testListener,0,0,1,t)
	}
	testListener.Clear()

	changed, err = monitor.Poll()
	if err != nil {
		t.Error("Should not return an error")
	}
	if changed {
		t.Error("Should not have changed")
	}

	os.Remove(tempFile.Name())
	changed, err = monitor.Poll()
	if err != nil {
		t.Error("Should not return an error")
	}
	if !changed {
		t.Error("Should have changed")
	}
	if changed {
		checkListener(testListener,0,1,0,t)
	}
	testListener.Clear()

	tempFile, err = os.Create(filepath.Join(dirName, "temp-test-file.goo"))
	defer os.Remove(tempFile.Name())
	wr = bufio.NewWriter(tempFile)
	wr.WriteString("I like traffic lights.\n")
	wr.Flush()

	changed, err = monitor.Poll()
	if err != nil {
		t.Error("Should not return an error because the extension did not match")
	}
	if changed {
		t.Error("Should not have changed")
	}
	if !changed {
		checkListener(testListener,0,0,0,t)
	}
	testListener.Clear()

}

func checkListener(testListener *TestDirectoryListener, add, rem, chng int, t *testing.T) {
	if len(testListener.AddedFiles) != add {
		t.Error("Unexpected number of added files (expected %d, got %d)", add, len(testListener.AddedFiles))
	}
	if len(testListener.RemovedFiles) != rem {
		t.Error("Unexpected number of removed files (expected %d, got %d)", rem, len(testListener.RemovedFiles))
	}
	if len(testListener.ChangedFiles) != chng {
		t.Error("Unexpected number of changed files (expected %d, got %d)", chng, len(testListener.AddedFiles))
	}
}
