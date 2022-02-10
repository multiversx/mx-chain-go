package mock

import logger "github.com/ElrondNetwork/elrond-go-logger"

// LoggerStub -
type LoggerStub struct {
	LogLineCalled  func(level string, message string, args ...interface{})
	SetLevelCalled func(logLevel logger.LogLevel)
}

// Trace -
func (l *LoggerStub) Trace(message string, args ...interface{}) {
	if l.LogLineCalled != nil {
		l.LogLineCalled("TRACE", message, args...)
	}
}

// Debug -
func (l *LoggerStub) Debug(message string, args ...interface{}) {
	if l.LogLineCalled != nil {
		l.LogLineCalled("DEBUG", message, args...)
	}
}

// Info -
func (l *LoggerStub) Info(message string, args ...interface{}) {
	if l.LogLineCalled != nil {
		l.LogLineCalled("INFO", message, args...)
	}
}

// Warn -
func (l *LoggerStub) Warn(message string, args ...interface{}) {
	if l.LogLineCalled != nil {
		l.LogLineCalled("WARN", message, args...)
	}
}

// Error -
func (l *LoggerStub) Error(message string, args ...interface{}) {
	if l.LogLineCalled != nil {
		l.LogLineCalled("ERROR", message, args...)
	}
}

// LogIfError -
func (l *LoggerStub) LogIfError(err error, args ...interface{}) {
	if l.LogLineCalled != nil && err != nil {
		l.LogLineCalled("ERROR", err.Error(), args...)
	}
}

// LogLine -
func (l *LoggerStub) LogLine(line *logger.LogLine) {
	if l.LogLineCalled != nil {
		l.LogLineCalled("Log", "line", line)
	}
}

// Log -
func (l *LoggerStub) Log(logLevel logger.LogLevel, message string, args ...interface{}) {
	if l.LogLineCalled != nil {
		l.LogLineCalled(string(logLevel), message, args...)
	}
}

// SetLevel -
func (l *LoggerStub) SetLevel(logLevel logger.LogLevel) {
	if l.SetLevelCalled != nil {
		l.SetLevelCalled(logLevel)
	}
}

// GetLevel -
func (l *LoggerStub) GetLevel() logger.LogLevel {
	return logger.LogNone
}

// IsInterfaceNil -
func (l *LoggerStub) IsInterfaceNil() bool {
	return l == nil
}
