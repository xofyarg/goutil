// +build !windows

package log

import (
	"fmt"
	"log/syslog"
	"os"
	"path"
)

func (l *Logger) UseSyslog() error {
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

func (l *Logger) writeSyslog(lvl level, v ...interface{}) {
	var msg string
	n := len(v)
	if n == 1 {
		msg = fmt.Sprintf("%v", v[0])
	} else {
		msg = fmt.Sprintf(v[0].(string), v[1:]...)
	}

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
