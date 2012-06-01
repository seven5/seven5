package main

import (
	"net/http"
	"path/filepath"
	"seven5/util"
)

type sourceWatcher struct {
	workingDir string
	allSource  *util.DirectoryMonitor
	vocab      *util.DirectoryMonitor
}

//newSourceWatcher should be passed the _root_ directory of the project, not
//the source directory.  this is only use in validated projects so there is
//no problem with the path working out
func newSourceWatcher(cwd string) (*sourceWatcher, error) {
	var err error
	parent := filepath.Base(cwd)
	srcpath := filepath.Join(cwd, "src", parent)
	self := &sourceWatcher{}
	if self.allSource, err = util.NewDirectoryMonitor(srcpath, ".go"); err != nil {
		return nil, err
	}
	if self.vocab, err = util.NewDirectoryMonitor(srcpath, "_vocab.go"); err != nil {
		return nil, err
	}
	return self, nil
}

//pollAllSource checks to see if any source code has changed.  if this is true
//then you need to rebuild user library first before proceeding.
func (self *sourceWatcher) pollAllSource(writer http.ResponseWriter, request *http.Request, 
	logger util.SimpleLogger) (bool,error) {

	changed, err := self.allSource.Poll()
	if err != nil {
		logger.Error("Unable to read the source directory for the project: %s",
			err.Error())
		return false, err
	}
	return changed, nil
}
