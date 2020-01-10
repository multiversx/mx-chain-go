package core

import (
	"fmt"
	"sync"
	"time"
)

// MeasurementsLoggerFormat contains the formatting string to output elapsed time in seconds in a consistent way
const MeasurementsLoggerFormat = "%.4fs"

type stopWatch struct {
	mut         sync.Mutex
	identifiers []string
	started     map[string]time.Time
	elapsed     map[string]time.Duration
}

// NewStopWatch returns a new stopWatch instance used to measure duration between finished and started events
func NewStopWatch() *stopWatch {
	return &stopWatch{
		identifiers: make([]string, 0),
		started:     make(map[string]time.Time),
		elapsed:     make(map[string]time.Duration),
	}
}

// Start marks a start event for a provided identifier
func (sw *stopWatch) Start(identifier string) {
	sw.mut.Lock()
	_, hasStarted := sw.started[identifier]
	_, hasElapsed := sw.elapsed[identifier]
	if !hasStarted && !hasElapsed {
		sw.identifiers = append(sw.identifiers, identifier)
	}

	sw.started[identifier] = time.Now()
	sw.mut.Unlock()
}

func (sw *stopWatch) addIdentifier(identifier string) {
	_, hasStarted := sw.started[identifier]
	if hasStarted {
		return
	}

	_, hasElapsed := sw.elapsed[identifier]
	if hasElapsed {
		return
	}

	sw.identifiers = append(sw.identifiers, identifier)
}

// Stop marks a finish event for a provided identifier
func (sw *stopWatch) Stop(identifier string) {
	sw.mut.Lock()
	defer sw.mut.Unlock()

	timeStarted, ok := sw.started[identifier]
	if !ok {
		return
	}

	sw.elapsed[identifier] += time.Since(timeStarted)
	delete(sw.started, identifier)
}

// GetMeasurements returns a logger compatible slice of interface{} containing pairs of (identifier, duration)
func (sw *stopWatch) GetMeasurements() []interface{} {
	data, newIdentifiers := sw.getContainingDuration()

	output := make([]interface{}, 0)
	for _, identifier := range newIdentifiers {
		duration := data[identifier]
		output = append(output, identifier)
		output = append(output, fmt.Sprintf(MeasurementsLoggerFormat, duration.Seconds()))
	}

	return output
}

// getContainingDuration returns the containing map of (identifier, duration) pairs and the identifiers
func (sw *stopWatch) getContainingDuration() (map[string]time.Duration, []string) {
	sw.mut.Lock()

	output := make(map[string]time.Duration)
	newIdentifiers := make([]string, 0)
	for _, identifier := range sw.identifiers {
		duration, ok := sw.elapsed[identifier]
		if !ok {
			continue
		}

		output[identifier] = duration
		newIdentifiers = append(newIdentifiers, identifier)
	}
	sw.mut.Unlock()

	return output, newIdentifiers
}

// Add adds a time measure containing duration list to self
func (sw *stopWatch) Add(src *stopWatch) {
	sw.mut.Lock()
	data, _ := src.getContainingDuration()
	for identifier, duration := range data {
		sw.addIdentifier(identifier)
		sw.elapsed[identifier] += duration
	}
	sw.mut.Unlock()
}
