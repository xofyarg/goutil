// Log:
//
// Simple wrapper for standard log package with level control.
package util

import (
	"errors"
	"fmt"
	olog "log"
	"log/syslog"
	"os"
	"path"
	"runtime"
	"strings"
)

const (
	none uint8 = iota
	fatal
	warn
	info
	debug
)

// global verbose log level
var verboseLevel = warn

// global switch whether to use syslog
var syslogWriter *syslog.Writer

// mapping between numberic log level and their corresponding one
var levelStr = map[uint8]string{
	fatal: "FATL",
	warn:  "WARN",
	info:  "INFO",
	debug: "DBUG",
}

func log(l uint8, v ...interface{}) {
	if l > verboseLevel {
		return
	}

	if syslogWriter != nil {
		var msg string
		n := len(v)
		if n == 1 {
			msg = fmt.Sprintf("%v", v[0])
		} else {
			msg = fmt.Sprintf(v[0].(string), v[1:]...)
		}

		switch l {
		case fatal:
			syslogWriter.Crit(msg)
		case warn:
			syslogWriter.Warning(msg)
		case info:
			syslogWriter.Info(msg)
		case debug:
			syslogWriter.Debug(msg)
		}
	} else {
		var preamble string
		if l == debug {
			_, file, line, _ := runtime.Caller(1)
			preamble = fmt.Sprintf("[%s %s:%d] ", levelStr[l], path.Base(file), line)
		} else {
			preamble = fmt.Sprintf("[%s] ", levelStr[l])
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
//
// Available levels are: Fatal, Warn, Info, Debug.
// func SetLevel(l uint8) {
// 	verboseLevel = l
// }
func SetLevel(l string) error {
	switch strings.ToLower(l) {
	case "fatal":
		verboseLevel = fatal
	case "warn":
		verboseLevel = warn
	case "info":
		verboseLevel = info
	case "debug":
		verboseLevel = debug
	default:
		return errors.New("invalid log level")
	}
	return nil
}

func UseSyslog() {
	w, err := syslog.New(syslog.LOG_INFO|syslog.LOG_USER,
		fmt.Sprintf("%s", path.Base(os.Args[0])))
	if err != nil {
		log(fatal, "cannot open syslog for write")
	}
	syslogWriter = w
}

func Fatal(v ...interface{}) {
	log(fatal, v...)
}

func Warn(v ...interface{}) {
	log(warn, v...)
}

func Info(v ...interface{}) {
	log(info, v...)
}

func Debug(v ...interface{}) {
	log(debug, v...)
}
