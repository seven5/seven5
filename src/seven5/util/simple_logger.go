package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

//number of lines above and below to show
const FILE_ERROR_CONTEXT = 3


//FileErrorLogItem represents what will show to the user when we display an error
//message in the browser.
type FileErrorLogItem struct {
	Path string
	Msg string
	Line int
}

//SimpleLogger is an interface representing a logger. There is a need for
//different types of loggers in the system so we can run tests on the
//console but "normally" run against a http response.
type SimpleLogger interface {
	Debug(fmtString string, obj ...interface{})
	Info(fmtString string, obj ...interface{})
	Warn(fmtString string, obj ...interface{})
	//Error is called when the _client_ program has a fatal error.
	Error(fmtString string, obj ...interface{})
	//Dump out a protocol trace
	DumpRequest(level int, title string, req *http.Request)
	//Dump out a blob of Json
	DumpJson(level int, title string, json string)
	//Dump out a bunch of terminal data, usually a compile result
	DumpTerminal(level int, title string, out string)
	//Print string to terminal without formatting help
	Raw(s string)
	//Get the current log level
	GetLogLevel() string
	//FileError displays a list of errors to the terminal based on compiler
	//output. The parameter is a list of seven5.FileErrorLogInfo
	FileError(*BetterList) 
}

// LoggerImpl encodes the real difference between the logger types.  The
// SimpleLogger interface is just sugar around this meat.
type LoggerImpl interface {
	//Print something to the output, with log style formatting
	Print(level int, isProto bool, fmtString string, obj ...interface{})
	//Print something directly to the output (no logger formatting)
	Raw(s string)
	//Dump out a protocol trace
	DumpRequest(level int, title string, req *http.Request)
	//Dump out a bunch of JSON
	DumpJson(level int, title string, json string)
	//Dump out a bunch of terminal data, usually a result from cmd line program
	DumpTerminal(level int, title string, out string)
	//Display an error in a program file
	DumpFile(path string, msg string, currLine int, targetLine int, file *os.File)
}

// BaseLogger is the common portion between the two logging types
type BaseLogger struct {
	level int
	impl  LoggerImpl
}

// TerminalLoggerImpl is the terminal specific stuff for logging.
type TerminalLoggerImpl struct {
	writer io.Writer
}

// HtmlLoggerImpl is the web specific stuff for logging.
type HtmlLoggerImpl struct {
	writer io.Writer
}

//NewHtmlLogger creates a new HTML logger at the given level and protocol
//setings.  It will output to the given writer.
func NewHtmlLogger(level int, writer io.Writer, header bool) SimpleLogger {
	impl := &HtmlLoggerImpl{writer}
	result := &BaseLogger{level, impl}
	if header {
		impl.Start()
	}
	return result
}

//NewTerminalLogger creates a new terminal-oriented logger at the given
//protocol and level settings
func NewTerminalLogger(writer io.Writer, level int) SimpleLogger {
	impl := &TerminalLoggerImpl{writer: writer}
	result := &BaseLogger{level, impl}
	return result
}

// Logging level constants
const (
	DEBUG = iota
	INFO
	WARN
	ERROR
)

//DumpHttpRequest 
func (self *BaseLogger) DumpRequest(level int, title string, req *http.Request) {
	if self.level > level {
		return
	}
	self.impl.DumpRequest(level, title, req)
}

//Raw 
func (self *BaseLogger) Raw(s string) {
	self.impl.Raw(s)
}

//DumpJson
func (self *BaseLogger) DumpJson(level int, title string, blob string) {
	if self.level > level {
		return
	}
	var buffer bytes.Buffer
	if err := json.Indent(&buffer, []byte(blob), "", " "); err != nil {
		fmt.Fprintf(os.Stderr, "error in indent: %s\n", err)
	}
	self.impl.DumpJson(level, title, buffer.String())
}

//DumpTerminal
func (self *BaseLogger) DumpTerminal(level int, title string, out string) {
	if self.level > level {
		return
	}
	self.impl.DumpTerminal(level, title, out)
}

// Debug prints a message at DEBUG level.
func (self *BaseLogger) Debug(fmtString string, obj ...interface{}) {
	if self.level <= DEBUG {
		self.impl.Print(DEBUG, false, fmtString, obj...) // no sense in doing a comparison
	}
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
// GetLogLevel is used to transfer log levels "over the wire"
func (self *BaseLogger) GetLogLevel() string {
	return levelToString(self.level)
}

//FileError shows a list of nicely formatted error messages
func (self *BaseLogger) FileError(list *BetterList) {
	if self.level > ERROR {
		return //not sure this is useful or meaningful
	}

	for e := list.Front(); e != nil; e = e.Next() {
		item := e.Value.(*FileErrorLogItem)
		min := 0
		if item.Line > FILE_ERROR_CONTEXT{
			min = item.Line - FILE_ERROR_CONTEXT
		}
		rd, err := os.Open(item.Path)
		if err!=nil {
			self.Error("Internal error reading file with error %+v:%s",item,err)
			return
		}
		
		//read min lines
		var i int
		for i=0; i<min; i++ {
			_, err := ReadLine(rd)
			if err!=nil {
				self.Error("Internal error reading file for error message %+v:%s",item,err)
				return
			}
		}
		
		self.impl.DumpFile(item.Path,item.Msg,i+1,item.Line, rd)
		rd.Close()
	}	
}

//
// TERMINAL LOGGER IMPL
//

//Print just prints out a simple text version on the terminal.
func (self *TerminalLoggerImpl) Print(level int, isProto bool, fmtString string, obj ...interface{}) {
	lastElement, line := GetCallerAndLine(3)
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

func (self *TerminalLoggerImpl) DumpJson(level int, title string, json string) {
	self.Raw(fmt.Sprintf("-----> %s (%5s) <-----", title, levelToString(level)))
	self.Raw("========= JSON START ============")
	self.Raw(json)
	self.Raw("========= JSON   END ============")
}

func (self *TerminalLoggerImpl) DumpTerminal(level int, title string, out string) {
	self.Raw(fmt.Sprintf("-----> %s (%5s) <-----", title, levelToString(level)))
	self.Raw("=========TERMINAL OUTPUT START ===========")
	self.Raw(out)
	self.Raw("=========TERMINAL OUTPUT  END  ===========")
}

// Dump prints the protocol message contained in the parameter to the
// terminal with a bit of framing to make it look different than normal
// log output.
func (self *TerminalLoggerImpl) DumpRequest(level int, title string, req *http.Request) {
	self.Raw(fmt.Sprintf("-----> %s (%5s) <-----", title, levelToString(level)))
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

func (self *TerminalLoggerImpl) DumpFile(path string, msg string, currLine int, 
	targetLine int, file *os.File) {
	
	self.Raw(fmt.Sprintf("ERROR: %s:%s:%s",filepath.Base(path),targetLine,msg))
	self.Raw(fmt.Sprintf("========= %s ===========",filepath.Base(path)))
	
	for i:=currLine; i<targetLine+FILE_ERROR_CONTEXT;i++ {
		l, err:= ReadLine(file)
		if err==io.EOF {
			break
		}
		if err!=nil {
			self.Print(ERROR,false,"internal error reading %s:%s",filepath.Base(path),err)
			break
		}
		val:=fmt.Sprintf("%5d",i)
		if i==targetLine {
			val=">>>>>"
		}
		self.Raw(fmt.Sprintf("%s%s\n",val,l))
	}
	self.Raw(fmt.Sprintf("========= %s ===========",filepath.Base(path)))
	return
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
	}
	return levelName
}


func LogLevelStringToLevel(levelString string) int {
	switch levelString {
	case "info":
		return INFO
	case "warn":
		return WARN
	case "error":
		return ERROR
	}
	return DEBUG
}

//getCallerAndLine is a utility for getting the calling file and line number
//of the log routines.
func GetCallerAndLine(level int) (string, int) {
	_, file, line, ok := runtime.Caller(level)
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

	lastElement, line := GetCallerAndLine(3)

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
		}
	}
	buffer.WriteString(fmt.Sprintf("<span class=\"%s %s\">", typeName, levelName))

	if !isProto {
		prefix := fmt.Sprintf("[%02d:%02d]%s:%-4d", hour, minute, lastElement, line)
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

// Start should be called before any output to the result takes place.
// It places the necessary CSS cruft on the stream to allow future log messages to
// look purty.
func (self *HtmlLoggerImpl) Start() {
	self.Raw(HTMLHeader)
}

// Dump prints the protocol message contained in the parameter to the log
// if the protocol parameter is enabled.
func (self *HtmlLoggerImpl) DumpRequest(level int, title string, req *http.Request) {
	self.Raw("<div class=\"protobox http\">\n")
	self.Raw(fmt.Sprintf("<H4>%s: %s</H4>", levelToString(level), title))
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
func (self *HtmlLoggerImpl) DumpJson(level int, title string, json string) {
	self.Raw("<div class=\"protobox json\"><pre>\n")
	tag := "H4"
	switch level {
	case INFO:
		tag = "H3"
	case WARN, ERROR:
		tag = "H2"
	}
	self.Raw(fmt.Sprintf("<%s>%s: %s</%s>", tag, levelToString(level), title, tag))
	self.Raw(json)
	self.Raw("</pre></div>\n")
}

// Dump prints the protocol message contained in the parameter to the log
// if the protocol parameter is enabled.
func (self *HtmlLoggerImpl) DumpTerminal(level int, title string, out string) {
	self.Raw("<div class=\"protobox terminal\"><pre>\n")
	tag := "H4"
	switch level {
	case INFO:
		tag = "H3"
	case WARN, ERROR:
		tag = "H2"
	}
	self.Raw(fmt.Sprintf("<%s>%s: %s</%s>", tag, levelToString(level), title, tag))
	self.Raw(out)
	self.Raw("</pre></div>\n")
}

func (self *HtmlLoggerImpl) DumpFile(path string, msg string, currLine int, 
	targetLine int, file *os.File) {
	
	self.Raw("<div class=\"protobox file\"><pre>\n")
	tag := "H2"
	self.Raw(fmt.Sprintf("<%s>%s: %s</%s>", tag, filepath.Base(path), msg, tag))
	
	for i:=currLine; i<targetLine+FILE_ERROR_CONTEXT;i++ {
		l, err:= ReadLine(file)
		if err==io.EOF {
			break
		}
		if err!=nil {
			self.Print(ERROR,false,"internal error reading %s:%s",filepath.Base(path),err)
			break
		}
		val:=fmt.Sprintf("%5d",i)
		if i==targetLine {
			self.Raw("<span class=\"errorline\">")
		}
		self.Raw(fmt.Sprintf("%s %s\n",val,l))
		if i==targetLine {
			self.Raw("</span>\n")
		}
	}
	
	self.Raw("</pre></div>\n")
	return
}
const HTMLHeader = `
<html>
<head>
<style type="text/css">

.log {
    font-family: Courier;
	color: black;
	margin-left: 20px;
}
.proto {
}

.protobox {
	background-color: white;
	border: 1px black solid;
	padding: 10px;
    font-family: Courier;
	margin-left: 20px;
	#overflow: hidden;
	#text-overflow: ellipsis;
	#white-space : nowrap;
}

.http {
	color: green;
}

.file {
	color: black;
}

.file .errorline {
	color: red;
	border: 1px solid red;
}

.json {
	color: orange;
}

.json h4 {
	color: black;
}

.json h3 {
	color: blue;
}

.json h2 {
	color: red;
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
