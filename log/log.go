// Log:
//
// Simple wrapper for standard log package with level control.
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
}

func NewLogger() *Logger {
	return &Logger{
		level:     warn,
		useSyslog: false,
	}
}

var defaultLogger *Logger

func init() {
	defaultLogger = NewLogger()
}

func SetLevel(lvl string) error {
	return defaultLogger.SetLevel(lvl)
}

func Fatal(v ...interface{}) {
	defaultLogger.Fatal(v...)
}

func Warn(v ...interface{}) {
	defaultLogger.Warn(v...)
}

func Info(v ...interface{}) {
	defaultLogger.Info(v...)
}

func Debug(v ...interface{}) {
	defaultLogger.Debug(v...)
}

func (l *Logger) Fatal(v ...interface{}) {
	l.log(fatal, v...)
}

func (l *Logger) Warn(v ...interface{}) {
	l.log(warn, v...)
}

func (l *Logger) Info(v ...interface{}) {
	l.log(info, v...)
}

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
			_, file, line, _ := runtime.Caller(1)
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

// Set the maximum verbose level.
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
