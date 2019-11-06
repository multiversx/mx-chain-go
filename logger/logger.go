package logger

import (
	"sync"
	"time"
)

// LogLine is the structure used to hold a log line
type LogLine struct {
	Message   string
	LogLevel  LogLevel
	Args      []interface{}
	Timestamp time.Time
}

func newLogLine(message string, logLevel LogLevel, args ...interface{}) *LogLine {
	return &LogLine{
		Message:   message,
		LogLevel:  logLevel,
		Args:      args,
		Timestamp: time.Now(),
	}
}

// logger is the primary structure used to interact with the productive code
type logger struct {
	name      string
	mutLevel  sync.RWMutex
	logLevel  LogLevel
	logOutput LogOutputHandler
}

// newLogger create a new logger instance
func newLogger(name string, logLevel LogLevel, logOutput LogOutputHandler) *logger {
	log := &logger{
		name:      name,
		logLevel:  logLevel,
		logOutput: logOutput,
	}

	return log
}

func (l *logger) shouldOutput(compareLogLevel LogLevel) bool {
	l.mutLevel.RLock()
	shouldOutput := l.logLevel > compareLogLevel
	l.mutLevel.RUnlock()

	return shouldOutput
}

func (l *logger) outputMessageFromLogLevel(level LogLevel, message string, args ...interface{}) {
	if l.shouldOutput(level) {
		return
	}

	logLine := newLogLine(message, level, args...)
	l.logOutput.Output(logLine)
}

// Trace outputs a tracing log message with optional provided arguments
func (l *logger) Trace(message string, args ...interface{}) {
	l.outputMessageFromLogLevel(LogTrace, message, args...)
}

// Debug outputs a debugging log message with optional provided arguments
func (l *logger) Debug(message string, args ...interface{}) {
	l.outputMessageFromLogLevel(LogDebug, message, args...)
}

// Info outputs an information log message with optional provided arguments
func (l *logger) Info(message string, args ...interface{}) {
	l.outputMessageFromLogLevel(LogInfo, message, args...)
}

// Warn outputs a warning log message with optional provided arguments
func (l *logger) Warn(message string, args ...interface{}) {
	l.outputMessageFromLogLevel(LogWarning, message, args...)
}

// Error outputs an error log message with optional provided arguments
func (l *logger) Error(message string, args ...interface{}) {
	l.outputMessageFromLogLevel(LogError, message, args...)
}

// LogIfError outputs an error log message with optional provided arguments if the provided error parameter is not nil
func (l *logger) LogIfError(err error, args ...interface{}) {
	if err == nil {
		return
	}

	l.Error(err.Error(), args...)
}

// SetLevel sets the current level of the logger
func (l *logger) SetLevel(logLevel LogLevel) {
	l.mutLevel.Lock()
	l.logLevel = logLevel
	l.mutLevel.Unlock()
}
