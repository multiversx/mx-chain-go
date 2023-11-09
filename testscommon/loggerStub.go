package testscommon

import logger "github.com/multiversx/mx-chain-logger-go"

// LoggerStub -
type LoggerStub struct {
	TraceCalled      func(message string, args ...interface{})
	DebugCalled      func(message string, args ...interface{})
	InfoCalled       func(message string, args ...interface{})
	WarnCalled       func(message string, args ...interface{})
	ErrorCalled      func(message string, args ...interface{})
	LogIfErrorCalled func(err error, args ...interface{})
	LogCalled        func(logLevel logger.LogLevel, message string, args ...interface{})
	LogLineCalled    func(line *logger.LogLine)
	SetLevelCalled   func(logLevel logger.LogLevel)
	GetLevelCalled   func() logger.LogLevel
}

// Log -
func (stub *LoggerStub) Log(logLevel logger.LogLevel, message string, args ...interface{}) {
	if stub.LogCalled != nil {
		stub.LogCalled(logLevel, message, args...)
	}
}

// LogLine -
func (stub *LoggerStub) LogLine(line *logger.LogLine) {
	if stub.LogLineCalled != nil {
		stub.LogLineCalled(line)
	}
}

// Trace -
func (stub *LoggerStub) Trace(message string, args ...interface{}) {
	if stub.TraceCalled != nil {
		stub.TraceCalled(message, args...)
	}
}

// Debug -
func (stub *LoggerStub) Debug(message string, args ...interface{}) {
	if stub.DebugCalled != nil {
		stub.DebugCalled(message, args...)
	}
}

// Info -
func (stub *LoggerStub) Info(message string, args ...interface{}) {
	if stub.InfoCalled != nil {
		stub.InfoCalled(message, args...)
	}
}

// Warn -
func (stub *LoggerStub) Warn(message string, args ...interface{}) {
	if stub.WarnCalled != nil {
		stub.WarnCalled(message, args...)
	}
}

// Error -
func (stub *LoggerStub) Error(message string, args ...interface{}) {
	if stub.ErrorCalled != nil {
		stub.ErrorCalled(message, args...)
	}
}

// LogIfError -
func (stub *LoggerStub) LogIfError(err error, args ...interface{}) {
	if stub.LogIfErrorCalled != nil {
		stub.LogIfErrorCalled(err, args...)
	}
}

// SetLevel -
func (stub *LoggerStub) SetLevel(logLevel logger.LogLevel) {
	if stub.SetLevelCalled != nil {
		stub.SetLevelCalled(logLevel)
	}
}

// GetLevel -
func (stub *LoggerStub) GetLevel() logger.LogLevel {
	if stub.GetLevelCalled != nil {
		return stub.GetLevelCalled()
	}

	return logger.LogTrace
}

// IsInterfaceNil -
func (stub *LoggerStub) IsInterfaceNil() bool {
	return stub == nil
}
