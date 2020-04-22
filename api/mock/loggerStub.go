package mock

import logger "github.com/ElrondNetwork/elrond-go-logger"

// LoggerStub -
type LoggerStub struct {
	LogCalled      func(level string, message string, args ...interface{})
	SetLevelCalled func(logLevel logger.LogLevel)
}

// Trace -
func (l *LoggerStub) Trace(message string, args ...interface{}) {
	if l.LogCalled != nil {
		l.LogCalled("TRACE", message, args...)
	}
}

// Debug -
func (l *LoggerStub) Debug(message string, args ...interface{}) {
	if l.LogCalled != nil {
		l.LogCalled("DEBUG", message, args...)
	}
}

// Info -
func (l *LoggerStub) Info(message string, args ...interface{}) {
	if l.LogCalled != nil {
		l.LogCalled("INFO", message, args...)
	}
}

// Warn -
func (l *LoggerStub) Warn(message string, args ...interface{}) {
	if l.LogCalled != nil {
		l.LogCalled("WARN", message, args...)
	}
}

// Error -
func (l *LoggerStub) Error(message string, args ...interface{}) {
	if l.LogCalled != nil {
		l.LogCalled("ERROR", message, args...)
	}
}

// LogIfError -
func (l *LoggerStub) LogIfError(err error, args ...interface{}) {
	if l.LogCalled != nil && err != nil {
		l.LogCalled("ERROR", err.Error(), args...)
	}
}

// Log -
func (l *LoggerStub) Log(line *logger.LogLine) {
	if l.LogCalled != nil {
		l.LogCalled("Log", "line", line)
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
