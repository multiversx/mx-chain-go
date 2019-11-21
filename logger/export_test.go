package logger

import "io"

var DefaultLogLevel = &defaultLogLevel

func (los *logOutputSubject) Observers() ([]io.Writer, []Formatter) {
	los.mutObservers.RLock()
	defer los.mutObservers.RUnlock()

	return los.writers, los.formatters
}

func NewLogger(name string, logLevel LogLevel, logOutput LogOutputHandler) *logger {
	return newLogger(name, logLevel, logOutput)
}

func (l *logger) LogLevel() LogLevel {
	return l.logLevel
}
