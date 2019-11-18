package mock

import "github.com/ElrondNetwork/elrond-go/logger"

type LoggerStub struct {
	Log            func(level string, message string, args ...interface{})
	SetLevelCalled func(logLevel logger.LogLevel)
}

func (l *LoggerStub) Trace(message string, args ...interface{}) {
	if l.Log != nil {
		l.Log("TRACE", message, args...)
	}
}

func (l *LoggerStub) Debug(message string, args ...interface{}) {
	if l.Log != nil {
		l.Log("DEBUG", message, args...)
	}
}

func (l *LoggerStub) Info(message string, args ...interface{}) {
	if l.Log != nil {
		l.Log("INFO", message, args...)
	}
}

func (l *LoggerStub) Warn(message string, args ...interface{}) {
	if l.Log != nil {
		l.Log("WARN", message, args...)
	}
}

func (l *LoggerStub) Error(message string, args ...interface{}) {
	if l.Log != nil {
		l.Log("ERROR", message, args...)
	}
}

func (l *LoggerStub) LogIfError(err error, args ...interface{}) {
	if l.Log != nil && err != nil {
		l.Log("ERROR", err.Error(), args...)
	}
}

func (l *LoggerStub) SetLevel(logLevel logger.LogLevel) {
	if l.SetLevelCalled != nil {
		l.SetLevelCalled(logLevel)
	}
}

func (l *LoggerStub) IsInterfaceNil() bool {
	return l == nil
}
