package main

import (
	"net/http"
	"os"
	"seven5/util"
)

//this is the object used to watch the configuration files
type configWatcher struct {
	monitor    *util.DirectoryMonitor
	workingDir string
	app        *singleFileListener
	groupie    *singleFileListener
}

//poll looks at the app.json and groupie.json to see if there were any
//changes ... returns app.changed, groupie.changed, and err
func (self *configWatcher) poll(writer http.ResponseWriter, request *http.Request, logger util.SimpleLogger) (bool, bool, error) {
	var err error

	self.app.changed = false
	self.groupie.changed = false
	
	_, err = self.monitor.Poll()
	if err != nil {
		logger.Error("Problem reading the directory %s: %s", self.workingDir, err.Error())
		return false, false, err
	}
	
	return self.app.changed, self.groupie.changed, nil
}

//newConfigWatcher creates a watcher for the json files in a given directory
func newConfigWatcher(dir string) (*configWatcher, error) {
	var err error

	result := &configWatcher{workingDir: dir}
	result.monitor, err = util.NewDirectoryMonitor(dir, ".json")
	result.app = &singleFileListener{name: "app.json"}
	result.groupie = &singleFileListener{name: "groupie.json"}
	if err != nil {
		return nil, err
	}
	result.monitor.Listen(result.app)
	result.monitor.Listen(result.groupie)
	return result, nil
}

//
// singleFileListener
//
type singleFileListener struct {
	name  string
	changed bool
}

func (self *singleFileListener) checkForMatch(info os.FileInfo)  {
	self.changed = (info.Name() == self.name)
}

func (self *singleFileListener) FileAdded(info os.FileInfo) {
	self.checkForMatch(info)
}
func (self *singleFileListener) FileChanged(info os.FileInfo) {
	self.checkForMatch(info)
}
func (self *singleFileListener) FileRemoved(info os.FileInfo) {
	self.checkForMatch(info)
}
