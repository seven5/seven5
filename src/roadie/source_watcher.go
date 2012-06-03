package main

import (
	"net/http"
	"path/filepath"
	"seven5/util"
)

type sourceWatcher struct {
	workingDir string
	allSource  *util.DirectoryMonitor
	vocabMon   *util.DirectoryMonitor
	vocab      *fileCollector
}

const VOCAB_SUFFIX = "_vocab.go"

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
	if self.vocabMon, err = util.NewDirectoryMonitor(srcpath, VOCAB_SUFFIX); err != nil {
		return nil, err
	}
	if self.vocab ,err = newFileCollector(VOCAB_SUFFIX, self.vocabMon); err!=nil {
		return nil, err
	}
	
	return self, nil
}

//pollAllSource checks to see if any source code has changed.  if this is true
//then you need to rebuild user library first before proceeding.
func (self *sourceWatcher) pollAllSource(writer http.ResponseWriter, request *http.Request,
	logger util.SimpleLogger) (bool, error) {

	changed, err := self.allSource.Poll()
	if err != nil {
		logger.Error("Unable to read the source directory for the project: %s",
			err.Error())
		return false, err
	}
	return changed, nil
}

//pollVocab checks to see if vocab's have changed.  if anything changed, it
//return a list of files to be rebuilt.  if nothing changed this array is
//empty (not nil).
func (self *sourceWatcher) pollVocab(writer http.ResponseWriter, request *http.Request,
	logger util.SimpleLogger) ([]string, error) {
	
	logger.Debug("About try polling the vocabulary monitor..")
	changed, err := self.vocabMon.Poll()
	if err != nil {
		return []string{}, err
	}
	if changed {
		return self.vocab.GetFileList(), nil
	}
	return []string{}, nil
}
