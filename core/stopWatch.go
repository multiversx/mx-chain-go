package core

import (
	"fmt"
	"sync"
	"time"
)

// MeasurementsLoggerFormat contains the formatting string to output elapsed time in seconds in a consistent way
const MeasurementsLoggerFormat = "%.4fs"

// StopWatch is used to measure duration
type StopWatch struct {
	mut         sync.RWMutex
	identifiers []string
	started     map[string]time.Time
	elapsed     map[string]time.Duration
}

// NewStopWatch returns a new stopWatch instance used to measure duration between finished and started events
func NewStopWatch() *StopWatch {
	return &StopWatch{
		identifiers: make([]string, 0),
		started:     make(map[string]time.Time),
		elapsed:     make(map[string]time.Duration),
	}
}

// Start marks a start event for a provided identifier
func (sw *StopWatch) Start(identifier string) {
	sw.mut.Lock()
	sw.addIdentifier(identifier)

	sw.started[identifier] = time.Now()
	sw.mut.Unlock()
}

func (sw *StopWatch) addIdentifier(identifier string) {
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
func (sw *StopWatch) Stop(identifier string) {
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
func (sw *StopWatch) GetMeasurements() []interface{} {
	data, newIdentifiers := sw.getContainingDuration()

	output := make([]interface{}, 0)
	for _, identifier := range newIdentifiers {
		duration := data[identifier]
		output = append(output, identifier)
		output = append(output, fmt.Sprintf(MeasurementsLoggerFormat, duration.Seconds()))
	}

	return output
}

// GetMeasurementsMap returns the measurements as a map of (identifier, duration in seconds)
func (sw *StopWatch) GetMeasurementsMap() map[string]float64 {
	sw.mut.RLock()
	defer sw.mut.RUnlock()

	output := make(map[string]float64)

	for identifier, duration := range sw.elapsed {
		output[identifier] = duration.Seconds()
	}

	return output
}

// GetMeasurement returns the measurement by identifier
func (sw *StopWatch) GetMeasurement(identifier string) time.Duration {
	sw.mut.RLock()
	defer sw.mut.RUnlock()

	duration, ok := sw.elapsed[identifier]
	if ok {
		return duration
	}

	return 0
}

// getContainingDuration returns the containing map of (identifier, duration) pairs and the identifiers
func (sw *StopWatch) getContainingDuration() (map[string]time.Duration, []string) {
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
func (sw *StopWatch) Add(src *StopWatch) {
	data, identifiers := src.getContainingDuration()

	sw.mut.Lock()
	for _, identifier := range identifiers {
		sw.addIdentifier(identifier)
		sw.elapsed[identifier] += data[identifier]
	}
	sw.mut.Unlock()
}
