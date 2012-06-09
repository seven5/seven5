package main

import (
	"net/http"
	"os"
	"seven5"
	"seven5/util"
	"seven5/cmd"
)

//this is the object used to watch the configuration files
type configWatcher struct {
	monitor    *util.DirectoryMonitor
	workingDir string
	project        *singleFileListener
	command    *singleFileListener
}

//poll looks at the project.json and command.json to see if there were any
//changes ... returns project.changed, command.changed, and err
func (self *configWatcher) poll(writer http.ResponseWriter, request *http.Request, logger util.SimpleLogger) (bool, bool, error) {
	var err error

	self.project.changed = false
	self.command.changed = false
	
	_, err = self.monitor.Poll()
	if err != nil {
		logger.Error("Problem reading the directory %s: %s", self.workingDir, err.Error())
		return false, false, err
	}
	
	return self.project.changed, self.command.changed, nil
}

//newConfigWatcher creates a watcher for the json files in a given directory
func newConfigWatcher(dir string) (*configWatcher, error) {
	var err error

	result := &configWatcher{workingDir: dir}
	result.monitor, err = util.NewDirectoryMonitor(dir, ".json")
	result.project = &singleFileListener{name: cmd.PROJECT_CONFIG_FILE}
	result.command = &singleFileListener{name: seven5.COMMAND_CONFIG_FILE}
	if err != nil {
		return nil, err
	}
	result.monitor.Listen(result.project)
	result.monitor.Listen(result.command)
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
