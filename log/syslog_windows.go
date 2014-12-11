package log

func (l *Logger) UseSyslog() error {
	panic("syslog is not supported under windows")
}

func (l *Logger) writeSyslog(lvl level, v ...interface{}) {
}

func UseSyslog() error {
	panic("syslog is not supported under windows")
}
