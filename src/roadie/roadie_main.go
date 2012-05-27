package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"seven5"
	"seven5/plugin"
	"seven5/util"
	"strings"
	"time"
)

//wire is our connection to the seven5 binary
var wire *seven5.Wire

func main() {
	var wd string
	var err error
	overrideGOPATH := false

	if len(os.Args) > 3 {
		fmt.Fprintf(os.Stderr, "%s: usage %s [directory] [GOPATH override?]\n", os.Args[0], os.Args[0])
		return
	}

	if len(os.Args) > 1 {
		if err = os.Chdir(os.Args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "%s: unable to change to %s: %s\n", os.Args[0], os.Args[1], err)
			return
		}
	}

	if len(os.Args) == 3 {
		overrideGOPATH = true
	}

	if wd, err = os.Getwd(); err != nil {
		fmt.Fprintf(os.Stderr, "%s: unable to get working directory: %s\n", os.Args[0], err)
		return
	}

	gopath := strings.SplitAfter(os.Getenv("GOPATH"), string(filepath.ListSeparator))
	ok := overrideGOPATH
	for _, pathElem := range gopath {
		if pathElem == wd {
			ok = true
			break
		}
	}

	if !ok {
		fmt.Fprintf(os.Stderr, "%s uses other go tools that depend on the GOPATH "+
			"environment variable.\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "    please add this to your GOPATH: %s\n", wd)
		return
	}

	s := &http.Server{
		Addr:         ":9009",
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}
	http.HandleFunc("/echo", echo)

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "internal seven5 err:%s\n", r)
		}
	}()
	
	fmt.Fprintf(os.Stderr,"roadie error waiting on connections: %s",
		s.ListenAndServe().Error())
}

func canVerifyWire(writer http.ResponseWriter, request *http.Request, logger util.SimpleLogger) bool{
	if wire==nil || !wire.HaveSeven5() {
		currentSeven5 := seven5.Bootstrap(writer, request, logger)
		wire = seven5.NewWire(currentSeven5)
		if !wire.HaveSeven5() {
			return false
		}
	}
	return true
}

func echo(writer http.ResponseWriter, request *http.Request) {
	logger:=util.NewHtmlLogger(util.DEBUG, true, writer, true)
	if canVerifyWire(writer,request,logger) {
		wire.Dispatch(plugin.ECHO, writer, request, logger)
	}
}
