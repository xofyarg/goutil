// +build !windows

package log

import (
	"fmt"
	"log/syslog"
	"os"
	"path"
)

// write log to syslog with default settings:
//   syslog.LOG_INFO|syslog.LOG_USER
func (l *logger) UseSyslog() error {
	l.useSyslog = true
	if l.w == nil {
		w, err := syslog.New(syslog.LOG_INFO|syslog.LOG_USER,
			fmt.Sprintf("%s", path.Base(os.Args[0])))
		if err != nil {
			return ErrOpenSyslog
		}
		l.w = w
	}
	return nil
}

func (l *logger) writeSyslog(lvl level, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)

	switch lvl {
	case fatal:
		l.w.(*syslog.Writer).Crit(msg)
	case warn:
		l.w.(*syslog.Writer).Warning(msg)
	case info:
		l.w.(*syslog.Writer).Info(msg)
	case debug:
		l.w.(*syslog.Writer).Debug(msg)
	}

}

func UseSyslog() error {
	return defaultLogger.UseSyslog()
}
