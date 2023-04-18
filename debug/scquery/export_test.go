package scquery

import "time"

// Metrics -
type Metrics struct {
	RunningQueries            map[int]int64
	MaxNumOfConcurrentQueries int
	NumFinishedCalls          int
	IndexOfHeavierSCCall      int
	DurationOfHeaviestSCCall  time.Duration
}

// GetMetrics -
func (debugger *scQueryDebugger) GetMetrics() Metrics {
	result := Metrics{
		RunningQueries: make(map[int]int64),
	}

	debugger.mutData.Lock()
	defer debugger.mutData.Unlock()

	for key, value := range debugger.runningQueries {
		result.RunningQueries[key] = value
	}
	result.IndexOfHeavierSCCall = debugger.indexOfHeavierSCCall
	result.MaxNumOfConcurrentQueries = debugger.maxNumOfConcurrentQueries
	result.NumFinishedCalls = debugger.numFinishedCalls
	result.DurationOfHeaviestSCCall = debugger.durationOfHeaviestSCCall

	return result
}

// Reset -
func (debugger *scQueryDebugger) Reset() {
	debugger.mutData.Lock()
	defer debugger.mutData.Unlock()

	debugger.resetMetrics()
}

// FullReset -
func (debugger *scQueryDebugger) FullReset() {
	debugger.mutData.Lock()
	defer debugger.mutData.Unlock()

	debugger.resetMetrics()
	debugger.runningQueries = make(map[int]int64)
}

// NewTestSCQueryDebugger -
func NewTestSCQueryDebugger(args ArgsSCQueryDebugger, timestampHandler func() int64) (*scQueryDebugger, error) {
	return newSCQueryDebugger(args, timestampHandler)
}
