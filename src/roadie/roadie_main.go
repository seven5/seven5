package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"seven5"
	"seven5/groupie"
	"seven5/util"
	"time"
)

//wire is our connection to the seven5 binary
var wire *seven5.Wire
var cfg *configWatcher
var monitor *util.BetterList
var workingDirectory string


//this is the object used to watch the configuration files
type configWatcher struct {
	watched []string
	appConfig bool
	groupiesConfig bool
}

//main is the entry point of the roadie 
func main() {
	var err error

	if len(os.Args) > 2 {
		fmt.Fprintf(os.Stderr, "%s usage: %s [directory]\n", os.Args[0], os.Args[0])
		return
	}

	if len(os.Args) > 1 {
		if err = os.Chdir(os.Args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "%s: unable to change to %s: %s\n", os.Args[0], os.Args[1], err)
			return
		}
	}

	if workingDirectory, err = os.Getwd(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: unable to get working directory: %s\n", os.Args[0], err)
		return
	}

	modifiedGoPath := os.Getenv("GOPATH")+string(filepath.ListSeparator)+workingDirectory
	if err = os.Setenv("GOPATH", modifiedGoPath); err!=nil {
		fmt.Fprintf(os.Stderr, "%s: unable to set GOPATH: %s\n", os.Args[0], err)
		return
	}

	
	s := &http.Server{
		Addr:         ":9009",
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}
	
	http.HandleFunc("/seven5/build", build)

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "internal seven5 err:%s\n", r)
		}
	}()
	
	fmt.Fprintf(os.Stderr,"roadie error waiting on connections: %s",
		s.ListenAndServe().Error())
}

//checkListeners looks at all the listeners and if any of them indicate
//something changed, takes action
func checkListeners(writer http.ResponseWriter, request *http.Request, logger util.SimpleLogger) {
	var err error
	
	if cfg==nil {
		//setup the listeners for config files	
		base := filepath.Base(workingDirectory)
		cfg = newConfigWatcher(base)
		var d *util.DirectoryMonitor
		if d, err = util.NewDirectoryMonitor(workingDirectory,"json"); err!=nil {
			fmt.Fprintf(os.Stderr, "%s: unable to create directory monitor: %s\n", os.Args[0], err)
			return
		}
		fmt.Printf("----about to start listening\n")
		d.Listen(cfg)
		monitor= util.NewBetterList()
		monitor.PushBack(d)
		
		//first time through, force a build
		cfg.groupiesConfig = true
		cfg.appConfig = true
	} else {
		cfg.groupiesConfig = false
		cfg.appConfig = false
		pollAllDirectories()
	}
			
	//rebuild seven5 itself?
	if cfg.groupiesConfig {
		logger.Debug("groupies.json has been changed, rebuilding seven5...")
		if !canVerifyWire(true, writer, request, logger) {
			return
		}
	}
	if cfg.appConfig {
		logger.Debug("app.json has been changed, validating your app...")
		wire.Dispatch(groupie.VALIDATEPROJECT, writer, request, logger)
		
		//verify that the new layout and application config are ok
		//read in application config
		//verify the app is ok
	}
}

// canVerifyWire checks to see if we can connect to seven5.  we will fail
// this if the seven5 execute is not or cannot be built.  returns true
// if everything is ok for communication to the seven5 excutable.  You can
// use force to be sure of an attempt to build--useful if you know things
// are out of date.
func canVerifyWire(force bool, writer http.ResponseWriter, request *http.Request, logger util.SimpleLogger) bool{
	if force || wire==nil || !wire.HaveSeven5() {
		currentSeven5 := seven5.Bootstrap(writer, request, logger)
		wire = seven5.NewWire(currentSeven5)
		if !wire.HaveSeven5() {
			return false
		}
	}
	return true
}

// echo is a simple groupie to be used an example
func echo(writer http.ResponseWriter, request *http.Request) {
	logger:=util.NewHtmlLogger(util.DEBUG, true, writer, true)
	if canVerifyWire(false, writer,request,logger) {
		wire.Dispatch(groupie.ECHO, writer, request, logger)
	}
}

// called to build seven5 bythe roadie or by the user
func build(writer http.ResponseWriter, request *http.Request) {
	logger:=util.NewHtmlLogger(util.DEBUG, true, writer, true) 
	fmt.Printf("hitting the listeners")
	checkListeners(writer, request, logger)
}


//Ping all the monitors we have running right now so we can force updates
func pollAllDirectories() {
	for e := monitor.Front(); e != nil; e = e.Next() {
		m := e.Value.(*util.DirectoryMonitor)
		m.Poll()
	}
}


//
// configWatcher implementation for roadie
//
func newConfigWatcher(appName string) *configWatcher {
	r:= &configWatcher{watched : []string{ "app.json", "groupies.json"}}
	return r
}

func (self *configWatcher) checkChangeName(fileInfo os.FileInfo) {
	fmt.Printf("filename is %s\n",fileInfo.Name())
	for _, i := range self.watched {
		if i == fileInfo.Name() {
			switch i {
				case "app.json":
					self.appConfig=true
				case "groupies.json":
					self.groupiesConfig=true
			}
		}
	}
}

//
// These three fns are for the API to the DirectoryMonitor
//
func (self *configWatcher) FileChanged(fileInfo os.FileInfo) {
	self.checkChangeName(fileInfo)
}
func (self *configWatcher) FileRemoved(fileInfo os.FileInfo) {
	self.checkChangeName(fileInfo)
}
func (self *configWatcher) FileAdded(fileInfo os.FileInfo) {
	self.checkChangeName(fileInfo)
}
