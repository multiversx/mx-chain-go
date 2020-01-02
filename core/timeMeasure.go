package core

import (
	"fmt"
	"sync"
	"time"
)

type timeMeasure struct {
	mut         sync.Mutex
	identifiers []string
	started     map[string]time.Time
	elapsed     map[string]time.Duration
}

// NewTimeMeasure returns a new timeMeasure instance used to measure duration between finished and started events
func NewTimeMeasure() *timeMeasure {
	return &timeMeasure{
		identifiers: make([]string, 0),
		started:     make(map[string]time.Time),
		elapsed:     make(map[string]time.Duration),
	}
}

// Start marks a start event for a provided identifier
func (tm *timeMeasure) Start(identifier string) {
	tm.mut.Lock()
	_, hasStarted := tm.started[identifier]
	_, hasElapsed := tm.elapsed[identifier]
	if !hasStarted && !hasElapsed {
		tm.identifiers = append(tm.identifiers, identifier)
	}

	tm.started[identifier] = time.Now()
	tm.mut.Unlock()
}

func (tm *timeMeasure) addIdentifier(identifier string) {
	_, hasStarted := tm.started[identifier]
	if hasStarted {
		return
	}

	_, hasElapsed := tm.elapsed[identifier]
	if hasElapsed {
		return
	}

	tm.identifiers = append(tm.identifiers, identifier)
}

// Finish marks a finish event for a provided identifier
func (tm *timeMeasure) Finish(identifier string) {
	tm.mut.Lock()
	defer tm.mut.Unlock()

	timeStarted, ok := tm.started[identifier]
	if !ok {
		return
	}

	tm.elapsed[identifier] += time.Now().Sub(timeStarted)
	delete(tm.started, identifier)
}

// GetMeasurements returns a logger compatible slice of interface{} containing pairs of (identifier, duration)
func (tm *timeMeasure) GetMeasurements() []interface{} {
	data, newIdentifiers := tm.GetContainingDuration()

	tm.mut.Lock()

	output := make([]interface{}, 0)
	for _, identifier := range newIdentifiers {
		duration := data[identifier]
		output = append(output, identifier)
		output = append(output, fmt.Sprintf("%.2fs", duration.Seconds()))
	}
	tm.mut.Unlock()

	return output
}

// GetContainingDuration returns the containing map of (identifier, duration) pairs and the identifiers
func (tm *timeMeasure) GetContainingDuration() (map[string]time.Duration, []string) {
	tm.mut.Lock()
	defer tm.mut.Unlock()

	output := make(map[string]time.Duration)
	newIdentifiers := make([]string, 0)
	for _, identifier := range tm.identifiers {
		duration, ok := tm.elapsed[identifier]
		if !ok {
			continue
		}

		output[identifier] = duration
		newIdentifiers = append(newIdentifiers, identifier)
	}

	return output, newIdentifiers
}

// Add adds a time measure containing duration list to self
func (tm *timeMeasure) Add(src *timeMeasure) {
	tm.mut.Lock()
	data, _ := src.GetContainingDuration()
	for identifier, duration := range data {
		tm.addIdentifier(identifier)
		tm.elapsed[identifier] += duration
	}
	tm.mut.Unlock()
}
