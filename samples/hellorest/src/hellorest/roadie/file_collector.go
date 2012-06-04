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
	name   string
	excluder func(string)bool
	includer func(string)bool
}

//create a new fileCollector with a given suffix and attached to a given
//monitor. if the name of the suffixs is _foo.go the result name of this
//object is foo.  It defaults ta having all the items in its list.
func newFileCollector(name string, monitor *util.DirectoryMonitor,
	includer func(string)bool, excluder func(string)bool) (*fileCollector, error) {
	result := &fileCollector{content: util.NewBetterList(), name: name, 
		excluder: excluder, includer:includer}
	monitor.Listen(result)
	f, err := os.Open(monitor.Path)
	if err!=nil {
		return nil, err
	}
	names, err:= f.Readdirnames(-1)
	if err!=nil {
		return nil, err
	}
	//initial fill
	for _, n:= range names {
		//monitor will only show things with the suffix
		if (!strings.HasSuffix(n,monitor.Extension)) {
			continue;
		}
		if result.includer!=nil && result.includer(n) {
			result.content.PushBack(n)
			continue
		}
		if result.excluder!=nil && result.excluder(n) {
			continue
		}
		result.content.PushBack(n)
	}
	return result, nil
}

func (self *fileCollector) GetFileList() []string {
	result := []string{}
	fmt.Printf("calling GetFileList() %s and size is %d\n",self.name,self.content.Len())
	for e := self.content.Front(); e != nil; e = e.Next() {
		result = append(result, e.Value.(string))
	}
	return result
}

func (self *fileCollector) FileChanged(fileInfo os.FileInfo) {
	if self.excluder!=nil && self.excluder(fileInfo.Name()) {
		return
	}
	//already on the list?
	if !self.content.Contains(fileInfo.Name()) {
		fmt.Fprintf(os.Stderr, "Whoa! Out of sync %s collector: %s not found!\n",
			self.name, fileInfo.Name())
		self.content.PushBack(fileInfo.Name())
	}
}
func (self *fileCollector) FileAdded(fileInfo os.FileInfo) {
	if self.excluder!=nil && self.excluder(fileInfo.Name()) {
		return
	}
	if self.content.Contains(fileInfo.Name()) {
		fmt.Fprintf(os.Stderr, "Whoa! Out of sync %s collector: %s discovered already!\n",
			self.name, fileInfo.Name())
	} else {
		self.content.PushBack(fileInfo.Name())
	}
}
func (self *fileCollector) FileRemoved(fileInfo os.FileInfo) {
	if self.excluder!=nil && self.excluder(fileInfo.Name()) {
		return
	}
	if !self.content.Contains(fileInfo.Name()) {
		fmt.Fprintf(os.Stderr, "Whoa! Out of sync %s collector: %s not found!\n",
			self.name, fileInfo.Name())
	} else {
		self.content.RemoveValue(fileInfo.Name())
	}
}
