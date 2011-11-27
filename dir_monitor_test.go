package seven5

import (
	"os"
	"fmt"
	"bufio"
	"testing"
	//"path/filepath"
)

type TestDirectoryListener struct {
	AddedFiles []os.FileInfo
	RemovedFiles []os.FileInfo
	ChangedFiles []os.FileInfo
}
func (listener *TestDirectoryListener) Clear(){
	listener.AddedFiles = []os.FileInfo{}
	listener.RemovedFiles = []os.FileInfo{}
	listener.ChangedFiles = []os.FileInfo{}
}
func (listener *TestDirectoryListener) FileChanged(fileInfo os.FileInfo){
	listener.ChangedFiles = append(listener.AddedFiles, fileInfo)
}
func (listener *TestDirectoryListener) FileRemoved(fileInfo os.FileInfo){
	listener.RemovedFiles = append(listener.AddedFiles, fileInfo)
}
func (listener *TestDirectoryListener) FileAdded(fileInfo os.FileInfo){
	listener.AddedFiles = append(listener.AddedFiles, fileInfo)
}

func TestInit(t *testing.T) {
	
	info, _ := os.Stat("./")
	fmt.Printf("PWD mode %v\n", info.Mode)
	
	const dirName string = "./temp-test-dir"
	monitor, err := NewDirectoryMonitor(dirName)
	if monitor != nil { t.Error("Should not return a monitor for a non-existant directory")}
	if err == nil { t.Error("Should return an error for a non-existant directory")}

	os.Mkdir(dirName, 16877)
	defer os.RemoveAll(dirName)
	monitor, err = NewDirectoryMonitor(dirName)
	if monitor == nil { t.Error("Should return a monitor for a directory")}
	if err != nil { t.Error("Should not return an error for a directory")}
	
	var testListener *TestDirectoryListener = new(TestDirectoryListener)
	monitor.Listen(testListener)

	changed, err := monitor.Poll()
	if err != nil { t.Error("Should not return an error")}
	if changed { t.Error("Should not be changed on the first poll") }
	if len(testListener.AddedFiles) != 0 { t.Error()}
	if len(testListener.RemovedFiles) != 0 { t.Error()}
	if len(testListener.ChangedFiles) != 0 { t.Error()}

	changed, err = monitor.Poll()
	if err != nil { t.Error("Should not return an error")}
	if changed { t.Error("Nothing should have changed", testListener.ChangedFiles[0].Name) }
	if len(testListener.AddedFiles) != 0 { t.Error()}
	if len(testListener.RemovedFiles) != 0 { t.Error()}
	if len(testListener.ChangedFiles) != 0 { t.Error()}
	
	tempFile,err := os.Create(dirName + "/temp-test-file")
	defer os.Remove(tempFile.Name())
	tempFileInfo,err := tempFile.Stat()
	wr := bufio.NewWriter(tempFile) 
	wr.WriteString("I like traffic lights.\n") 
	wr.Flush()
	
	changed, err = monitor.Poll()
	if err != nil { t.Error("Should not return an error")}
	if !changed { t.Error("Should have changed") }
	if len(testListener.AddedFiles) != 1 { t.Error("Should have an added file")}
	if len(testListener.RemovedFiles) != 0 { t.Error("Should not have received a removed file")}
	if len(testListener.ChangedFiles) != 0 { t.Error("Should not have received a changed file")}
	if testListener.AddedFiles[0].Name != tempFileInfo.Name { t.Error("Should have added the temp file") }
	testListener.Clear()

	wr.WriteString("But not when they are red.\n") 
	wr.Flush()
	changed, err = monitor.Poll()
	if err != nil { t.Error("Should not return an error")}
	if !changed { t.Error("Should have changed") }
	if len(testListener.AddedFiles) != 0 { t.Error("Should not receive added file", testListener.AddedFiles)}
	if len(testListener.RemovedFiles) != 0 { t.Error("Should not have received a removed file", testListener.RemovedFiles)}
	if len(testListener.ChangedFiles) != 1 { t.Error("Should have received a changed file")}
	testListener.Clear()

	os.Remove(tempFile.Name())
	changed, err = monitor.Poll()
	if err != nil { t.Error("Should not return an error")}
	if !changed { t.Error("Should have changed") }
	if len(testListener.AddedFiles) != 0 { t.Error("Should not receive added file")}
	if len(testListener.RemovedFiles) != 1 { t.Error("Should have received a removed file")}
	if len(testListener.ChangedFiles) != 0 { t.Error("Should not have received a changed file")}
	testListener.Clear()
	
}