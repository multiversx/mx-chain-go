package mock

import "github.com/ElrondNetwork/elrond-go/logger"

// LoggerStub -
type LoggerStub struct {
	Log            func(level string, message string, args ...interface{})
	SetLevelCalled func(logLevel logger.LogLevel)
}

// Trace -
func (l *LoggerStub) Trace(message string, args ...interface{}) {
	if l.Log != nil {
		l.Log("TRACE", message, args...)
	}
}

// Debug -
func (l *LoggerStub) Debug(message string, args ...interface{}) {
	if l.Log != nil {
		l.Log("DEBUG", message, args...)
	}
}

// Info -
func (l *LoggerStub) Info(message string, args ...interface{}) {
	if l.Log != nil {
		l.Log("INFO", message, args...)
	}
}

// Warn -
func (l *LoggerStub) Warn(message string, args ...interface{}) {
	if l.Log != nil {
		l.Log("WARN", message, args...)
	}
}

// Error -
func (l *LoggerStub) Error(message string, args ...interface{}) {
	if l.Log != nil {
		l.Log("ERROR", message, args...)
	}
}

// LogIfError -
func (l *LoggerStub) LogIfError(err error, args ...interface{}) {
	if l.Log != nil && err != nil {
		l.Log("ERROR", err.Error(), args...)
	}
}

// SetLevel -
func (l *LoggerStub) SetLevel(logLevel logger.LogLevel) {
	if l.SetLevelCalled != nil {
		l.SetLevelCalled(logLevel)
	}
}

// IsInterfaceNil -
func (l *LoggerStub) IsInterfaceNil() bool {
	return l == nil
}
