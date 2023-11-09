package mock

import logger "github.com/multiversx/mx-chain-logger-go"

// LoggerStub -
type LoggerStub struct {
	LogCalled      func(level logger.LogLevel, message string, args ...interface{})
	LogLineCalled  func(line *logger.LogLine)
	SetLevelCalled func(logLevel logger.LogLevel)
}

// Log -
func (l *LoggerStub) Log(logLevel logger.LogLevel, message string, args ...interface{}) {
	if l.LogCalled != nil {
		l.LogCalled(logLevel, message, args...)
	}
}

// LogLine -
func (l *LoggerStub) LogLine(line *logger.LogLine) {
	if l.LogLineCalled != nil {
		l.LogLineCalled(line)
	}
}

// Trace -
func (l *LoggerStub) Trace(message string, args ...interface{}) {
	if l.LogCalled != nil {
		l.LogCalled(logger.LogTrace, message, args...)
	}
}

// Debug -
func (l *LoggerStub) Debug(message string, args ...interface{}) {
	if l.LogCalled != nil {
		l.LogCalled(logger.LogDebug, message, args...)
	}
}

// Info -
func (l *LoggerStub) Info(message string, args ...interface{}) {
	if l.LogCalled != nil {
		l.LogCalled(logger.LogInfo, message, args...)
	}
}

// Warn -
func (l *LoggerStub) Warn(message string, args ...interface{}) {
	if l.LogCalled != nil {
		l.LogCalled(logger.LogWarning, message, args...)
	}
}

// Error -
func (l *LoggerStub) Error(message string, args ...interface{}) {
	if l.LogCalled != nil {
		l.LogCalled(logger.LogError, message, args...)
	}
}

// LogIfError -
func (l *LoggerStub) LogIfError(err error, args ...interface{}) {
	if l.LogCalled != nil && err != nil {
		l.LogCalled(logger.LogError, err.Error(), args...)
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
