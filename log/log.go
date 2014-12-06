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
	"path"
	"runtime"
	"strings"
)

type level uint8

const (
	none level = iota
	fatal
	warn
	info
	debug
)

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

type Logger struct {
	level     level
	useSyslog bool
	w         interface{} // syslog writer
	nest      int         // call nest level
}

// create a logger with different destination and/or log level
func NewLogger() *Logger {
	return &Logger{
		level:     warn,
		useSyslog: false,
		nest:      2,
	}
}

var defaultLogger *Logger

func init() {
	defaultLogger = NewLogger()
	defaultLogger.nest = 3
}

// set log level for default logger.
//
// lvl can be one of this: "debug", "info", "warn", "fatal"
func SetLevel(lvl string) error {
	return defaultLogger.SetLevel(lvl)
}

// log fatal message for default logger
func Fatal(v ...interface{}) {
	defaultLogger.Fatal(v...)
}

// log warning message for default logger
func Warn(v ...interface{}) {
	defaultLogger.Warn(v...)
}

// log info message for default logger
func Info(v ...interface{}) {
	defaultLogger.Info(v...)
}

// log debug message for default logger
func Debug(v ...interface{}) {
	defaultLogger.Debug(v...)
}

// log fatal message
func (l *Logger) Fatal(v ...interface{}) {
	l.log(fatal, v...)
}

// log warnning message
func (l *Logger) Warn(v ...interface{}) {
	l.log(warn, v...)
}

// log info message
func (l *Logger) Info(v ...interface{}) {
	l.log(info, v...)
}

// log debug message
func (l *Logger) Debug(v ...interface{}) {
	l.log(debug, v...)
}

func (l *Logger) log(lvl level, v ...interface{}) {
	if lvl > l.level {
		return
	}

	// l.useSyslog
	if l.useSyslog {
		l.writeSyslog(lvl, v...)
	} else {
		var preamble string
		if lvl == debug {
			_, file, line, _ := runtime.Caller(l.nest)
			preamble = fmt.Sprintf("[%s %s:%d] ", levelStr[lvl],
				path.Base(file), line)
		} else {
			preamble = fmt.Sprintf("[%s] ", levelStr[lvl])
		}

		n := len(v)
		if n == 1 {
			olog.Printf(preamble+"%v", v[0])
		} else {
			olog.Printf(preamble+v[0].(string), v[1:]...)
		}
	}
}

// set log level for logger l.
//
// lvl can be one of this: "debug", "info", "warn", "fatal"
func (l *Logger) SetLevel(lvl string) error {
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
