package logger

import (
	"fmt"
	"io"
	"sync"

	"github.com/ElrondNetwork/elrond-go/core/check"
)

// logOutputSubject follows the observer-subject pattern by which it holds n Writer and n Formatters.
// Each time a call to the Output method is done, it iterates through the containing formatters and writers
// in order to output the data
type logOutputSubject struct {
	mutObservers sync.RWMutex
	writers      []io.Writer
	formatters   []Formatter
}

// NewLogOutputSubject returns an initialized, empty logOutputSubject with no observers
func NewLogOutputSubject() *logOutputSubject {
	return &logOutputSubject{
		writers:    make([]io.Writer, 0),
		formatters: make([]Formatter, 0),
	}
}

// Output triggers calls to all containing formatters and writers in order to output provided log line
func (los *logOutputSubject) Output(line *LogLine) {
	los.mutObservers.RLock()

	convertedLine := los.convertLogLine(line)
	for i := 0; i < len(los.writers); i++ {
		format := los.formatters[i]
		buff := format.Output(convertedLine)
		_, _ = los.writers[i].Write(buff)
	}

	los.mutObservers.RUnlock()
}

func (los *logOutputSubject) convertLogLine(logLine *LogLine) LogLineHandler {
	if logLine == nil {
		return nil
	}

	line := &LogLineWrapper{}
	line.Message = logLine.Message
	line.LogLevel = int32(logLine.LogLevel)
	line.Args = make([]string, len(logLine.Args))
	line.Timestamp = logLine.Timestamp.Unix()

	mutDisplayByteSlice.RLock()
	displayHandler := displayByteSlice
	mutDisplayByteSlice.RUnlock()

	for i, obj := range logLine.Args {
		switch obj.(type) {
		case []byte:
			line.Args[i] = displayHandler(obj.([]byte))
		default:
			line.Args[i] = fmt.Sprintf("%v", obj)
		}
	}

	return line
}

// AddObserver adds a writer + formatter (called here observer) to the containing observer-like lists
func (los *logOutputSubject) AddObserver(w io.Writer, format Formatter) error {
	if w == nil {
		return ErrNilWriter
	}
	if check.IfNil(format) {
		return ErrNilFormatter
	}

	los.mutObservers.Lock()
	los.writers = append(los.writers, w)
	los.formatters = append(los.formatters, format)
	los.mutObservers.Unlock()

	return nil
}

// RemoveObserver will remove the observer based on the writer provided. The comparision is done on pointers.
// If the provided writer is not contained, the function will return an error.
func (los *logOutputSubject) RemoveObserver(w io.Writer) error {
	if w == nil {
		return ErrNilWriter
	}

	los.mutObservers.Lock()
	defer los.mutObservers.Unlock()

	for i := 0; i < len(los.writers); i++ {
		if los.writers[i] == w {
			los.writers = append(los.writers[0:i], los.writers[i+1:]...)
			los.formatters = append(los.formatters[0:i], los.formatters[i+1:]...)
			return nil
		}
	}

	return ErrWriterNotFound
}

// ClearObservers clears the observers lists
func (los *logOutputSubject) ClearObservers() {
	los.mutObservers.Lock()

	los.writers = make([]io.Writer, 0)
	los.formatters = make([]Formatter, 0)

	los.mutObservers.Unlock()
}

// IsInterfaceNil returns true if there is no value under the interface
func (los *logOutputSubject) IsInterfaceNil() bool {
	return los == nil
}
