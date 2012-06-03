package main

import (
	"fmt"
	"os"
	"seven5/util"
	"strings"
)

//fileCollector is a type that can be used to keep track of a set of files that are
//going to be polled with a directory monitor. It implements directory listener.
type fileCollector struct {
	content *util.BetterList
	name    string
}

//create a new fileCollector with a given suffix and attached to a given
//monitor. if the name of the suffixs is _foo.go the result name of this
//object is foo.  It defaults ta having all the items in its list.
func newFileCollector(suffix string, monitor *util.DirectoryMonitor) (*fileCollector, error) {
	n := suffix
	if strings.HasSuffix(n, ".go") {
		n = strings.Replace(n, ".go", "", 1)
	}
	if strings.HasPrefix(n, "_") {
		n = strings.Replace(n, "_", "", 1)
	}
	result := &fileCollector{content: util.NewBetterList(), name: n}
	monitor.Listen(result)
	f, err := os.Open(monitor.Path)
	if err!=nil {
		return nil, err
	}
	names, err:= f.Readdirnames(-1)
	if err!=nil {
		return nil, err
	}
	for _, n:= range names {
		if strings.HasSuffix(n,suffix) {
			result.content.PushBack(n)
		}
	}
	return result, nil
}

func (self *fileCollector) GetFileList() []string {
	result := []string{}
	for e := self.content.Front(); e != nil; e = e.Next() {
		result = append(result, e.Value.(string))
	}
	return result
}

func (self *fileCollector) FileChanged(fileInfo os.FileInfo) {
	fmt.Printf("file changed %+v\n",fileInfo)
	//already on the list?
	if !self.content.Contains(fileInfo.Name()) {
		fmt.Fprintf(os.Stderr, "Whoa! Out of sync %s collector: %s not found!\n",
			self.name, fileInfo.Name())
		self.content.PushBack(fileInfo.Name())
	}
}
func (self *fileCollector) FileAdded(fileInfo os.FileInfo) {
	fmt.Printf("file added %+v\n",fileInfo)
	if self.content.Contains(fileInfo.Name()) {
		fmt.Fprintf(os.Stderr, "Whoa! Out of sync %s collector: %s discovered already!\n",
			self.name, fileInfo.Name())
	} else {
		self.content.PushBack(fileInfo.Name())
	}
}
func (self *fileCollector) FileRemoved(fileInfo os.FileInfo) {
	fmt.Printf("file removed %+v\n",fileInfo)
	if !self.content.Contains(fileInfo.Name()) {
		fmt.Fprintf(os.Stderr, "Whoa! Out of sync %s collector: %s not found!\n",
			self.name, fileInfo.Name())
	} else {
		self.content.RemoveValue(fileInfo.Name())
	}
}
