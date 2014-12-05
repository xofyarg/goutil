package log

func (l *Logger) UseSyslog() {
	panic("syslog is not supported under windows")
}

func (l *Logger) writeSyslog(lvl level, v ...interface{}) {
}
