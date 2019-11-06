package logger

import "io"

// Logger defines the behavior of a data logger component
type Logger interface {
	Trace(message string, args ...interface{})
	Debug(message string, args ...interface{})
	Info(message string, args ...interface{})
	Warn(message string, args ...interface{})
	Error(message string, args ...interface{})
	LogIfError(err error, args ...interface{})
	SetLevel(logLevel LogLevel)
}

// Formatter describes what a log formatter should be able to do
type Formatter interface {
	Output(line *LogLine) []byte
}

// LogOutputHandler defines the properties of a subject-observer component
// able to output log lines
type LogOutputHandler interface {
	Output(line *LogLine)
	AddObserver(w io.Writer, format Formatter) error
	RemoveObserver(w io.Writer) error
}
