package logger

import (
	"io"
	"sync"
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

	for i := 0; i < len(los.writers); i++ {
		format := los.formatters[i]
		buff := format.Output(line)
		_, _ = los.writers[i].Write(buff)
	}

	los.mutObservers.RUnlock()
}

// AddObserver adds a writer + formatter (called here observer) to the containing observer-like lists
func (los *logOutputSubject) AddObserver(w io.Writer, format Formatter) error {
	if w == nil {
		return ErrNilWriter
	}
	if format == nil {
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
