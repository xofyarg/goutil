// Log:
//
// Simple wrapper for standard log package with level control.
package util

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
	"path"
	"runtime"
)

const (
	none uint8 = iota
	Fatal
	Warn
	Info
	Debug
)

// global verbose log level
var verboseLevel = Warn

// global switch whether to use syslog
var syslogWriter *syslog.Writer

// mapping between numberic log level and their corresponding one
var levelStr = map[uint8]string{
	Fatal: "FATL",
	Warn:  "WARN",
	Info:  "INFO",
	Debug: "DBUG",
}

func Log(l uint8, v ...interface{}) {
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
		case Fatal:
			syslogWriter.Crit(msg)
		case Warn:
			syslogWriter.Warning(msg)
		case Info:
			syslogWriter.Info(msg)
		case Debug:
			syslogWriter.Debug(msg)
		}
	} else {
		var preamble string
		if l == Debug {
			_, file, line, _ := runtime.Caller(1)
			preamble = fmt.Sprintf("[%s %s:%d] ", levelStr[l], path.Base(file), line)
		} else {
			preamble = fmt.Sprintf("[%s] ", levelStr[l])
		}

		n := len(v)
		if n == 1 {
			log.Printf(preamble+"%v", v[0])
		} else {
			log.Printf(preamble+v[0].(string), v[1:]...)
		}
	}
}

// Set the maximum verbose level.
//
// Available levels are: Fatal, Warn, Info, Debug.
func SetLevel(l uint8) {
	verboseLevel = l
}

func UseSyslog() {
	w, err := syslog.New(syslog.LOG_INFO|syslog.LOG_USER,
		fmt.Sprintf("%s", path.Base(os.Args[0])))
	if err != nil {
		Log(Fatal, "cannot open syslog for write")
	}
	syslogWriter = w
}
