// Simple wrapper for standard log and syslog package with level
// control.
//
// Features:
//   1. level control. use SetLevel() to change level.
//   2. combined the logf and log function. log for one argument, logf
//      for multiple arguments.
//   3. output source file and line number information in debug log.
//
// Note:
//   1. level supported: fatal, warn, info, debug(with source file
//      information).
//   2. syslog feature is not supported by windows.
//
package log

import (
	"errors"
	"fmt"
	olog "log"
	"os"
	"path"
	"runtime"
	"strings"
)

// general interface for basic logger
type Logger interface {
	SetLevel(l string) error
	Fatal(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Info(format string, args ...interface{})
	Debug(format string, args ...interface{})
}

// interface used to extend the basic logger
type LoggerExtend interface {
	Logger
	IncNest(n int)
	UseSyslog() error
}

type level uint8

const (
	none level = iota
	fatal
	warn
	info
	debug
)

const defaultNest = 2

var (
	ErrInvLogLevel = errors.New("invalid log level")
	ErrOpenSyslog  = errors.New("error open syslog for write")
)

// mapping between numberic log level and their corresponding one
var levelStr = map[level]string{
	fatal: "FATL",
	warn:  "WARN",
	info:  "INFO",
	debug: "DBUG",
}

type logger struct {
	level     level
	useSyslog bool
	w         interface{} // syslog writer
	nest      int         // call nest level
}

// create a logger with different destination and/or log level
func NewLogger() LoggerExtend {
	return &logger{
		level:     warn,
		useSyslog: false,
		nest:      defaultNest,
	}
}

var defaultLogger LoggerExtend

func init() {
	defaultLogger = NewLogger()
	defaultLogger.IncNest(1)
}

// increase nest level for default logger
func IncNest(n int) {
	defaultLogger.IncNest(n)
}

// increase nest level for file/line info display. useful when
// extending the logging module
func (l *logger) IncNest(n int) {
	l.nest += n
}

// set log level for default logger.
//
// lvl can be one of this: "debug", "info", "warn", "fatal"
func SetLevel(lvl string) error {
	return defaultLogger.SetLevel(lvl)
}

// log fatal message for default logger
func Fatal(format string, v ...interface{}) {
	defaultLogger.Fatal(format, v...)
}

// log warning message for default logger
func Warn(format string, v ...interface{}) {
	defaultLogger.Warn(format, v...)
}

// log info message for default logger
func Info(format string, v ...interface{}) {
	defaultLogger.Info(format, v...)
}

// log debug message for default logger
func Debug(format string, v ...interface{}) {
	defaultLogger.Debug(format, v...)
}

// log fatal message
func (l *logger) Fatal(format string, v ...interface{}) {
	l.log(fatal, format, v...)
	os.Exit(1)
}

// log warnning message
func (l *logger) Warn(format string, v ...interface{}) {
	l.log(warn, format, v...)
}

// log info message
func (l *logger) Info(format string, v ...interface{}) {
	l.log(info, format, v...)
}

// log debug message
func (l *logger) Debug(format string, v ...interface{}) {
	l.log(debug, format, v...)
}

func (l *logger) log(lvl level, format string, v ...interface{}) {
	if lvl > l.level {
		return
	}

	if l.useSyslog {
		l.writeSyslog(lvl, format, v...)
	} else {
		var preamble string
		if lvl == debug {
			_, file, line, ok := runtime.Caller(l.nest)
			if !ok {
				file = "???"
				line = 1
			}
			preamble = fmt.Sprintf("[%s %s:%d] ", levelStr[lvl],
				path.Base(file), line)
		} else {
			preamble = fmt.Sprintf("[%s] ", levelStr[lvl])
		}

		olog.Printf(preamble+format, v...)
	}
}

// set log level for logger l.
//
// lvl can be one of this: "debug", "info", "warn", "fatal"
func (l *logger) SetLevel(lvl string) error {
	switch strings.ToLower(lvl) {
	case "fatal":
		l.level = fatal
	case "warn":
		l.level = warn
	case "info":
		l.level = info
	case "debug":
		l.level = debug
	default:
		return ErrInvLogLevel
	}
	return nil
}
