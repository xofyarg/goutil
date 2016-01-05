package log

func (l *logger) UseSyslog() error {
	panic("syslog is not supported under windows")
}

func (l *logger) writeSyslog(lvl level, format string, v ...interface{}) {
}

func UseSyslog() error {
	panic("syslog is not supported under windows")
}
