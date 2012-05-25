package seven5

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
	"time"
)

// HtmlLogger is a logger that understands how to emit nice-looking HTML 
// to an http.ResponseWriter.
type HtmlLogger struct {
	Level         int
	Proto         bool
	Writer        http.ResponseWriter
	emittedHeader bool
}

const (
	DEBUG = iota
	INFO
	WARN
	ERROR
	PANIC
)

func (self *HtmlLogger) Debug(fmtString string, obj ...interface{}) {
	self.Print(DEBUG, false, fmtString, obj...) // no sense in doing a comparison
}

func (self *HtmlLogger) Info(fmtString string, obj ...interface{}) {
	if self.Level <= INFO {
		self.Print(INFO, false, fmtString, obj...)
	}
}

func (self *HtmlLogger) Warn(fmtString string, obj ...interface{}) {
	if self.Level <= WARN {
		self.Print(WARN, false, fmtString, obj...)
	}
}

func (self *HtmlLogger) Error(fmtString string, obj ...interface{}) {
	if self.Level <= ERROR {
		self.Print(ERROR, false, fmtString, obj...)
	}
}

func (self *HtmlLogger) Panic(fmtString string, obj ...interface{}) {
	self.Print(PANIC, false, fmtString, obj...)
	panic(fmt.Sprintf(fmtString, obj...))
}

func (self *HtmlLogger) Protocol(fmtString string, obj ...interface{}) {
	if self.Proto {
		self.Print(DEBUG, true, fmtString, obj...)
	}
}

func (self *HtmlLogger) Print(level int, isProto bool, fmtString string, obj ...interface{}) {
	var buffer bytes.Buffer
	_, file, line, ok := runtime.Caller(3)
	now := time.Now()
	hour := now.Hour()
	minute := now.Minute()

	self.SetupHTML()

	if !ok {
		log.Panicf("aborting due to failure to understand call stack")
	}

	split := strings.Split(file, "/")
	if len(split) == 1 && split[0] == file {
		split = strings.Split(file, "\\")
	}
	lastElement := split[len(split)-1]

	typeName := "log"
	levelName := "debug"

	if isProto {
		typeName = "proto"
		levelName = ""
	} else {
		if level == INFO {
			levelName = "info"
		} else if level == WARN {
			levelName = "warn"
		} else if level == ERROR {
			levelName = "error"
		} else if level == PANIC {
			levelName = "panic"
		}
	}
	buffer.WriteString(fmt.Sprintf("<span class=\"%s %s\">", typeName, levelName))

	if !isProto {
		prefix := fmt.Sprintf("[%02d:%02d]%s:%-4d|", hour, minute, lastElement, line)
		buffer.WriteString(prefix)
	}

	s := fmt.Sprintf(fmtString, obj...)
	buffer.WriteString(s)

	buffer.WriteString(fmt.Sprintf("</span><br>"))
	self.Writer.Write(buffer.Bytes())
}

// SetupHTML should be called before any output to the result takes place.
// It can be called any number of times, it only outputs once.  It places
// the necessary CSS cruft on the stream to allow future log messages to
// look purty.
func (self *HtmlLogger) SetupHTML() {
	if self.emittedHeader {
		return
	}
	self.emittedHeader = true
	self.Writer.Write([]byte(HTMLHeader))
}

// Dump prints the protocol message contained in the parameter to the log
// if the protocol parameter is enabled 
func (self *HtmlLogger) Dump(req *http.Request) {
	if !self.Proto {
		return
	}
	
	self.Writer.Write([]byte("<div class=\"protobox\">\n"));
	
	header := req.Header
	for k, v := range header {
		if len(v) > 1 {
			self.Protocol("%20s:", k)
			for l := range v {
				self.Protocol("%20d:", l)
			}
		} else if len(v) == 1 {
			self.Protocol("%s:%s", k, v[0])
		} else {
			self.Protocol("%s", k)
		}
	}
	self.Writer.Write([]byte("</div>\n"));
	
}

const HTMLHeader = `
<html>
<head>
<style type="text/css">

.log {
    font-family: Courier;
    font-size: 1.25em;
	color: black;
	margin-left: 10px;
}
.proto {
}

.protobox {
	background-color: white;
	border: 1px black solid;
	padding: 10px;
    font-family: Courier;
    font-size: 0.7em;
	color: green;
	margin-left: 40px;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space : nowrap;
}

.debug {
	color: black;
}
.info {
	color: blue;
}
.warn {
	color: yellow;
}
.error {
	color: red;
}

body {background-color:lightgrey;}
</style>
</head>
<body>
`
