package logger

import "io"

// Formatter describes what an log formatter should be able to do
type Formatter interface {
	Output(line *LogLine) []byte
}

// LogOutputHandler defines the properties of a subject-observer component
// able to output log lines
type LogOutputHandler interface {
	Output(line *LogLine)
	AddObserver(w io.Writer, format Formatter) error
}
