package seven5

import (
	"fmt"
	"log/syslog"
	"os"
	"runtime"
	"strings"
	"time"
)

func init() {
	SetDefaultLogger(NewSimpleLogger("seven5", syslog.LOG_DEBUG, true))
}

//Logger is a simple logging interface that corresponds to the first 5 levels of Syslog.
type Logger interface {
	Debugf(spec string, other ...interface{})
	Infof(spec string, other ...interface{})
	Noticef(spec string, other ...interface{})
	Warnf(spec string, other ...interface{})
	Errf(spec string, other ...interface{})
	//same as err followed by panic
	Panicf(spec string, other ...interface{})
	//same as Error followed by os.Exit(1)
	Fatalf(spec string, other ...interface{})

	//useful for async loggers when you are about to exit the main program
	WaitToDie()

	//convenience method for recording a short message about the location and the error
	Error(string, error)
	//Same as Error, followed by panic
	ErrorPanic(string, error)
	//Same as Error,  followed by os.Exit(1)
	ErrorFatal(string, error)
}

type SimpleLogger struct {
	thresh syslog.Priority
	prefix string
	date   bool
	Out    *os.File
	Err    *os.File
}

func PriorityToString(self syslog.Priority) string {
	switch self {
	case syslog.LOG_DEBUG:
		return "DEBUG"
	case syslog.LOG_INFO:
		return "INFO"
	case syslog.LOG_WARNING:
		return "WARN"
	case syslog.LOG_ERR:
		return "ERR"
	case syslog.LOG_ALERT:
		return "ALRT"
	case syslog.LOG_CRIT:
		return "CRIT"
	case syslog.LOG_EMERG:
		return "EMERG"
	}
	return "NOTICE"
}

//L is for convenience of notation so that seven5.L.Fatalf() works.  It's also kept here
//to avoid needing to pass the logger all over through all functions.
var L Logger

//SetDefaultLogger is a way to change the behavior of the logger for using an external
//service or provding extra logging capability.
func SetDefaultLogger(l Logger) {
	L = l
}

//NewSimpleLogger creates a new simple logger with the provided prefix and threshold.
//If date is true date will be included in log lines.  The fields Out and Err are
//set to stdout and stderr, respectively but can be changed by the caller.
func NewSimpleLogger(prefix string, thresh syslog.Priority, date bool) *SimpleLogger {
	result := &SimpleLogger{
		thresh: thresh,
		prefix: "[" + prefix + "]",
		date:   date,
		Out:    os.Stdout,
		Err:    os.Stderr,
	}
	return result
}

func (self *SimpleLogger) Outf(prio syslog.Priority, spec string, other ...interface{}) {
	if self.Out != nil {
		if prio <= self.thresh {
			if self.date {
				fmt.Fprintf(self.Out, "%s ", time.Now().Format(time.RFC3339))
			}
			fmt.Fprintf(self.Out, self.prefix+" %s ", PriorityToString(prio))
			_, f, l, ok := runtime.Caller(2)
			if ok {
				pieces := strings.Split(f, "/")
				fmt.Fprintf(self.Out, "%s:%d:", pieces[len(pieces)-1], l)
			}
			fmt.Fprintf(self.Out, spec, other...)
			if !strings.HasSuffix(spec, "\n") {
				fmt.Fprintf(self.Out, "\n")
			}
		}
	}
}

func (self *SimpleLogger) ErrorLevelf(prio syslog.Priority, where string, err error) {
	if self.Err != nil {
		if prio <= self.thresh {
			fmt.Fprintf(self.Err, "---------------- ERROR [Type:%T] ---------------------\n", err)
			fmt.Fprintf(self.Err, "%+v\n", err)
			if self.date {
				fmt.Fprintf(self.Err, "%s ", time.Now().Format(time.RFC3339))
			}
			fmt.Fprintf(self.Err, "%s:%s\n", where, err.Error())
			for _, i := range []int{2, 3, 4, 5} {
				_, f, l, ok := runtime.Caller(i)
				if ok {
					fmt.Fprintf(self.Err, "%s:%d\n", f, l)
				}
			}
			fmt.Fprintf(self.Err, "--------------------------------------------\n")
		}
	}
}

func (self *SimpleLogger) Debugf(spec string, other ...interface{}) {
	self.Outf(syslog.LOG_DEBUG, spec, other...)
}

func (self *SimpleLogger) Infof(spec string, other ...interface{}) {
	self.Outf(syslog.LOG_INFO, spec, other...)
}

func (self *SimpleLogger) Warnf(spec string, other ...interface{}) {
	self.Outf(syslog.LOG_WARNING, spec, other...)
}

func (self *SimpleLogger) Noticef(spec string, other ...interface{}) {
	self.Outf(syslog.LOG_NOTICE, spec, other...)
}

func (self *SimpleLogger) Errf(spec string, other ...interface{}) {
	self.Outf(syslog.LOG_ERR, spec, other...)
}

func (self *SimpleLogger) Fatalf(spec string, other ...interface{}) {
	self.Outf(syslog.LOG_ERR, spec, other...)
	self.WaitToDie()
	os.Exit(1)
}
func (self *SimpleLogger) Panicf(spec string, other ...interface{}) {
	msg := fmt.Sprintf(spec, other...)
	self.Outf(syslog.LOG_ERR, spec, other...)
	self.WaitToDie()
	panic(msg)
}

func (self *SimpleLogger) Error(where string, err error) {
	self.ErrorLevelf(syslog.LOG_ERR, where, err)
}

func (self *SimpleLogger) ErrorPanic(where string, err error) {
	self.ErrorLevelf(syslog.LOG_ERR, where, err)
	panic(err)
}

func (self *SimpleLogger) ErrorFatal(where string, err error) {
	self.ErrorLevelf(syslog.LOG_ERR, where, err)
	self.WaitToDie()
	os.Exit(1)
}

//only needed for async loggers
func (self *SimpleLogger) WaitToDie() {
}
