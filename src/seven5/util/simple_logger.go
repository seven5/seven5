package util

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"encoding/json"
)

//SimpleLogger is an interface representing a logger. There is a need for
//different types of loggers in the system so we can run tests on the
//console but "normally" run against a http response.
type SimpleLogger interface {
	Debug(fmtString string, obj ...interface{})
	Info(fmtString string, obj ...interface{})
	Warn(fmtString string, obj ...interface{})
	//Error is called when the _client_ program has a fatal error.
	Error(fmtString string, obj ...interface{})
	//Panic is called when the seven5 system itself cannot continue.
	Panic(fmtString string, obj ...interface{})
	//Dump out a protocol trace
	DumpRequest(req *http.Request)
	//Dump out a blob of Json
	DumpJson(json string)
	//Dump out a bunch of terminal data, usually a compile result
	DumpTerminal(out string)
}

// LoggerImpl encodes the real difference between the logger types.  The
// SimpleLogger interface is just sugar around this meat.
type LoggerImpl interface {
	//Print something to the output, with log style formatting
	Print(level int, isProto bool, fmtString string, obj ...interface{})
	//Print something directly to the output (no logger formatting)
	Raw(s string)
	//Dump out a protocol trace
	DumpRequest(req *http.Request)
	//Dump out a bunch of JSON
	DumpJson(json string)
	//Dump out a bunch of terminal data, usually a compile result
	DumpTerminal(out string)
}

// BaseLogger is the common portion between the two logging types
type BaseLogger struct {
	level int
	dumps bool
	impl  LoggerImpl
}

// TerminalLoggerImpl is the terminal specific stuff for logging.
type TerminalLoggerImpl struct {
	writer io.Writer
}

// HtmlLoggerImpl is the web specific stuff for logging.
type HtmlLoggerImpl struct {
	writer        io.Writer
	emittedHeader bool
}

//NewHtmlLogger creates a new HTML logger at the given level and protocol
//setings.  It will output to the given writer.
func NewHtmlLogger(level int, proto bool, writer io.Writer) SimpleLogger {
	impl := &HtmlLoggerImpl{writer, false}
	result := &BaseLogger{level, proto, impl}
	return result
}

//NewTerminalLogger creates a new terminal-oriented logger at the given
//protocol and level settings
func NewTerminalLogger(writer io.Writer, level int, proto bool) SimpleLogger {
	impl := &TerminalLoggerImpl{writer}
	result := &BaseLogger{level, proto, impl}
	return result
}

// Logging level constants
const (
	DEBUG = iota
	INFO
	WARN
	ERROR
	PANIC
)

//DumpHttpRequest 
func (self *BaseLogger) DumpRequest(req *http.Request) {
	if !self.dumps {
		return
	}
	self.impl.DumpRequest(req)
}

//DumpJson
func (self *BaseLogger) DumpJson(blob string) {
	if !self.dumps {
		return
	}
	var buffer bytes.Buffer
	json.Indent(&buffer,[]byte(blob),""," ")
	self.impl.DumpJson(buffer.String())
}

//DumpTerminal
func (self *BaseLogger) DumpTerminal(out string) {
	if !self.dumps {
		return
	}
	self.impl.DumpTerminal(out)
}

// Debug prints a message at DEBUG level.
func (self *BaseLogger) Debug(fmtString string, obj ...interface{}) {
	self.impl.Print(DEBUG, false, fmtString, obj...) // no sense in doing a comparison
}

// Info prints a message at INFO level.
func (self *BaseLogger) Info(fmtString string, obj ...interface{}) {
	if self.level <= INFO {
		self.impl.Print(INFO, false, fmtString, obj...)
	}
}

// Warn prints a message at WARN level.
func (self *BaseLogger) Warn(fmtString string, obj ...interface{}) {
	if self.level <= WARN {
		self.impl.Print(WARN, false, fmtString, obj...)
	}
}

// Erorr is called when the user program has an fatal error but Seven5
// can continue running.
func (self *BaseLogger) Error(fmtString string, obj ...interface{}) {
	if self.level <= ERROR {
		self.impl.Print(ERROR, false, fmtString, obj...)
	}
}

// Panic is called when Seven5 cannot continue operating because of
// an error.
func (self *BaseLogger) Panic(fmtString string, obj ...interface{}) {
	self.impl.Print(PANIC, false, fmtString, obj...)
	panic(fmt.Sprintf(fmtString, obj...))
}

//
// TERMINAL LOGGER IMPL
//

//Print just prints out a simple text version on the terminal.
func (self *TerminalLoggerImpl) Print(level int, isProto bool, fmtString string, obj ...interface{}) {
	lastElement, line := getCallerAndLine()
	now := time.Now()
	hour := now.Hour()
	minute := now.Minute()
	levelName := levelToString(level)
	self.Raw(fmt.Sprintf("[%s]%02d:%02d(%s:%d)%s", levelName, hour,
		minute, lastElement, line, fmt.Sprintf(fmtString, obj...)))
}

//Print raw message on the terminal.  Use stdout, not stderr so tests can pass
//quietly.
func (self *TerminalLoggerImpl) Raw(s string) {
	fmt.Fprintln(self.writer, s)
}

func (self *TerminalLoggerImpl) DumpJson(json string) {
	self.Raw("=========JSON START===========")
	self.Raw(json)
	self.Raw("=========JSON END===========")
}

func (self *TerminalLoggerImpl) DumpTerminal(out string) {
	self.Raw("=========TERMINAL OUTPUT START===========")
	self.Raw(out)
	self.Raw("=========TERMINAL OUTPUT END===========")
}

// Dump prints the protocol message contained in the parameter to the
// terminal with a bit of framing to make it look different than normal
// log output.
func (self *TerminalLoggerImpl) DumpRequest(req *http.Request) {
	self.Raw("=========REQUEST START===========")

	header := req.Header
	for k, v := range header {
		if len(v) > 1 {
			self.Raw(fmt.Sprintf("%10s:", k))
			for l := range v {
				self.Raw(fmt.Sprintf("%20d:", l))
			}
		} else if len(v) == 1 {
			self.Raw(fmt.Sprintf("%s:%s", k, v[0]))
		} else {
			self.Raw(fmt.Sprintf("%s", k))
		}
	}
	self.Raw("=========REQUEST END===========")
}



//levelToString is a utility function for getting the human readable level name.
//It's also used for class names in CSS for HTML.
func levelToString(level int) string {
	levelName := "debug"

	if level == INFO {
		levelName = "info"
	} else if level == WARN {
		levelName = "warn"
	} else if level == ERROR {
		levelName = "error"
	} else if level == PANIC {
		levelName = "panic"
	}
	return levelName
}

//getCallerAndLine is a utility for getting the calling file and line number
//of the log routines.
func getCallerAndLine() (string, int) {
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		log.Panicf("aborting due to failure to understand call stack")
	}
	split := strings.Split(file, string(filepath.Separator))
	return split[len(split)-1], line
}

//
// HTML LOGGER IMPL

//Print (at the HtmlLogger) prints a nicely formatted Html output message.
func (self *HtmlLoggerImpl) Print(level int, isProto bool, fmtString string, obj ...interface{}) {
	var buffer bytes.Buffer
	now := time.Now()
	hour := now.Hour()
	minute := now.Minute()

	self.SetupHTML()
	lastElement, line := getCallerAndLine()

	typeName := "log"
	levelName := levelToString(level)

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
	self.Raw(buffer.String())
}

//HtmlRaw produces no log frontmatter on the HTML page.
func (self *HtmlLoggerImpl) Raw(s string) {
	self.writer.Write([]byte(s))
}


// SetupHTML should be called before any output to the result takes place.
// It can be called any number of times, it only outputs once.  It places
// the necessary CSS cruft on the stream to allow future log messages to
// look purty.
func (self *HtmlLoggerImpl) SetupHTML() {
	if self.emittedHeader {
		return
	}
	self.emittedHeader = true
	self.Raw(HTMLHeader)
}


// Dump prints the protocol message contained in the parameter to the log
// if the protocol parameter is enabled.
func (self *HtmlLoggerImpl) DumpRequest(req *http.Request) {
	self.Raw("<div class=\"protobox http\">\n")

	header := req.Header
	for k, v := range header {
		if len(v) > 1 {
			self.Raw(fmt.Sprintf("%20s:", k))
			for l := range v {
				self.Raw(fmt.Sprintf("%20d:", l))
			}
		} else if len(v) == 1 {
			self.Raw(fmt.Sprintf("%s:%s", k, v[0]))
		} else {
			self.Raw(fmt.Sprintf("%s", k))
		}
	}
	self.Raw("</div>\n")

}

// Dump prints the protocol message contained in the parameter to the log
// if the protocol parameter is enabled.
func (self *HtmlLoggerImpl) DumpJson(json string) {
	self.Raw("<div class=\"protobox json\"><pre>\n")
	self.Raw(json)
	self.Raw("</pre></div>\n")
}

// Dump prints the protocol message contained in the parameter to the log
// if the protocol parameter is enabled.
func (self *HtmlLoggerImpl) DumpTerminal(out string) {
	self.Raw("<div class=\"protobox terminal\"><pre>\n")
	self.Raw(out)
	self.Raw("</pre></div>\n")
}

const HTMLHeader = `
<html>
<head>
<style type="text/css">

.log {
    font-family: Courier;
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
	margin-left: 40px;
	#overflow: hidden;
	#text-overflow: ellipsis;
	#white-space : nowrap;
}

.http {
	color: green;
}

.json {
	color: orange;
}

.terminal {
	color: black;
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
